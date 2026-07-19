import { createHash, randomBytes, randomUUID } from "node:crypto";
import { access, chmod, mkdir, mkdtemp, readFile, rename, rm, writeFile } from "node:fs/promises";
import { isIP } from "node:net";
import { join } from "node:path";
import type { DeviceRecord } from "../domain/records.js";
import { runCommand } from "../infrastructure/processes.js";
import type { ResourceStore } from "../infrastructure/resource-store.js";
import type { SecretStore } from "../infrastructure/secret-store.js";

const PAIRING_LIFETIME_MS = 10 * 60 * 1000;
const DEVICE_LIFETIME_DAYS = 365;

export interface PairingOffer {
  readonly payload: string;
  readonly expireTime: Date;
}

export interface PairedDeviceCredentials {
  readonly device: DeviceRecord;
  readonly pkcs12: Uint8Array;
  readonly pkcs12Password: string;
  readonly deviceToken: string;
  readonly caCertificatePem: Uint8Array;
}

interface PkiPaths {
  readonly directory: string;
  readonly caKey: string;
  readonly caCertificate: string;
  readonly serverKey: string;
  readonly serverCertificate: string;
}

export class PairingAuthority {
  private operation: Promise<unknown> = Promise.resolve();

  private constructor(
    private readonly store: ResourceStore,
    private readonly secrets: SecretStore,
    private readonly opensslPath: string,
    private readonly serviceUrl: string,
    private readonly pairingUrl: string,
    readonly pki: PkiPaths,
  ) {}

  static async open(options: {
    store: ResourceStore;
    secrets: SecretStore;
    opensslPath: string;
    dataDirectory: string;
    serviceUrl: string;
    pairingUrl: string;
  }): Promise<PairingAuthority> {
    const directory = join(options.dataDirectory, "pki");
    const authority = new PairingAuthority(
      options.store,
      options.secrets,
      options.opensslPath,
      options.serviceUrl,
      options.pairingUrl,
      {
        directory,
        caKey: join(directory, "ca.key.pem"),
        caCertificate: join(directory, "ca.cert.pem"),
        serverKey: join(directory, "server.key.pem"),
        serverCertificate: join(directory, "server.cert.pem"),
      },
    );
    await authority.ensurePki();
    return authority;
  }

  async createOffer(): Promise<PairingOffer> {
    const secret = randomBytes(32);
    const expireTime = new Date(Date.now() + PAIRING_LIFETIME_MS);
    const caCertificate = await readFile(this.pki.caCertificate);
    this.store.createPairingSecret(this.secrets.hashPairingSecret(secret), expireTime);
    return {
      expireTime,
      payload: JSON.stringify({
        version: 1,
        serviceUrl: this.serviceUrl,
        pairingUrl: this.pairingUrl,
        pairingSecret: secret.toString("base64url"),
        caCertificatePem: caCertificate.toString("base64"),
        caSha256: createHash("sha256").update(caCertificate).digest("hex"),
        expireTime: expireTime.toISOString(),
      }),
    };
  }

  renewServerCertificate(): Promise<void> {
    return this.exclusive(() => this.issueServerCertificate());
  }

  pair(secret: Uint8Array, displayName: string): Promise<PairedDeviceCredentials> {
    return this.exclusive(async () => {
      if (secret.byteLength !== 32) {
        throw new Error("Invalid pairing secret");
      }
      if (!this.store.consumePairingSecret(this.secrets.hashPairingSecret(secret))) {
        throw new Error("Pairing secret is invalid, expired, or already used");
      }
      return this.issueDevice(displayName);
    });
  }

  private async ensurePki(): Promise<void> {
    await mkdir(this.pki.directory, { recursive: true, mode: 0o700 });
    const paths = [
      this.pki.caKey,
      this.pki.caCertificate,
      this.pki.serverKey,
      this.pki.serverCertificate,
    ];
    const exists = await Promise.all(paths.map(fileExists));
    if (exists.every(Boolean)) {
      return;
    }
    if (exists.some(Boolean)) {
      throw new Error("PKI directory is incomplete; restore it or remove it before reinitializing");
    }
    await this.openssl([
      "req",
      "-x509",
      "-newkey",
      "rsa:3072",
      "-sha256",
      "-days",
      "3650",
      "-nodes",
      "-subj",
      "/CN=Super Manager Private CA",
      "-addext",
      "basicConstraints=critical,CA:TRUE",
      "-addext",
      "keyUsage=critical,keyCertSign,cRLSign",
      "-keyout",
      this.pki.caKey,
      "-out",
      this.pki.caCertificate,
    ]);
    await chmod(this.pki.caKey, 0o600);
    await this.issueServerCertificate();
  }

