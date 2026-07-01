import { useState, useEffect } from 'react';

interface ServicesHealth {
  services: { gateway: string; postgres: string; redis: string; searxng: string };
}

async function getServicesHealth(): Promise<ServicesHealth | null> {
  try {
    const res = await fetch('/api/services');
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export function useHealth() {
  const [health, setHealth] = useState<ServicesHealth | null>(null);

  useEffect(() => {
    let cancelled = false;
    const poll = async () => {
      const h = await getServicesHealth();
      if (!cancelled) setHealth(h);
      if (!cancelled) setTimeout(poll, 5000);
    };
    poll();
    return () => { cancelled = true; };
  }, []);

  return health;
}
