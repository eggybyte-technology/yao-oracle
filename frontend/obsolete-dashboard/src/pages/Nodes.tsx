// Cache node instances page

import { useEffect, useState } from 'react';
import { useMetricsStore } from '../stores/metricsStore';
import { fetchNodes, fetchNodeTimeseries } from '../api/client';
import { StatusBadge } from '../components/StatusBadge';
import { BarChart } from '../components/charts/BarChart';
import { LineChart } from '../components/charts/LineChart';
import type { TimeSeriesPoint } from '../types/metrics';

export function Nodes() {
    const { nodes, setNodes, setError, setLoading } = useMetricsStore();
    const [selectedNode, setSelectedNode] = useState<string | null>(null);
    const [timeseriesData, setTimeseriesData] = useState<TimeSeriesPoint[]>([]);

    useEffect(() => {
        const loadData = async () => {
            try {
                setLoading(true);
                const data = await fetchNodes();
                setNodes(data.nodes);
                if (data.nodes.length > 0 && !selectedNode) {
                    setSelectedNode(data.nodes[0].id);
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
    }, [setNodes, setError, setLoading, selectedNode]);

    useEffect(() => {
        if (!selectedNode) return;

        const loadTimeseries = async () => {
            try {
                const data = await fetchNodeTimeseries(selectedNode);
                setTimeseriesData(data.metrics);
            } catch (err) {
                console.error('Failed to load timeseries:', err);
            }
        };

        loadTimeseries();
        const interval = setInterval(loadTimeseries, 10000);
        return () => clearInterval(interval);
    }, [selectedNode]);

    const selectedNodeData = nodes.find((n) => n.id === selectedNode);

    const memoryData = nodes.map((n) => ({
        name: n.id,
        value: (n.memory.used / n.memory.max) * 100,
    }));

    const calculateHitRatio = (node: typeof nodes[0]) => {
        const total = node.hit_count + node.miss_count;
        return total > 0 ? (node.hit_count / total) * 100 : 0;
    };

    return (
        <div className="page">
            <h1>Cache Nodes</h1>

            <table className="data-table">
                <thead>
                    <tr>
                        <th>Instance</th>
                        <th>IP Address</th>
                        <th>Keys</th>
                        <th>Memory</th>
                        <th>Hit Ratio</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {nodes.map((node) => (
                        <tr
                            key={node.id}
                            onClick={() => setSelectedNode(node.id)}
                            className={selectedNode === node.id ? 'selected' : ''}
                            style={{ cursor: 'pointer' }}
                        >
                            <td>{node.id}</td>
                            <td>{node.ip}</td>
                            <td>{node.key_count.toLocaleString()}</td>
                            <td>
                                {node.memory.used} / {node.memory.max} MB (
                                {((node.memory.used / node.memory.max) * 100).toFixed(1)}%)
                            </td>
                            <td>{calculateHitRatio(node).toFixed(2)}%</td>
                            <td>
                                <StatusBadge status={node.status} />
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {selectedNodeData && (
                <div className="details-section">
                    <h2>Details: {selectedNodeData.id}</h2>
                    <div className="detail-info">
                        <div className="info-item">
                            <span className="label">IP:</span>
                            <span className="value">{selectedNodeData.ip}</span>
                        </div>
                        <div className="info-item">
                            <span className="label">Uptime:</span>
                            <span className="value">
                                {Math.floor(selectedNodeData.uptime / 3600)}h
                            </span>
                        </div>
                        <div className="info-item">
                            <span className="label">Hit Count:</span>
                            <span className="value">
                                {selectedNodeData.hit_count.toLocaleString()}
                            </span>
                        </div>
                        <div className="info-item">
                            <span className="label">Miss Count:</span>
                            <span className="value">
                                {selectedNodeData.miss_count.toLocaleString()}
                            </span>
                        </div>
                        <div className="info-item">
                            <span className="label">Hot Keys:</span>
                            <span className="value">{selectedNodeData.hot_keys.length}</span>
                        </div>
                    </div>

                    {timeseriesData.length > 0 && (
                        <div className="charts-grid" style={{ marginTop: '1rem' }}>
                            <div className="chart-container">
                                <LineChart
                                    datasets={[
                                        {
                                            label: 'Hit Ratio',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: (d.hit_ratio || 0) * 100,
                                            })),
                                            color: '#10b981',
                                            fill: true,
                                        },
                                    ]}
                                    title={`Hit Ratio Trend - ${selectedNodeData.id}`}
                                    yAxisLabel="Hit Ratio %"
                                />
                            </div>

                            <div className="chart-container">
                                <LineChart
                                    datasets={[
                                        {
                                            label: 'Memory Used',
                                            data: timeseriesData.map((d) => ({
                                                timestamp: d.timestamp,
                                                value: d.memory?.used || 0,
                                            })),
                                            color: '#8b5cf6',
                                            fill: true,
                                        },
                                    ]}
                                    title={`Memory Usage - ${selectedNodeData.id}`}
                                    yAxisLabel="Memory (MB)"
                                />
                            </div>
                        </div>
                    )}

                    <div style={{ marginTop: '2.5rem' }}>
                        <h2 style={{ marginBottom: '1.5rem', color: 'var(--color-text-primary)' }}>
                            📊 Node Statistics
                        </h2>
                        <div className="charts-grid">
                            <div className="chart-container" style={{
                                background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.08), rgba(16, 185, 129, 0.02))',
                                borderLeft: '3px solid #10b981'
                            }}>
                                <div style={{ padding: '2rem', textAlign: 'center' }}>
                                    <div style={{
                                        fontSize: '3.5rem',
                                        fontWeight: 'bold',
                                        marginBottom: '0.75rem',
                                        background: 'linear-gradient(135deg, #10b981, #00f5ff)',
                                        WebkitBackgroundClip: 'text',
                                        WebkitTextFillColor: 'transparent',
                                        backgroundClip: 'text',
                                    }}>
                                        {calculateHitRatio(selectedNodeData).toFixed(1)}%
                                    </div>
                                    <div style={{
                                        color: 'var(--color-text-secondary)',
                                        fontSize: '0.875rem',
                                        textTransform: 'uppercase',
                                        letterSpacing: '0.1em',
                                        marginBottom: '2rem',
                                        fontWeight: 600
                                    }}>
                                        Cache Hit Ratio
                                    </div>
                                    <div style={{ marginTop: '1.5rem', display: 'flex', justifyContent: 'space-around' }}>
                                        <div>
                                            <div style={{
                                                color: '#10b981',
                                                fontSize: '1.75rem',
                                                fontWeight: 700,
                                                fontFamily: 'JetBrains Mono, monospace'
                                            }}>
                                                {selectedNodeData.hit_count.toLocaleString()}
                                            </div>
                                            <div style={{
                                                color: 'var(--color-text-muted)',
                                                fontSize: '0.75rem',
                                                marginTop: '0.5rem',
                                                textTransform: 'uppercase',
                                                letterSpacing: '0.05em'
                                            }}>
                                                Hits
                                            </div>
                                        </div>
                                        <div>
                                            <div style={{
                                                color: '#f43f5e',
                                                fontSize: '1.75rem',
                                                fontWeight: 700,
                                                fontFamily: 'JetBrains Mono, monospace'
                                            }}>
                                                {selectedNodeData.miss_count.toLocaleString()}
                                            </div>
                                            <div style={{
                                                color: 'var(--color-text-muted)',
                                                fontSize: '0.75rem',
                                                marginTop: '0.5rem',
                                                textTransform: 'uppercase',
                                                letterSpacing: '0.05em'
                                            }}>
                                                Misses
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className="chart-container" style={{
                                background: 'linear-gradient(135deg, rgba(168, 85, 247, 0.08), rgba(168, 85, 247, 0.02))',
                                borderLeft: '3px solid #a855f7'
                            }}>
                                <div style={{ padding: '2rem', textAlign: 'center' }}>
                                    <div style={{
                                        fontSize: '3.5rem',
                                        fontWeight: 'bold',
                                        marginBottom: '0.75rem',
                                        background: 'linear-gradient(135deg, #a855f7, #00f5ff)',
                                        WebkitBackgroundClip: 'text',
                                        WebkitTextFillColor: 'transparent',
                                        backgroundClip: 'text',
                                    }}>
                                        {((selectedNodeData.memory.used / selectedNodeData.memory.max) * 100).toFixed(1)}%
                                    </div>
                                    <div style={{
                                        color: 'var(--color-text-secondary)',
                                        fontSize: '0.875rem',
                                        textTransform: 'uppercase',
                                        letterSpacing: '0.1em',
                                        marginBottom: '2rem',
                                        fontWeight: 600
                                    }}>
                                        Memory Usage
                                    </div>
                                    <div style={{ marginTop: '1.5rem' }}>
                                        <div style={{
                                            color: '#a855f7',
                                            fontSize: '1.4rem',
                                            fontWeight: 700,
                                            fontFamily: 'JetBrains Mono, monospace',
                                            marginBottom: '0.75rem'
                                        }}>
                                            {selectedNodeData.memory.used} / {selectedNodeData.memory.max} MB
                                        </div>
                                        <div style={{
                                            color: 'var(--color-text-muted)',
                                            fontSize: '0.875rem',
                                            marginTop: '0.75rem'
                                        }}>
                                            {selectedNodeData.key_count.toLocaleString()} Keys Stored
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {selectedNodeData.hot_keys.length > 0 && (
                        <div style={{ marginTop: '2.5rem' }}>
                            <h2 style={{ marginBottom: '1.5rem', color: 'var(--color-text-primary)' }}>
                                🔥 Hot Keys (Top Accessed)
                            </h2>
                            <div className="charts-grid">
                                <div className="chart-container" style={{ gridColumn: 'span 2' }}>
                                    <BarChart
                                        data={selectedNodeData.hot_keys.map((hk) => ({
                                            name: hk.key,
                                            value: hk.frequency,
                                        }))}
                                        title={`Hot Keys Frequency - ${selectedNodeData.id}`}
                                        color="#f59e0b"
                                        horizontal={false}
                                    />
                                </div>
                            </div>
                            <table className="data-table" style={{ marginTop: '1rem' }}>
                                <thead>
                                    <tr>
                                        <th>Rank</th>
                                        <th>Key</th>
                                        <th>Access Frequency</th>
                                        <th>% of Total</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {selectedNodeData.hot_keys.map((hk, idx) => {
                                        const totalFreq = selectedNodeData.hot_keys.reduce(
                                            (sum, k) => sum + k.frequency,
                                            0
                                        );
                                        const percentage = ((hk.frequency / totalFreq) * 100).toFixed(
                                            1
                                        );
                                        return (
                                            <tr key={idx}>
                                                <td>
                                                    <span
                                                        style={{
                                                            display: 'inline-flex',
                                                            alignItems: 'center',
                                                            justifyContent: 'center',
                                                            width: '28px',
                                                            height: '28px',
                                                            borderRadius: '50%',
                                                            background:
                                                                idx === 0
                                                                    ? 'linear-gradient(135deg, #fbbf24, #f59e0b)'
                                                                    : idx === 1
                                                                        ? 'linear-gradient(135deg, #a855f7, #8b5cf6)'
                                                                        : idx === 2
                                                                            ? 'linear-gradient(135deg, #00f5ff, #06b6d4)'
                                                                            : 'rgba(255,255,255,0.1)',
                                                            color: '#fff',
                                                            fontSize: '0.8125rem',
                                                            fontWeight: 700,
                                                            boxShadow: idx < 3 ? '0 2px 8px rgba(0,0,0,0.3)' : 'none',
                                                        }}
                                                    >
                                                        {idx + 1}
                                                    </span>
                                                </td>
                                                <td>
                                                    <code
                                                        style={{
                                                            backgroundColor: 'rgba(0, 245, 255, 0.1)',
                                                            padding: '0.5rem 1rem',
                                                            borderRadius: '8px',
                                                            fontFamily: 'JetBrains Mono, monospace',
                                                            fontSize: '0.875rem',
                                                            color: '#00f5ff',
                                                            border: '1px solid rgba(0, 245, 255, 0.25)',
                                                        }}
                                                    >
                                                        {hk.key}
                                                    </code>
                                                </td>
                                                <td>
                                                    <span style={{
                                                        fontFamily: 'JetBrains Mono, monospace',
                                                        fontWeight: 600,
                                                        color: 'var(--color-text-primary)'
                                                    }}>
                                                        {hk.frequency.toLocaleString()}
                                                    </span>
                                                </td>
                                                <td>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                        <div
                                                            style={{
                                                                flex: '0 0 60px',
                                                                height: '8px',
                                                                backgroundColor: 'rgba(255,255,255,0.1)',
                                                                borderRadius: '4px',
                                                                overflow: 'hidden',
                                                            }}
                                                        >
                                                            <div
                                                                style={{
                                                                    width: `${percentage}%`,
                                                                    height: '100%',
                                                                    background: 'linear-gradient(90deg, #fbbf24, #f59e0b)',
                                                                    borderRadius: '4px',
                                                                }}
                                                            />
                                                        </div>
                                                        <span>{percentage}%</span>
                                                    </div>
                                                </td>
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            )}

            <div className="charts-grid">
                <div className="chart-container">
                    <BarChart
                        data={memoryData}
                        title="Memory Usage per Node (%)"
                        color="#8b5cf6"
                        horizontal
                    />
                </div>
            </div>
        </div>
    );
}

