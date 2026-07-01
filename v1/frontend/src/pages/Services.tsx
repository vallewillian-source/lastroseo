import { Flex, Heading, View, Text, Grid } from '@adobe/react-spectrum';
import { useHealth } from '../hooks/useHealth';
import StatusBadge from '../components/StatusBadge';

export default function Services() {
  const health = useHealth();
  const s = health?.services;

  return (
    <Flex direction="column" gap="size-300">
      <Heading level={1}>Status dos Serviços</Heading>
      <Grid columns={['1fr', '1fr']} gap="size-200">
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text><strong>Gateway API</strong> (:8081)</Text>
          <br />
          <StatusBadge status={s?.gateway || 'down'} />
        </View>
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text><strong>PostgreSQL</strong> (:5432)</Text>
          <br />
          <StatusBadge status={s?.postgres || 'down'} />
        </View>
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text><strong>Redis</strong> (:6379)</Text>
          <br />
          <StatusBadge status={s?.redis || 'down'} />
        </View>
        <View backgroundColor="gray-100" padding="size-300" borderRadius="medium">
          <Text><strong>SearXNG</strong> (:8080)</Text>
          <br />
          <StatusBadge status={s?.searxng || 'down'} />
        </View>
      </Grid>
    </Flex>
  );
}
