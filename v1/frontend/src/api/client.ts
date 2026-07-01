const BASE = '/api/v1';

export interface Project {
  id: string;
  name: string;
  business_desc: string;
  target_audience: string;
  created_at: string;
  updated_at: string;
}

export interface Keyword {
  id: string;
  project_id: string;
  keyword: string;
  is_seed: boolean;
  cluster_name: string | null;
  intent: string | null;
  source: string | null;
  created_at: string;
}

export interface Job {
  id: string;
  project_id: string;
  type: string;
  status: string;
  payload: string;
  result: string | null;
  created_at: string;
  updated_at: string;
}

export interface SERPResult {
  id: string;
  keyword_id: string;
  keyword: string;
  position: number;
  url: string;
  title: string;
  snippet: string;
  crawled_at: string;
}

export interface Cluster {
  id: string;
  project_id: string;
  name: string;
  intent: string | null;
  keyword_count: number;
  created_at: string;
  updated_at: string;
}

export interface ContentGap {
  id: string;
  project_id: string;
  keyword_id: string;
  keyword: string;
  gaps: string;
  created_at: string;
}

export interface Competitor {
  id: string;
  project_id: string;
  name: string;
  url: string;
  created_at: string;
}

export interface InspectKeyword {
  keyword: string;
  score: number;
  count: number;
}

export interface InspectResult {
  competitor_id: string;
  name: string;
  url: string;
  keywords: InspectKeyword[];
  count: number;
}

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

export async function getProjects(): Promise<Project[]> {
  const data = await fetchJSON<{ projects: Project[] }>(`${BASE}/projects`);
  return data.projects ?? [];
}

export async function getProject(id: string): Promise<Project> {
  return fetchJSON<Project>(`${BASE}/projects/${id}`);
}

export async function createProject(body: {
  name: string;
  business_desc?: string;
  target_audience?: string;
  seed_keywords?: string[];
}): Promise<{ project_id: string; job_id: string; status: string }> {
  return fetchJSON(`${BASE}/projects`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function getKeywords(projectId: string): Promise<Keyword[]> {
  const data = await fetchJSON<{ keywords: Keyword[] }>(
    `${BASE}/projects/${projectId}/keywords`
  );
  return data.keywords ?? [];
}

export async function getSERPResults(projectId: string): Promise<SERPResult[]> {
  const data = await fetchJSON<{ results: SERPResult[] }>(
    `${BASE}/projects/${projectId}/serp`
  );
  return data.results ?? [];
}

export async function getJob(jobId: string): Promise<Job> {
  return fetchJSON<Job>(`${BASE}/jobs/${jobId}`);
}

export async function getClusters(projectId: string): Promise<Cluster[]> {
  const data = await fetchJSON<{ clusters: Cluster[] }>(
    `${BASE}/projects/${projectId}/clusters`
  );
  return data.clusters ?? [];
}

export async function getGaps(projectId: string): Promise<ContentGap[]> {
  const data = await fetchJSON<{ gaps: ContentGap[] }>(
    `${BASE}/projects/${projectId}/gaps`
  );
  return data.gaps ?? [];
}

export async function getHealth(): Promise<{
  status: string;
  checks: { postgres: string; redis: string };
} | null> {
  try {
    return await fetchJSON('/api/readiness');
  } catch {
    return null;
  }
}

// ── Competitors ─────────────────────────────────────────────

export async function getCompetitors(projectId: string): Promise<Competitor[]> {
  const data = await fetchJSON<{ competitors: Competitor[] }>(
    `${BASE}/projects/${projectId}/competitors`
  );
  return data.competitors ?? [];
}

export async function createCompetitor(
  projectId: string,
  body: { name: string; url: string }
): Promise<Competitor> {
  return fetchJSON<Competitor>(`${BASE}/projects/${projectId}/competitors`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function deleteCompetitor(id: string): Promise<void> {
  await fetchJSON<{ status: string }>(`${BASE}/competitors/${id}`, {
    method: 'DELETE',
  });
}

export async function inspectCompetitor(id: string): Promise<InspectResult> {
  return fetchJSON<InspectResult>(`${BASE}/competitors/${id}/inspect`, {
    method: 'POST',
  });
}
