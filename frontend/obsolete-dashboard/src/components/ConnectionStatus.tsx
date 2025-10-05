// WebSocket connection status indicator

import type { ConnectionStatus } from '../api/websocket';

interface Props {
    status: ConnectionStatus;
}

export function ConnectionStatus({ status }: Props) {
    const getStatusInfo = () => {
        switch (status) {
            case 'connected':
                return {
                    color: '#10b981',
                    icon: '●',
                    text: 'Connected',
                    pulse: false,
                };
            case 'connecting':
                return {
                    color: '#f59e0b',
                    icon: '●',
                    text: 'Connecting...',
                    pulse: true,
                };
            case 'disconnected':
                return {
                    color: '#6b7280',
                    icon: '●',
                    text: 'Disconnected',
                    pulse: false,
                };
            case 'error':
                return {
                    color: '#ef4444',
                    icon: '●',
                    text: 'Error',
                    pulse: true,
                };
        }
    };

    const info = getStatusInfo();

    return (
        <div
            className="connection-status"
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.5rem 1rem',
                backgroundColor: 'rgba(0, 0, 0, 0.2)',
                borderRadius: '0.5rem',
                fontSize: '0.875rem',
            }}
        >
            <span
                style={{
                    color: info.color,
                    fontSize: '1.5rem',
                    lineHeight: '1',
                    animation: info.pulse ? 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite' : 'none',
                }}
            >
                {info.icon}
            </span>
            <span style={{ color: '#e5e7eb', fontWeight: 500 }}>{info.text}</span>
        </div>
    );
}

