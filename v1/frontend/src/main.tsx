import React from 'react';
import ReactDOM from 'react-dom/client';
import { defaultTheme, Provider } from '@adobe/react-spectrum';
import { BrowserRouter } from 'react-router-dom';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider theme={defaultTheme} colorScheme="dark" height="100vh">
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </Provider>
  </React.StrictMode>
);
