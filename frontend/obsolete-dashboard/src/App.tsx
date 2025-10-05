// Main App component with routing and WebSocket integration

import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom';
import { MetricsWebSocket } from './api/websocket';
import { useMetricsStore } from './stores/metricsStore';
import { ConnectionStatus } from './components/ConnectionStatus';
import { Overview } from './pages/Overview';
import { Proxies } from './pages/Proxies';
import { Nodes } from './pages/Nodes';
import { Namespaces } from './pages/Namespaces';
import { CacheQuery } from './pages/CacheQuery';
import { Events } from './pages/Events';
import type { WebSocketMessage } from './types/metrics';
import './App.css';

function App() {
  const {
    updateOverview,
    updateProxy,
    updateNode,
    addEvent,
    error,
    wsStatus,
    setWsStatus,
  } = useMetricsStore();

  useEffect(() => {
    const ws = new MetricsWebSocket();

    ws.connect(
      (message: WebSocketMessage) => {
        switch (message.type) {
          case 'overview_update':
            updateOverview(message.data);
            break;
          case 'proxy_update':
            updateProxy(message.data);
            break;
          case 'node_update':
            updateNode(message.data);
            break;
          case 'event':
            addEvent(message.data);
            break;
        }
      },
      (status) => {
        setWsStatus(status);
      }
    );

    return () => {
      ws.disconnect();
    };
  }, [updateOverview, updateProxy, updateNode, addEvent, setWsStatus]);

  return (
    <BrowserRouter>
      <div className="app">
        <nav className="navbar">
          <div className="navbar-brand">
            <img
              src="https://yao-verse.eggybyte.com/favicon.png"
              alt="Yao-Oracle Logo"
              className="logo"
            />
            <h1>Yao-Oracle Dashboard</h1>
          </div>
          <div className="navbar-links">
            <Link to="/overview">Overview</Link>
            <Link to="/proxies">Proxies</Link>
            <Link to="/nodes">Nodes</Link>
            <Link to="/namespaces">Namespaces</Link>
            <Link to="/cache">Cache Query</Link>
            <Link to="/events">Events</Link>
          </div>
          <ConnectionStatus status={wsStatus} />
        </nav>

        {error && (
          <div className="error-banner">
            ⚠️ Error: {error}
          </div>
        )}

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Navigate to="/overview" replace />} />
            <Route path="/overview" element={<Overview />} />
            <Route path="/proxies" element={<Proxies />} />
            <Route path="/nodes" element={<Nodes />} />
            <Route path="/namespaces" element={<Namespaces />} />
            <Route path="/cache" element={<CacheQuery />} />
            <Route path="/events" element={<Events />} />
          </Routes>
        </main>

        <footer className="footer">
          <p>
            Yao-Oracle Distributed Cache System • Real-time monitoring powered by
            WebSocket
          </p>
        </footer>
      </div>
    </BrowserRouter>
  );
}

export default App;
