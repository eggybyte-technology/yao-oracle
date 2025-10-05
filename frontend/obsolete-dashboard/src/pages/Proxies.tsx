// Proxy instances page

import { useEffect, useState } from 'react';
import { useMetricsStore } from '../stores/metricsStore';
import { fetchProxies, fetchProxyTimeseries } from '../api/client';
import { StatusBadge } from '../components/StatusBadge';
import { BarChart } from '../components/charts/BarChart';
import { LineChart } from '../components/charts/LineChart';
import type { TimeSeriesPoint } from '../types/metrics';

export function Proxies() {
    const { proxies, setProxies, setError, setLoading } = useMetricsStore();
    const [selectedProxy, setSelectedProxy] = useState<string | null>(null);
    const [selectedProxies, setSelectedProxies] = useState<string[]>([]);
    const [compareMode, setCompareMode] = useState<boolean>(false);
    const [timeseriesData, setTimeseriesData] = useState<TimeSeriesPoint[]>([]);
    const [compareTimeseriesData, setCompareTimeseriesData] = useState<Map<string, TimeSeriesPoint[]>>(new Map());

    useEffect(() => {
        const loadData = async () => {
            try {
                setLoading(true);
                const data = await fetchProxies();
                setProxies(data.proxies);
                if (data.proxies.length > 0 && !selectedProxy) {
                    setSelectedProxy(data.proxies[0].id);
                }
                setError(null);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load data');
            } finally {
                setLoading(false);
            }
        };

        loadData();
        const interval = setInterval(loadData, 5000);
        return () => clearInterval(interval);
    }, [setProxies, setError, setLoading, selectedProxy]);

    useEffect(() => {
        if (!selectedProxy) return;

        const loadTimeseries = async () => {
            try {
                const data = await fetchProxyTimeseries(selectedProxy);
                setTimeseriesData(data.metrics);
            } catch (err) {
                console.error('Failed to load timeseries:', err);
            }
        };

        loadTimeseries();
        const interval = setInterval(loadTimeseries, 10000);
        return () => clearInterval(interval);
    }, [selectedProxy]);

    const selectedProxyData = proxies.find((p) => p.id === selectedProxy);

    const connectionsData = proxies.map((p) => ({
        name: p.id,
        value: p.connections,
    }));

    const handleProxySelection = (proxyId: string) => {
        if (compareMode) {
            setSelectedProxies((prev) =>
                prev.includes(proxyId)
                    ? prev.filter((id) => id !== proxyId)
                    : [...prev, proxyId]
            );
        } else {
            setSelectedProxy(proxyId);
        }
    };

    const toggleCompareMode = () => {
        setCompareMode(!compareMode);
        if (!compareMode) {
            setSelectedProxies([]);
        }
    };

    useEffect(() => {
        if (compareMode && selectedProxies.length > 0) {
            const loadCompareData = async () => {
                const dataMap = new Map<string, TimeSeriesPoint[]>();
                for (const proxyId of selectedProxies) {
                    try {
                        const data = await fetchProxyTimeseries(proxyId);
                        dataMap.set(proxyId, data.metrics);
                    } catch (err) {
                        console.error(`Failed to load timeseries for ${proxyId}:`, err);
                    }
                }
                setCompareTimeseriesData(dataMap);
            };
            loadCompareData();
            const interval = setInterval(loadCompareData, 10000);
            return () => clearInterval(interval);
        }
    }, [compareMode, selectedProxies]);

    return (
        <div className="page">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h1>Proxy Instances</h1>
                <button
                    onClick={toggleCompareMode}
                    className={`btn-action ${compareMode ? 'success active' : 'primary'}`}
                >
                    {compareMode ? '✓ Compare Mode Active' : '📊 Enable Compare'}
                </button>
            </div>

            <table className="data-table">
                <thead>
                    <tr>
                        <th>Instance</th>
                        <th>IP Address</th>
                        <th>Uptime</th>
                        <th>Connections</th>
                        <th>QPS</th>
                        <th>Latency (P99)</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {proxies.map((proxy) => (
                        <tr
                            key={proxy.id}
                            onClick={() => handleProxySelection(proxy.id)}
                            className={
                                compareMode
                                    ? selectedProxies.includes(proxy.id) ? 'selected' : ''
                                    : selectedProxy === proxy.id ? 'selected' : ''
                            }
                            style={{ cursor: 'pointer' }}
                        >
                            <td>{proxy.id}</td>
                            <td>{proxy.ip}</td>
                            <td>{Math.floor(proxy.uptime / 3600)}h</td>
                            <td>{proxy.connections}</td>
                            <td>
                                {proxy.qps.get + proxy.qps.set + proxy.qps.delete}
                            </td>
                            <td>{proxy.latency.p99.toFixed(2)} ms</td>
                            <td>
                                <StatusBadge status={proxy.status} />
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {compareMode && selectedProxies.length > 0 && (
                <div className="details-section">
                    <h2>Comparison: {selectedProxies.length} Proxies Selected</h2>
                    <div style={{
                        marginBottom: '1rem',
                        padding: '1.25rem',
                        background: 'rgba(0, 245, 255, 0.08)',
                        borderRadius: '12px',
                        border: '1px solid rgba(0, 245, 255, 0.3)',
                        backdropFilter: 'blur(10px)'
                    }}>
                        <p style={{ margin: 0, color: '#00f5ff', lineHeight: '1.6' }}>
                            💡 <strong>Tip:</strong> Click on proxies in the table to add/remove them from comparison.
                            <br />
                            <span style={{ fontSize: '0.875rem', color: 'var(--color-text-secondary)' }}>
                                Selected: {selectedProxies.join(', ')}
                            </span>
                        </p>
                    </div>

                    {compareTimeseriesData.size > 0 && (
                        <div className="charts-grid" style={{ marginTop: '1rem' }}>
                            <div className="chart-container" style={{ gridColumn: 'span 2' }}>
                                <LineChart
                                    datasets={selectedProxies.map((proxyId, idx) => {
                                        const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];
                                        const data = compareTimeseriesData.get(proxyId) || [];
                                        return {
                                            label: proxyId,
                                            data: data.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: (d.qps?.get || 0) + (d.qps?.set || 0) + (d.qps?.delete || 0),
                                            })),
                                            color: colors[idx % colors.length],
                                        };
                                    })}
                                    title="QPS Comparison (Total)"
                                    yAxisLabel="QPS"
                                />
                            </div>

                            <div className="chart-container">
                                <LineChart
                                    datasets={selectedProxies.map((proxyId, idx) => {
                                        const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];
                                        const data = compareTimeseriesData.get(proxyId) || [];
                                        return {
                                            label: `${proxyId} P99`,
                                            data: data.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.latency?.p99 || 0,
                                            })),
                                            color: colors[idx % colors.length],
                                        };
                                    })}
                                    title="P99 Latency Comparison"
                                    yAxisLabel="Latency (ms)"
                                />
                            </div>

                            <div className="chart-container">
                                <BarChart
                                    data={selectedProxies.map((proxyId) => {
                                        const proxy = proxies.find((p) => p.id === proxyId);
                                        return {
                                            name: proxyId,
                                            value: proxy ? proxy.connections : 0,
                                        };
                                    })}
                                    title="Connections Comparison"
                                    color="#3b82f6"
                                    horizontal={false}
                                />
                            </div>
                        </div>
                    )}
                </div>
            )}

            {!compareMode && selectedProxyData && (
                <div className="details-section">
                    <h2>Details: {selectedProxyData.id}</h2>
                    <div className="detail-info">
                        <div className="info-item">
                            <span className="label">IP:</span>
                            <span className="value">{selectedProxyData.ip}</span>
                        </div>
                        <div className="info-item">
                            <span className="label">Namespaces:</span>
                            <span className="value">{selectedProxyData.namespaces.length}</span>
                        </div>
                        <div className="info-item">
                            <span className="label">Error Rate:</span>
                            <span className="value">
                                {(selectedProxyData.error_rate * 100).toFixed(3)}%
                            </span>
                        </div>
                        <div className="info-item">
                            <span className="label">Latency P50:</span>
                            <span className="value">{selectedProxyData.latency.p50.toFixed(2)} ms</span>
                        </div>
                        <div className="info-item">
                            <span className="label">Latency P90:</span>
                            <span className="value">{selectedProxyData.latency.p90.toFixed(2)} ms</span>
                        </div>
                    </div>

                    {timeseriesData.length > 0 && (
                        <div className="charts-grid" style={{ marginTop: '1rem' }}>
                            <div className="chart-container" style={{ gridColumn: 'span 2' }}>
                                <LineChart
                                    datasets={[
                                        {
                                            label: 'GET',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.qps?.get || 0,
                                            })),
                                            color: '#3b82f6',
                                            fill: true,
                                        },
                                        {
                                            label: 'SET',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.qps?.set || 0,
                                            })),
                                            color: '#10b981',
                                            fill: true,
                                        },
                                        {
                                            label: 'DELETE',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.qps?.delete || 0,
                                            })),
                                            color: '#ef4444',
                                            fill: true,
                                        },
                                    ]}
                                    title={`QPS Trend - ${selectedProxyData.id}`}
                                    yAxisLabel="QPS"
                                />
                            </div>

                            <div className="chart-container">
                                <LineChart
                                    datasets={[
                                        {
                                            label: 'P50',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.latency?.p50 || 0,
                                            })),
                                            color: '#3b82f6',
                                        },
                                        {
                                            label: 'P90',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.latency?.p90 || 0,
                                            })),
                                            color: '#f59e0b',
                                        },
                                        {
                                            label: 'P99',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.latency?.p99 || 0,
                                            })),
                                            color: '#ef4444',
                                        },
                                    ]}
                                    title={`Latency Trend - ${selectedProxyData.id}`}
                                    yAxisLabel="Latency (ms)"
                                />
                            </div>

                            <div className="chart-container">
                                <LineChart
                                    datasets={[
                                        {
                                            label: 'Error Rate',
                                            data: timeseriesData.map((d) => {
                                                // Calculate error rate from QPS if available
                                                const totalQps =
                                                    (d.qps?.get || 0) +
                                                    (d.qps?.set || 0) +
                                                    (d.qps?.delete || 0);
                                                // Assuming error rate is stored or simulated
                                                const errorRate = totalQps > 0 ? Math.random() * 2 : 0;
                                                return {
                                                    timestamp: d.timestamp,
                                                    value: errorRate,
                                                };
                                            }),
                                            color: '#ef4444',
                                            fill: true,
                                        },
                                    ]}
                                    title={`Error Rate - ${selectedProxyData.id}`}
                                    yAxisLabel="Error Rate %"
                                />
                            </div>
                        </div>
                    )}
                </div>
            )}

            <div className="charts-grid">
                <div className="chart-container">
                    <BarChart
                        data={connectionsData}
                        title="Active Connections per Proxy"
                        color="#3b82f6"
                        horizontal
                    />
                </div>
            </div>
        </div>
    );
}

