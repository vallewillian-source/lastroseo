import { useState, useEffect, useCallback } from 'react';
import type { Keyword } from '../api/client';
import { getKeywords } from '../api/client';

export function useKeywords(projectId: string | undefined) {
  const [keywords, setKeywords] = useState<Keyword[]>([]);
  const [loading, setLoading] = useState(false);

  const fetch = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const kw = await getKeywords(projectId);
      setKeywords(kw ?? []);
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { fetch(); }, [fetch]);

  return { keywords, loading, refetch: fetch };
}
