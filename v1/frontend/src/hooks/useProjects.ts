import { useState, useEffect, useCallback } from 'react';
import type { Project } from '../api/client';
import { getProjects, createProject as apiCreate } from '../api/client';

export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    try {
      const p = await getProjects();
      setProjects(p ?? []);
      setError(null);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);

  const create = useCallback(async (body: {
    name: string;
    business_desc?: string;
    seed_keywords?: string[];
  }) => {
    const result = await apiCreate(body);
    await fetch();
    return result;
  }, [fetch]);

  return { projects, loading, error, refetch: fetch, create };
}
