// Events & Health monitoring page

import { useMetricsStore } from '../stores/metricsStore';
import type { ClusterEvent } from '../types/metrics';

export function Events() {
    const { events } = useMetricsStore();

    const getSeverityBadge = (severity: ClusterEvent['severity']) => {
        const styles = {
            info: {
                bg: 'rgba(59, 130, 246, 0.1)',
                border: '#3b82f6',
                color: '#60a5fa',
                icon: 'ℹ️',
            },
            warning: {
                bg: 'rgba(245, 158, 11, 0.1)',
                border: '#f59e0b',
                color: '#fbbf24',
                icon: '⚠️',
            },
            error: {
                bg: 'rgba(239, 68, 68, 0.1)',
                border: '#ef4444',
                color: '#f87171',
                icon: '❌',
            },
        };

        const style = styles[severity];

        return (
            <span
                style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    padding: '0.25rem 0.75rem',
                    backgroundColor: style.bg,
                    border: `1px solid ${style.border}`,
                    borderRadius: '0.375rem',
                    color: style.color,
                    fontSize: '0.875rem',
                    fontWeight: 500,
                }}
            >
                <span>{style.icon}</span>
                <span>{severity.toUpperCase()}</span>
            </span>
        );
    };

    const formatTimestamp = (timestamp: string) => {
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now.getTime() - date.getTime();
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (seconds < 60) return `${seconds}s ago`;
        if (minutes < 60) return `${minutes}m ago`;
        if (hours < 24) return `${hours}h ago`;
        return `${days}d ago`;
    };

    const getSeverityCounts = () => {
        const counts = { info: 0, warning: 0, error: 0 };
        events.forEach((event) => {
            counts[event.severity]++;
        });
        return counts;
    };

    const severityCounts = getSeverityCounts();

    return (
        <div className="page">
            <h1>Events & Health Status</h1>

            <div className="metric-cards">
                <div
                    className="metric-card"
                    style={{ backgroundColor: 'rgba(59, 130, 246, 0.1)', borderColor: '#3b82f6' }}
                >
                    <div className="metric-icon">ℹ️</div>
                    <div className="metric-content">
                        <h3 className="metric-title">Info Events</h3>
                        <div className="metric-value">{severityCounts.info}</div>
                    </div>
                </div>

                <div
                    className="metric-card"
                    style={{ backgroundColor: 'rgba(245, 158, 11, 0.1)', borderColor: '#f59e0b' }}
                >
                    <div className="metric-icon">⚠️</div>
                    <div className="metric-content">
                        <h3 className="metric-title">Warnings</h3>
                        <div className="metric-value">{severityCounts.warning}</div>
                    </div>
                </div>

                <div
                    className="metric-card"
                    style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', borderColor: '#ef4444' }}
                >
                    <div className="metric-icon">❌</div>
                    <div className="metric-content">
                        <h3 className="metric-title">Errors</h3>
                        <div className="metric-value">{severityCounts.error}</div>
                    </div>
                </div>

                <div
                    className="metric-card"
                    style={{ backgroundColor: 'rgba(16, 185, 129, 0.1)', borderColor: '#10b981' }}
                >
                    <div className="metric-icon">📊</div>
                    <div className="metric-content">
                        <h3 className="metric-title">Total Events</h3>
                        <div className="metric-value">{events.length}</div>
                        <div className="metric-subtitle">Last 50 events</div>
                    </div>
                </div>
            </div>

            <div className="events-container" style={{ marginTop: '2rem' }}>
                <h2 style={{ marginBottom: '1rem', color: '#e5e7eb' }}>Event Stream</h2>

                {events.length === 0 ? (
                    <div
                        style={{
                            padding: '4rem 2rem',
                            textAlign: 'center',
                            color: '#9ca3af',
                            backgroundColor: 'rgba(0, 0, 0, 0.2)',
                            borderRadius: '0.5rem',
                        }}
                    >
                        <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>🎉</div>
                        <h3>No Events</h3>
                        <p>All systems are running smoothly!</p>
                    </div>
                ) : (
                    <div className="events-list">
                        {events.map((event, idx) => (
                            <div
                                key={idx}
                                className="event-item"
                                style={{
                                    padding: '1rem',
                                    marginBottom: '0.75rem',
                                    backgroundColor: 'rgba(0, 0, 0, 0.2)',
                                    borderRadius: '0.5rem',
                                    borderLeft: `4px solid ${event.severity === 'error'
                                            ? '#ef4444'
                                            : event.severity === 'warning'
                                                ? '#f59e0b'
                                                : '#3b82f6'
                                        }`,
                                    display: 'flex',
                                    alignItems: 'flex-start',
                                    gap: '1rem',
                                }}
                            >
                                <div style={{ flex: '0 0 auto' }}>
                                    {getSeverityBadge(event.severity)}
                                </div>
                                <div style={{ flex: '1 1 auto' }}>
                                    <p
                                        style={{
                                            color: '#e5e7eb',
                                            margin: 0,
                                            fontSize: '0.95rem',
                                            lineHeight: '1.5',
                                        }}
                                    >
                                        {event.message}
                                    </p>
                                </div>
                                <div
                                    style={{
                                        flex: '0 0 auto',
                                        color: '#9ca3af',
                                        fontSize: '0.875rem',
                                        whiteSpace: 'nowrap',
                                    }}
                                >
                                    {formatTimestamp(event.timestamp)}
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            <div className="events-info" style={{ marginTop: '2rem', color: '#9ca3af' }}>
                <p>
                    💡 <strong>Tip:</strong> Events are pushed in real-time via WebSocket. The
                    system automatically tracks node status changes, proxy anomalies, hit ratio
                    fluctuations, and other important events.
                </p>
            </div>
        </div>
    );
}

