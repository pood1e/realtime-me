# OpenVPN management plane

Local `192.168.0.0/24` clients use the hosts' existing LAN addresses directly.
OpenVPN adds a separate, split-tunnel `10.66.0.0/24` path for remote clients; it
does not route or claim any `192.168.0.x` address. A remote location may therefore
also use `192.168.0.0/24` without a route collision.

```text
LAN owner/agents  192.168.0.x -------- direct --------> existing host LAN IPs
remote owner      10.66.0.100-200 -- OpenVPN --------> 10.66.0.1  Manager :443/:8443
                                                        10.66.0.1  Console :9443
remote agents                         OpenVPN --------> 10.66.0.10 Status  :18080
remote owner                          Console --------> LAN Status/Library addresses
optional direct owner                 OpenVPN --------> 10.66.0.11 Library :18081
```

Local subnet membership or a VPN certificate proves network reachability. Private application calls also
require `X-Realtime-Internal-Key` and an owner OIDC access token. Workload ingest
and Prometheus discovery retain their separate bearer credentials.

## Server

Use OpenVPN Community 2.7.4 or a maintained 2.6 release. Create a dedicated CA;
do not reuse the Manager device CA or Caddy's local HTTPS CA. Issue a unique
certificate and tls-crypt-v2 key for every client. The server certificate common
name must be `realtime-me-vpn-server` and include the TLS Web Server EKU; client
certificates must include the TLS Web Client EKU.

Install the templates on the Manager host:

```sh
sudo install -d -m 0700 /etc/openvpn/server/realtime-me-pki \
  /etc/openvpn/server/realtime-me-ccd /var/lib/openvpn/realtime-me
sudo install -m 0600 deploy/vpn/server/realtime-me.conf \
  /etc/openvpn/server/realtime-me.conf
sudo install -m 0600 deploy/vpn/server/ccd/status \
  deploy/vpn/server/ccd/library /etc/openvpn/server/realtime-me-ccd/
sudo install -m 0644 deploy/vpn/sysctl/90-realtime-me-openvpn.conf \
  /etc/sysctl.d/90-realtime-me-openvpn.conf
sudo sysctl --system
```

Place `ca.crt`, `server.crt`, `server.key`, and a current `crl.pem` in the PKI
directory. Generate the server and per-client control-channel keys without
printing them:

```sh
sudo openvpn --genkey tls-crypt-v2-server \
  /etc/openvpn/server/realtime-me-pki/tls-crypt-v2-server.key
sudo openvpn --tls-crypt-v2 /etc/openvpn/server/realtime-me-pki/tls-crypt-v2-server.key \
  --genkey tls-crypt-v2-client /secure/output/CLIENT.tls-crypt-v2.key
```

The Status and Library service clients use certificate common names `status` and
`library` and receive the committed overlay addresses `.10` and `.11`. Their APIs
bind both their pre-existing LAN address and that overlay address; no new LAN IP
is allocated. Copy `ccd/owner.example` to each owner's unique certificate common
name; `ccd-exclusive` rejects certificates without a matching file. Never enable
`duplicate-cn`, password-only authentication, `client-to-client`, compression, or
`redirect-gateway`.

Install the nftables rules into the VPN server's persistent ruleset, changing `tun0`
only if the device name was intentionally changed. They permit remote owner access to
Console and Library, all enrolled VPN clients to the token-protected Status
ingest port, established replies, and no other overlay client-to-client traffic. They
do not forward the home LAN subnet. Merge
the chains with an existing firewall rather than creating a second ruleset that
flushes it. Manager and Console remain private: their Caddy listeners bind only
the existing LAN address and `10.66.0.1`.

Allow and forward only UDP 1194 on the router/firewall. Do not forward TCP 443,
8443, or 9443 and do not publish a Tunnel hostname for Manager or Console. The
OpenVPN endpoint may use a dedicated public `vpn.example.com` DDNS record. Then
start the server:

```sh
sudo systemctl enable --now openvpn-server@realtime-me.service
sudo systemctl status openvpn-server@realtime-me.service
```

## Clients and private DNS

Install `client/realtime-me.conf.example` on Status, Library, and remote owner/agent
devices as a root-owned client profile, replace
the remote hostname, and install only that client's CA, certificate, private key,
and tls-crypt-v2 key. Use split DNS for the one canonical Console origin:

```text
LAN DNS:      console.realtime.internal -> existing Manager LAN address
OpenVPN DNS:  console.realtime.internal -> 10.66.0.1
LAN DNS:      manager.realtime.internal -> existing Manager LAN address
OpenVPN DNS:  manager.realtime.internal -> 10.66.0.1
LAN DNS:      status.realtime.internal  -> existing Status LAN address
OpenVPN DNS:  status.realtime.internal  -> 10.66.0.10
```

The Console uses Caddy's local HTTPS issuer. Export its root certificate from the
Manager host and trust it only on owner devices; do not publish it as a public
web PKI root. Verify the split-tunnel and bindings after connection:

```sh
ip route get 10.66.0.10
ip route get 1.1.1.1
curl --fail https://console.realtime.internal:9443/healthz
```

Revoke a lost client certificate, regenerate `crl.pem`, and reload OpenVPN. A VPN
certificate or tls-crypt-v2 key is never shared between hosts.

## Internal application key

Generate this credential once on an offline/admin host, copy the same root-only
file to Manager, Status, and Library, then delete the transfer copy:

```sh
umask 077
openssl rand -hex 32 > internal-api-key
sudo install -m 0400 -o root -g root internal-api-key \
  /etc/realtime-me/internal-api-key
```

Compose mounts the file as a secret. Manager and Console receive it through
systemd `LoadCredential`; it does not belong in an environment file, image,
repository, log, browser, Tunnel configuration, or mobile client. Rotate it as a
single maintenance operation across all four processes.
