import { Globe, Mail } from 'lucide-react';
import type { ReactElement } from 'react';
import { siDiscord, siGithub, siGmail, siTelegram } from 'simple-icons/icons';
import type { ProfileLink } from '@/gen/realtime/me/v1/profile_pb';
import { Button } from '@/components/ui/button';
import { BrandIcon } from '@/components/brand';

export function ContactLinks({ links }: { links?: ProfileLink[] }) {
  if (!links || links.length === 0) return null;
  return (
    <div className="flex items-center gap-0.5">
      {links.map((link) => {
        const label = link.label || link.platform || 'Link';
        return (
          <Button key={`${link.platform}:${link.uri}`} asChild variant="ghost" size="icon" aria-label={label} title={label}>
            <a href={link.uri} target="_blank" rel="noreferrer">{contactIcon(link)}</a>
          </Button>
        );
      })}
    </div>
  );
}

export function contactIcon(link: ProfileLink): ReactElement {
  const platform = link.platform.toLowerCase();
  if (platform === 'github') return <BrandIcon icon={siGithub} />;
  if (platform === 'telegram') return <BrandIcon icon={siTelegram} />;
  if (platform === 'discord') return <BrandIcon icon={siDiscord} />;
  if (platform === 'email' || link.uri.startsWith('mailto:')) {
    return link.uri.includes('gmail') ? <BrandIcon icon={siGmail} /> : <Mail />;
  }
  return <Globe />;
}
