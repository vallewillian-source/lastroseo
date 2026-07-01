import { useNavigate } from 'react-router-dom';
import {
  Flex, Heading, View, TableView, TableHeader, Column, TableBody, Row, Cell,
  ProgressCircle, Text,
} from '@adobe/react-spectrum';
import { useProjects } from '../hooks/useProjects';
import ProjectForm from '../components/ProjectForm';

export default function Projects() {
  const { projects, loading, create } = useProjects();
  const navigate = useNavigate();

  return (
    <Flex direction="column" gap="size-300">
      <Flex justifyContent="space-between" alignItems="center">
        <Heading level={1}>Projetos</Heading>
        <ProjectForm onCreate={create} />
      </Flex>

      {loading ? (
        <ProgressCircle isIndeterminate aria-label="Carregando" />
      ) : (
        <TableView
          aria-label="Lista de projetos"
          selectionMode="none"
          onAction={(key: string | number) => navigate(`/projects/${key}`)}
        >
          <TableHeader>
            <Column>Nome</Column>
            <Column>Descrição</Column>
            <Column>Criado em</Column>
          </TableHeader>
          <TableBody>
            {projects.map(p => (
              <Row key={p.id}>
                <Cell><Text><strong>{p.name}</strong></Text></Cell>
                <Cell>{p.business_desc?.slice(0, 100) || '—'}</Cell>
                <Cell>{new Date(p.created_at).toLocaleDateString('pt-BR')}</Cell>
              </Row>
            ))}
          </TableBody>
        </TableView>
      )}
    </Flex>
  );
}
