import { useParams, useNavigate } from 'react-router-dom';
import {
  Flex, Heading, Tabs, TabList, TabPanels, Item, Text, View,
  TableView, TableHeader, Column, TableBody, Row, Cell,
  ProgressCircle, Badge, Breadcrumbs, StatusLight,
} from '@adobe/react-spectrum';
import { useKeywords } from '../hooks/useKeywords';
import { useEffect, useState } from 'react';
import { getProject, getKeywords, getSERPResults, getClusters, getGaps, getJob, Project, Keyword, Job, SERPResult, Cluster, ContentGap, Competitor, InspectKeyword, getCompetitors, createCompetitor, deleteCompetitor, inspectCompetitor } from '../api/client';
import StatusBadge from '../components/StatusBadge';

export default function ProjectDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [project, setProject] = useState<Project | null>(null);
  const { keywords, loading: kwLoading, refetch } = useKeywords(id);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [serpResults, setSerpResults] = useState<SERPResult[]>([]);
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [gaps, setGaps] = useState<ContentGap[]>([]);
  const [serpLoading, setSerpLoading] = useState(true);
  const [clustersLoading, setClustersLoading] = useState(true);
  const [gapsLoading, setGapsLoading] = useState(true);
  const [competitors, setCompetitors] = useState<Competitor[]>([]);
  const [competitorsLoading, setCompetitorsLoading] = useState(true);
  const [newCompName, setNewCompName] = useState('');
  const [newCompUrl, setNewCompUrl] = useState('');
  const [addCompLoading, setAddCompLoading] = useState(false);
  const [inspectTarget, setInspectTarget] = useState<Competitor | null>(null);
  const [inspectResults, setInspectResults] = useState<InspectKeyword[]>([]);
  const [inspectLoading, setInspectLoading] = useState(false);
  const [inspectError, setInspectError] = useState<string | null>(null);
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => { if (id) getProject(id).then(setProject).catch(() => {}); }, [id]);
  useEffect(() => {
    const t = setInterval(() => {
      setElapsed(s => s + 1);
      refetch();
      if (id) getSERPResults(id).then(setSerpResults).catch(() => {}).finally(() => setSerpLoading(false));
      if (id) getClusters(id).then(setClusters).catch(() => {}).finally(() => setClustersLoading(false));
      if (id) getGaps(id).then(setGaps).catch(() => {}).finally(() => setGapsLoading(false));
      if (id) getCompetitors(id).then(setCompetitors).catch(() => {}).finally(() => setCompetitorsLoading(false));
    }, 3000);
    return () => clearInterval(t);
  }, [refetch, id]);

  const seeds = keywords.filter(k => k.source === 'seed');
  const autocomplete = keywords.filter(k => k.source === 'autocomplete');
  const serpCrawled = keywords.filter(k => k.source === 'serp_crawl');

  const handleAddCompetitor = async () => {
    if (!id || !newCompName.trim() || !newCompUrl.trim()) return;
    setAddCompLoading(true);
    try {
      let url = newCompUrl.trim();
      if (!url.startsWith('http://') && !url.startsWith('https://')) {
        url = 'https://' + url;
      }
      const c = await createCompetitor(id, { name: newCompName.trim(), url });
      setCompetitors(prev => [c, ...prev]);
      setNewCompName('');
      setNewCompUrl('');
    } catch (e) {
      alert('Erro ao adicionar concorrente: ' + (e instanceof Error ? e.message : String(e)));
    } finally {
      setAddCompLoading(false);
    }
  };

  const handleDeleteCompetitor = async (compId: string) => {
    try {
      await deleteCompetitor(compId);
      setCompetitors(prev => prev.filter(c => c.id !== compId));
    } catch (e) {
      alert('Erro ao remover: ' + (e instanceof Error ? e.message : String(e)));
    }
  };

  const handleInspect = async (comp: Competitor) => {
    setInspectTarget(comp);
    setInspectResults([]);
    setInspectLoading(true);
    setInspectError(null);
    try {
      const result = await inspectCompetitor(comp.id);
      setInspectResults(result.keywords ?? []);
    } catch (e) {
      setInspectError(e instanceof Error ? e.message : String(e));
    } finally {
      setInspectLoading(false);
    }
  };

  // Group SERP results by keyword
  const serpByKeyword = new Map<string, SERPResult[]>();
  for (const r of serpResults) {
    const kw = r.keyword || '(desconhecido)';
    if (!serpByKeyword.has(kw)) serpByKeyword.set(kw, []);
    serpByKeyword.get(kw)!.push(r);
  }

  return (
    <Flex direction="column" gap="size-300">
      <Breadcrumbs onAction={(k: string | number) => navigate(String(k))}>
        <Item key="projects">Projetos</Item>
        <Item>{project?.name || '...'}</Item>
      </Breadcrumbs>
      <Flex justifyContent="space-between" alignItems="center">
        <Heading level={1}>{project?.name || 'Carregando...'}</Heading>
        <StatusLight variant="positive">Live · {elapsed}s</StatusLight>
      </Flex>
      {project && <Text>{project.business_desc}</Text>}

      <Tabs>
        <TabList>
          <Item key="pipeline">Pipeline</Item>
          <Item key="keywords">Keywords ({keywords.length})</Item>
          <Item key="serp">SERP ({serpResults.length})</Item>
          <Item key="clusters">Clusters ({clusters.length})</Item>
          <Item key="gaps">Gaps ({gaps.length})</Item>
          <Item key="inspect">Inspect Keywords</Item>
        </TabList>
        <TabPanels>
          <Item key="pipeline">
            <View padding="size-200">
            <Flex direction="column" gap="size-200">
              <View backgroundColor="gray-75" padding="size-200" borderRadius="medium">
                <Text><strong>1. Seeds</strong> — {seeds.length} keywords iniciais</Text>
                <Flex gap="size-100" wrap>
                  {seeds.slice(0, 5).map(k => <Badge key={k.id} variant="positive">{k.keyword}</Badge>)}
                </Flex>
              </View>
              <View backgroundColor="gray-75" padding="size-200" borderRadius="medium">
                <Text><strong>2. Google Autocomplete</strong> — {autocomplete.length} sugestões</Text>
                <Flex gap="size-100" wrap>
                  {autocomplete.slice(0, 10).map(k => <Badge key={k.id} variant="info">{k.keyword}</Badge>)}
                </Flex>
              </View>
              <View backgroundColor="gray-75" padding="size-200" borderRadius="medium">
                <Text><strong>3. SERP Crawl</strong> — {serpCrawled.length} keywords rastreadas</Text>
                <StatusBadge status={serpCrawled.length > 0 ? 'ok' : 'PENDING'} />
              </View>
              <View backgroundColor="gray-75" padding="size-200" borderRadius="medium">
                <Text><strong>4. Analytics</strong> — {clusters.length} clusters, intent classificada</Text>
                <StatusBadge status={clusters.length > 0 ? 'ok' : 'PENDING'} />
              </View>
            </Flex>
            </View>
          </Item>

          <Item key="keywords">
            {kwLoading ? (
              <ProgressCircle isIndeterminate aria-label="Carregando" />
            ) : (
              <TableView aria-label="Keywords">
                <TableHeader>
                  <Column>Keyword</Column>
                  <Column>Source</Column>
                  <Column>Intent</Column>
                  <Column>Cluster</Column>
                </TableHeader>
                <TableBody>
                  {keywords.map(kw => (
                    <Row key={kw.id}>
                      <Cell>{kw.keyword}</Cell>
                      <Cell><Badge variant={kw.source === 'seed' ? 'positive' : kw.source === 'autocomplete' ? 'info' : 'neutral'}>{kw.source || '—'}</Badge></Cell>
                      <Cell><StatusBadge status={kw.intent || 'PENDING'} /></Cell>
                      <Cell>{kw.cluster_name || '—'}</Cell>
                    </Row>
                  ))}
                </TableBody>
              </TableView>
            )}
          </Item>

          <Item key="serp">
            {serpLoading && serpResults.length === 0 ? (
              <View padding="size-400">
                <Flex justifyContent="center">
                  <ProgressCircle isIndeterminate aria-label="Carregando SERP" />
                </Flex>
              </View>
            ) : serpResults.length === 0 ? (
              <View padding="size-400">
                <Text>Nenhum resultado SERP ainda. O crawler pode estar processando — aguarde alguns segundos.</Text>
              </View>
            ) : (
              <View padding="size-100">
                <Flex direction="column" gap="size-150">
                  {[...serpByKeyword.entries()].map(([keyword, results]) => (
                    <View
                      key={keyword}
                      backgroundColor="gray-75"
                      padding="size-150"
                      borderRadius="medium"
                    >
                      <Flex direction="column" gap="size-100">
                        <Flex alignItems="center" gap="size-100">
                          <Badge variant="positive">{keyword}</Badge>
                          <Text><strong>{results.length}</strong> resultados</Text>
                        </Flex>
                        <TableView aria-label={`SERP — ${keyword}`} density="compact">
                          <TableHeader>
                            <Column width={60}>#</Column>
                            <Column>URL</Column>
                            <Column>Título</Column>
                            <Column>Descrição</Column>
                          </TableHeader>
                          <TableBody>
                            {results.map(r => (
                              <Row key={r.id}>
                                <Cell>
                                  <Badge variant={r.position <= 3 ? 'positive' : r.position <= 10 ? 'info' : 'neutral'}>
                                    {r.position}
                                  </Badge>
                                </Cell>
                                <Cell>
                                  <a
                                    href={r.url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    style={{ color: 'var(--spectrum-global-color-blue-600)', fontSize: '0.85em' }}
                                  >
                                    {new URL(r.url).hostname}{new URL(r.url).pathname.slice(0, 30)}
                                  </a>
                                </Cell>
                                <Cell><Text>{r.title || '—'}</Text></Cell>
                                <Cell><Text>{r.snippet?.slice(0, 120) || '—'}</Text></Cell>
                              </Row>
                            ))}
                          </TableBody>
                        </TableView>
                      </Flex>
                    </View>
                  ))}
                </Flex>
              </View>
            )}
          </Item>

          <Item key="clusters">
            {clustersLoading && clusters.length === 0 ? (
              <View padding="size-400">
                <Flex justifyContent="center">
                  <ProgressCircle isIndeterminate aria-label="Carregando clusters" />
                </Flex>
              </View>
            ) : clusters.length === 0 ? (
              <View padding="size-400">
                <Text>Nenhum cluster ainda. O analytics roda após o SERP crawl — aguarde.</Text>
              </View>
            ) : (
              <View padding="size-100">
                <Flex direction="column" gap="size-150">
                  {clusters.map(c => (
                    <View key={c.id} backgroundColor="gray-75" padding="size-150" borderRadius="medium">
                      <Flex direction="column" gap="size-75">
                        <Flex alignItems="center" gap="size-100">
                          <Text><strong>{c.name}</strong></Text>
                          <Badge variant="info">{c.intent || '—'}</Badge>
                          <Text>· {c.keyword_count} keywords</Text>
                        </Flex>
                      </Flex>
                    </View>
                  ))}
                </Flex>
              </View>
            )}
          </Item>

          <Item key="gaps">
            {gapsLoading && gaps.length === 0 ? (
              <View padding="size-400">
                <Flex justifyContent="center">
                  <ProgressCircle isIndeterminate aria-label="Carregando gaps" />
                </Flex>
              </View>
            ) : gaps.length === 0 ? (
              <View padding="size-400">
                <Text>Nenhum gap detectado ainda. O analytics roda após o SERP crawl — aguarde.</Text>
              </View>
            ) : (
              <View padding="size-100">
                <Flex direction="column" gap="size-150">
                  {gaps.map(g => (
                    <View key={g.id} backgroundColor="gray-75" padding="size-150" borderRadius="medium">
                      <Flex direction="column" gap="size-100">
                        <Flex alignItems="center" gap="size-100">
                          <Badge variant="positive">{g.keyword}</Badge>
                        </Flex>
                        <Text>
                          {g.gaps.split('\n').map((line: string, i: number) => (
                            <span key={i}>{line}<br /></span>
                          ))}
                        </Text>
                      </Flex>
                    </View>
                  ))}
                </Flex>
              </View>
            )}
          </Item>

          <Item key="inspect">
            <View padding="size-200">
              <Flex direction="column" gap="size-200">
                {/* Add Competitor Form */}
                <View backgroundColor="gray-75" padding="size-200" borderRadius="medium">
                  <Flex direction="column" gap="size-100">
                    <Text><strong>Adicionar Concorrente</strong></Text>
                    <Flex gap="size-100" alignItems="end" wrap>
                      <div style={{ flex: 1, minWidth: 150 }}>
                        <label style={{ fontSize: 12, color: 'var(--spectrum-global-color-gray-600)' }}>Nome</label>
                        <input
                          type="text"
                          value={newCompName}
                          onChange={e => setNewCompName(e.target.value)}
                          placeholder="Ex: Concorrente A"
                          style={{ width: '100%', padding: '6px 10px', borderRadius: 6, border: '1px solid var(--spectrum-global-color-gray-400)', fontSize: 14 }}
                        />
                      </div>
                      <div style={{ flex: 2, minWidth: 200 }}>
                        <label style={{ fontSize: 12, color: 'var(--spectrum-global-color-gray-600)' }}>Site URL</label>
                        <input
                          type="text"
                          value={newCompUrl}
                          onChange={e => setNewCompUrl(e.target.value)}
                          placeholder="https://exemplo.com"
                          style={{ width: '100%', padding: '6px 10px', borderRadius: 6, border: '1px solid var(--spectrum-global-color-gray-400)', fontSize: 14 }}
                        />
                      </div>
                      <button
                        onClick={handleAddCompetitor}
                        disabled={addCompLoading || !newCompName.trim() || !newCompUrl.trim()}
                        style={{
                          padding: '6px 16px',
                          borderRadius: 6,
                          border: 'none',
                          background: 'var(--spectrum-global-color-blue-500)',
                          color: '#fff',
                          fontSize: 14,
                          cursor: 'pointer',
                          opacity: addCompLoading ? 0.6 : 1,
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {addCompLoading ? 'Adicionando...' : '+ Adicionar'}
                      </button>
                    </Flex>
                  </Flex>
                </View>

                {/* Competitor List */}
                {competitorsLoading ? (
                  <Flex justifyContent="center"><ProgressCircle isIndeterminate aria-label="Carregando concorrentes" /></Flex>
                ) : competitors.length === 0 ? (
                  <View padding="size-200">
                    <Text>Nenhum concorrente adicionado. Adicione concorrentes acima para inspecionar keywords de seus sites.</Text>
                  </View>
                ) : (
                  <Flex direction="column" gap="size-100">
                    {competitors.map(comp => (
                      <View key={comp.id} backgroundColor="gray-75" padding="size-150" borderRadius="medium">
                        <Flex justifyContent="space-between" alignItems="center">
                          <Flex direction="column" gap="size-50">
                            <Text><strong>{comp.name}</strong></Text>
                            <a
                              href={comp.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              style={{ color: 'var(--spectrum-global-color-blue-600)', fontSize: 13 }}
                            >
                              {comp.url}
                            </a>
                          </Flex>
                          <Flex gap="size-100">
                            <button
                              onClick={() => handleInspect(comp)}
                              disabled={inspectLoading && inspectTarget?.id === comp.id}
                              style={{
                                padding: '5px 12px', borderRadius: 6, border: 'none',
                                background: 'var(--spectrum-global-color-green-500)', color: '#fff',
                                fontSize: 13, cursor: 'pointer',
                                opacity: inspectLoading && inspectTarget?.id === comp.id ? 0.6 : 1,
                              }}
                            >
                              {inspectLoading && inspectTarget?.id === comp.id ? '⏳ Inspecionando...' : '🔍 Inspecionar'}
                            </button>
                            <button
                              onClick={() => handleDeleteCompetitor(comp.id)}
                              style={{
                                padding: '5px 10px', borderRadius: 6, border: 'none',
                                background: 'var(--spectrum-global-color-red-400)', color: '#fff',
                                fontSize: 13, cursor: 'pointer',
                              }}
                            >
                              ✕
                            </button>
                          </Flex>
                        </Flex>
                      </View>
                    ))}
                  </Flex>
                )}

                {/* Inspect Results Modal */}
                {inspectTarget && (
                  <div
                    style={{
                      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
                      background: 'rgba(0,0,0,0.5)', display: 'flex',
                      alignItems: 'center', justifyContent: 'center', zIndex: 9999,
                    }}
                    onClick={(e: React.MouseEvent) => { if (e.target === e.currentTarget) setInspectTarget(null); }}
                  >
                    <div
                      style={{
                        background: '#fff', borderRadius: 8, padding: 24,
                        maxWidth: 800, width: '90%', maxHeight: '80vh', overflow: 'auto',
                      }}
                    >
                      <Flex justifyContent="space-between" alignItems="center" marginBottom="size-200">
                        <Heading level={3}>Keywords — {inspectTarget.name}</Heading>
                        <button
                          onClick={() => setInspectTarget(null)}
                          style={{
                            background: 'none', border: 'none', fontSize: 20,
                            cursor: 'pointer', padding: '0 8px',
                          }}
                        >
                          ✕
                        </button>
                      </Flex>

                      {inspectLoading ? (
                        <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 32, alignItems: 'center', gap: 12 }}>
                          <ProgressCircle isIndeterminate aria-label="Extraindo keywords" />
                          <Text>A analisar com Gemma4 — isto pode demorar 10-30 segundos...</Text>
                        </div>
                      ) : inspectError ? (
                        <div style={{ padding: 20, background: '#fff3f3', borderRadius: 8, border: '1px solid #ffcccc' }}>
                          <Text><strong style={{ color: '#cc0000' }}>⚠ Erro na extração:</strong></Text>
                          <Text>{inspectError}</Text>
                          <Text><br/><em style={{ fontSize: 12, color: '#666' }}>Tente novamente ou verifique se o Ollama está a correr.</em></Text>
                        </div>
                      ) : inspectResults.length === 0 ? (
                        <View padding="size-200">
                          <Text>Nenhuma keyword extraída deste site.</Text>
                        </View>
                      ) : (
                        <TableView aria-label="Inspect Keywords" density="compact">
                          <TableHeader>
                            <Column width={40}>#</Column>
                            <Column>Keyword</Column>
                          </TableHeader>
                          <TableBody>
                            {inspectResults.map((kw, i) => (
                              <Row key={i}>
                                <Cell><Text>{i + 1}</Text></Cell>
                                <Cell><Text><strong>{kw.keyword}</strong></Text></Cell>
                              </Row>
                            ))}
                          </TableBody>
                        </TableView>
                      )}
                    </div>
                  </div>
                )}
              </Flex>
            </View>
          </Item>
        </TabPanels>
      </Tabs>
    </Flex>
  );
}
