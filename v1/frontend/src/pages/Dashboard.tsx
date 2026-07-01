import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Flex, View, Heading, Text, ProgressCircle, Grid,
} from '@adobe/react-spectrum';
import { useProjects } from '../hooks/useProjects';
import { useHealth } from '../hooks/useHealth';
import StatusBadge from '../components/StatusBadge';

export default function Dashboard() {
  const { projects, loading } = useProjects();
  const health = useHealth();
  const navigate = useNavigate();
  const [elapsed, setElapsed] = useState(0);
  useEffect(() => {
    const t = setInterval(() => setElapsed(s => s + 1), 1000);
    return () => clearInterval(t);
  }, []);

  return (
    <Flex direction="column" gap="size-300">
      <Heading level={1}>Dashboard</Heading>
      <Text>Sessão ativa: {Math.floor(elapsed / 60)}m {elapsed % 60}s</Text>

      <Grid
        columns={['1fr', '1fr', '1fr']}
        gap="size-200"
      >
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text>Projetos</Text>
          <Heading level={2}>{loading ? '...' : projects.length}</Heading>
        </View>
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text>Gateway</Text>
          <StatusBadge status={health?.services?.gateway || 'down'} />
        </View>
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text>PostgreSQL</Text>
          <StatusBadge status={health?.services?.postgres || 'down'} />
        </View>
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text>SearXNG</Text>
          <StatusBadge status={health?.services?.searxng || 'down'} />
        </View>
      </Grid>

      <Heading level={2}>Projetos Recentes</Heading>
      {loading ? (
        <ProgressCircle isIndeterminate aria-label="Carregando projetos" />
      ) : (
        <Grid columns={['1fr']} gap="size-100">
          {projects.slice(0, 5).map(p => (
            <div
              key={p.id}
              onClick={() => navigate(`/projects/${p.id}`)}
              style={{
                background: 'var(--spectrum-gray-75)', padding: 16,
                borderRadius: 8, border: '1px solid var(--spectrum-gray-300)',
                cursor: 'pointer', marginBottom: 8,
              }}
            >
              <Text><strong>{p.name}</strong></Text>
              <br />
              <Text>{p.business_desc?.slice(0, 80)}</Text>
            </div>
          ))}
          {projects.length === 0 && <Text>Nenhum projeto criado.</Text>}
        </Grid>
      )}
    </Flex>
  );
}