  private async issueServerCertificate(): Promise<void> {
    const temporary = await mkdtemp(join(this.pki.directory, ".server-"));
    try {
      const key = join(temporary, "server.key.pem");
      const certificate = join(temporary, "server.cert.pem");
      const csr = join(temporary, "server.csr.pem");
      const extensions = join(temporary, "server.ext.cnf");
      const hostname = new URL(this.serviceUrl).hostname.replace(/^\[|\]$/g, "");
      const subjectAltName = isIP(hostname) ? `IP:${hostname}` : `DNS:${hostname}`;
      await writeFile(
        extensions,
        [
          "basicConstraints=critical,CA:FALSE",
          "keyUsage=critical,digitalSignature,keyEncipherment",
          "extendedKeyUsage=serverAuth",
          `subjectAltName=${subjectAltName}`,
          "",
        ].join("\n"),
        { mode: 0o600 },
      );
      await this.openssl([
        "req",
        "-newkey",
        "rsa:2048",
        "-sha256",
        "-nodes",
        "-subj",
        `/CN=${hostname}`,
        "-keyout",
        key,
        "-out",
        csr,
      ]);
      await this.openssl([
        "x509",
        "-req",
        "-sha256",
        "-days",
        "825",
        "-in",
        csr,
        "-CA",
        this.pki.caCertificate,
        "-CAkey",
        this.pki.caKey,
        "-set_serial",
        randomSerialArgument(),
        "-extfile",
        extensions,
        "-out",
        certificate,
      ]);
      await this.openssl(["verify", "-CAfile", this.pki.caCertificate, certificate]);
      await chmod(key, 0o600);
      await rename(key, this.pki.serverKey);
      await rename(certificate, this.pki.serverCertificate);
    } finally {
      await rm(temporary, { recursive: true, force: true });
    }
  }

  private async issueDevice(displayName: string): Promise<PairedDeviceCredentials> {
    const temporary = await mkdtemp(join(this.pki.directory, ".device-"));
    try {
      const uid = randomUUID();
      const serial = randomSerialHex();
      const key = join(temporary, "device.key.pem");
      const csr = join(temporary, "device.csr.pem");
      const certificate = join(temporary, "device.cert.pem");
      const extensions = join(temporary, "device.ext.cnf");
      const passwordFile = join(temporary, "pkcs12-password");
      const pkcs12Path = join(temporary, "device.p12");
      const pkcs12Password = randomBytes(24).toString("base64url");
      const deviceToken = randomBytes(32).toString("base64url");
      const expireTime = new Date(Date.now() + DEVICE_LIFETIME_DAYS * 24 * 60 * 60 * 1000);
      await writeFile(
        extensions,
        [
          "basicConstraints=critical,CA:FALSE",
          "keyUsage=critical,digitalSignature",
          "extendedKeyUsage=clientAuth",
          `subjectAltName=URI:urn:super-manager:device:${uid}`,
          "",
        ].join("\n"),
        { mode: 0o600 },
      );
      await writeFile(passwordFile, pkcs12Password, { mode: 0o600 });
      await this.openssl([
        "req",
        "-newkey",
        "rsa:2048",
        "-sha256",
        "-nodes",
        "-subj",
        `/CN=${uid}`,
        "-keyout",
        key,
        "-out",
        csr,
      ]);
      await this.openssl([
        "x509",
        "-req",
        "-sha256",
        "-days",
        String(DEVICE_LIFETIME_DAYS),
        "-in",
        csr,
        "-CA",
        this.pki.caCertificate,
        "-CAkey",
        this.pki.caKey,
        "-set_serial",
        `0x${serial}`,
        "-extfile",
        extensions,
        "-out",
        certificate,
      ]);
      await this.openssl(["verify", "-CAfile", this.pki.caCertificate, certificate]);
      await this.openssl([
        "pkcs12",
        "-export",
        "-inkey",
        key,
        "-in",
        certificate,
        "-certfile",
        this.pki.caCertificate,
        "-name",
        `Super Manager ${displayName}`,
        "-out",
        pkcs12Path,
        "-passout",
        `file:${passwordFile}`,
      ]);
      const device = this.store.createDevice({
        uid,
        displayName,
        certificateSerial: serial,
        tokenHash: this.secrets.hashDeviceToken(deviceToken),
        expireTime,
      });
      return {
        device,
        pkcs12: await readFile(pkcs12Path),
        pkcs12Password,
        deviceToken,
        caCertificatePem: await readFile(this.pki.caCertificate),
      };
    } finally {
      await rm(temporary, { recursive: true, force: true });
    }
  }

  private async openssl(args: readonly string[]): Promise<void> {
    const result = await runCommand(this.opensslPath, args, { timeoutMs: 60_000 });
    if (result.exitCode !== 0) {
      throw new Error(result.stderr.trim().slice(-1024) || "OpenSSL command failed");
    }
  }

  private exclusive<T>(operation: () => Promise<T>): Promise<T> {
    const result = this.operation.then(operation, operation);
    this.operation = result.then(
      () => undefined,
      () => undefined,
    );
    return result;
  }
}

async function fileExists(path: string): Promise<boolean> {
  try {
    await access(path);
    return true;
  } catch {
    return false;
  }
}

function randomSerialHex(): string {
  return randomBytes(16).toString("hex").replace(/^0+/, "") || "1";
}

function randomSerialArgument(): string {
  return `0x${randomSerialHex()}`;
}
