import { useEffect, useState } from 'react';

const QUERY = '(prefers-reduced-motion: reduce)';

// usePrefersReducedMotion tracks the viewer's motion preference. CSS can pause a
// keyframe animation but not a GIF, so a component that animates through an
// <img> has to honour the preference itself.
export function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(() => window.matchMedia(QUERY).matches);

  useEffect(() => {
    const media = window.matchMedia(QUERY);
    const onChange = () => setReduced(media.matches);
    media.addEventListener('change', onChange);
    return () => media.removeEventListener('change', onChange);
  }, []);

  return reduced;
}
