import { useState, useEffect } from 'react';
import type { Job } from '../api/client';
import { getJob } from '../api/client';

export function useJob(jobId: string | undefined) {
  const [job, setJob] = useState<Job | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!jobId) return;
    let cancelled = false;
    setLoading(true);
    const poll = () =>
      getJob(jobId).then((j) => {
        if (!cancelled) { setJob(j); setLoading(false); }
        if (!cancelled && (j.status === 'PENDING' || j.status === 'PROCESSING')) {
          setTimeout(poll, 2000);
        }
      }).catch(() => {
        if (!cancelled) setLoading(false);
      });
    poll();
    return () => { cancelled = true; };
  }, [jobId]);

  return { job, loading };
}
