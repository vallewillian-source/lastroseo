import { Routes, Route } from 'react-router-dom';
import { Flex, IllustratedMessage, Content } from '@adobe/react-spectrum';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Projects from './pages/Projects';
import ProjectDetail from './pages/ProjectDetail';
import Services from './pages/Services';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/projects" element={<Projects />} />
        <Route path="/projects/:id" element={<ProjectDetail />} />
        <Route path="/services" element={<Services />} />
        <Route path="*" element={
          <Flex direction="column" alignItems="center" justifyContent="center" height="100%">
            <IllustratedMessage>
              <Content>Página não encontrada</Content>
            </IllustratedMessage>
          </Flex>
        } />
      </Route>
    </Routes>
  );
}
