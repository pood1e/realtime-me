import type { ReactElement } from 'react';
import type { SimpleIcon } from 'simple-icons';
import {
  siAlpinelinux,
  siAndroid,
  siApple,
  siArchlinux,
  siCentos,
  siDebian,
  siFedora,
  siKalilinux,
  siLinux,
  siLinuxmint,
  siPopos,
  siRedhat,
  siUbuntu,
  siWearos,
  siZorin,
} from 'simple-icons/icons';
import type { DeviceState } from '@/gen/realtime/me/v1/status_pb';

export function BrandIcon({ icon, className = 'size-4' }: { icon: SimpleIcon; className?: string }) {
  return (
    <svg aria-label={icon.title} className={`${className} shrink-0`} role="img" style={{ color: `#${icon.hex}` }} viewBox="0 0 24 24">
      <title>{icon.title}</title>
      <path d={icon.path} fill="currentColor" />
    </svg>
  );
}

export function deviceIcon(device: DeviceState | null | undefined, fallback: ReactElement): ReactElement {
  const icon = osIcon(device?.model || device?.displayName || '');
  return icon ? <BrandIcon icon={icon} /> : fallback;
}

function osIcon(value: string): SimpleIcon | null {
  const text = value.toLowerCase();
  if (text.includes('wear os')) return siWearos;
  if (text.includes('android')) return siAndroid;
  if (text.includes('macos') || text.includes('darwin')) return siApple;
  if (text.includes('kali')) return siKalilinux;
  if (text.includes('ubuntu')) return siUbuntu;
  if (text.includes('debian')) return siDebian;
  if (text.includes('fedora')) return siFedora;
  if (text.includes('arch')) return siArchlinux;
  if (text.includes('centos')) return siCentos;
  if (text.includes('red hat') || text.includes('rhel')) return siRedhat;
  if (text.includes('alpine')) return siAlpinelinux;
  if (text.includes('linux mint')) return siLinuxmint;
  if (text.includes('pop!_os') || text.includes('pop! os')) return siPopos;
  if (text.includes('zorin')) return siZorin;
  if (text.includes('linux')) return siLinux;
  return null;
}
