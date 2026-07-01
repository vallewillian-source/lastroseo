import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  Flex, View, Text,
} from '@adobe/react-spectrum';
import Home from '@spectrum-icons/workflow/Home';
import Project from '@spectrum-icons/workflow/Project';
import Devices from '@spectrum-icons/workflow/Devices';

interface NavItem { key: string; label: string; icon: React.ReactNode }

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();

  const items: NavItem[] = [
    { key: '/', label: 'Dashboard', icon: <Home /> },
    { key: '/projects', label: 'Projetos', icon: <Project /> },
    { key: '/services', label: 'Serviços', icon: <Devices /> },
  ];

  const selected = items.find(i =>
    location.pathname === i.key || location.pathname.startsWith(i.key + '/')
  );

  return (
    <Flex direction="row" height="100vh">
      <View
        backgroundColor="gray-100"
        padding="size-200"
        width="size-3000"
        minWidth="size-3000"
        borderWidth="thin"
        borderColor="gray-300"
      >
        <Flex direction="column" gap="size-200">
          <Text>
            <span style={{ fontSize: 18, fontWeight: 700 }}>LastroSEO</span>
          </Text>
          <Flex direction="column" gap="size-50">
            {items.map(i => (
              <View
                key={i.key}
                backgroundColor={selected?.key === i.key ? 'blue-500' : 'transparent'}
                padding="size-100"
                borderRadius="medium"
              >
                <button
                  onClick={() => navigate(i.key)}
                  style={{
                    background: 'none', border: 'none', color: 'inherit',
                    cursor: 'pointer', display: 'flex', gap: 8, alignItems: 'center',
                    fontSize: 14,
                  }}
                >
                  {i.icon} {i.label}
                </button>
              </View>
            ))}
          </Flex>
        </Flex>
      </View>
      <View flex padding="size-300" overflow="auto">
        <Outlet />
      </View>
    </Flex>
  );
}
