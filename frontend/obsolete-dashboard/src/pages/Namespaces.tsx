// Namespaces page

import { useEffect } from 'react';
import { useMetricsStore } from '../stores/metricsStore';
import { fetchNamespaces } from '../api/client';
import { BarChart } from '../components/charts/BarChart';

export function Namespaces() {
    const { namespaces, setNamespaces, setError, setLoading } = useMetricsStore();

    useEffect(() => {
        const loadData = async () => {
            try {
                setLoading(true);
                const data = await fetchNamespaces();
                setNamespaces(data.namespaces);
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
    }, [setNamespaces, setError, setLoading]);

    const memoryData = namespaces.map((ns) => ({
        name: ns.name,
        value: ns.memory_used,
    }));

    const qpsData = namespaces.map((ns) => ({
        name: ns.name,
        value: ns.qps.get + ns.qps.set + ns.qps.delete,
    }));

    return (
        <div className="page">
            <h1>Business Namespaces</h1>

            <table className="data-table">
                <thead>
                    <tr>
                        <th>Namespace</th>
                        <th>Description</th>
                        <th>API Key</th>
                        <th>Keys</th>
                        <th>Memory</th>
                        <th>QPS</th>
                        <th>Hit Ratio</th>
                    </tr>
                </thead>
                <tbody>
                    {namespaces.map((ns) => (
                        <tr key={ns.name}>
                            <td>
                                <strong>{ns.name}</strong>
                            </td>
                            <td>{ns.description}</td>
                            <td>
                                <code>{ns.api_key_masked}</code>
                            </td>
                            <td>
                                {ns.key_count.toLocaleString()} / {ns.max_keys.toLocaleString()}
                            </td>
                            <td>
                                {ns.memory_used} / {ns.memory_limit} MB (
                                {((ns.memory_used / ns.memory_limit) * 100).toFixed(1)}%)
                            </td>
                            <td>
                                {ns.qps.get + ns.qps.set + ns.qps.delete} (G:{ns.qps.get} S:
                                {ns.qps.set} D:{ns.qps.delete})
                            </td>
                            <td>{(ns.hit_ratio * 100).toFixed(2)}%</td>
                        </tr>
                    ))}
                </tbody>
            </table>

            <div className="namespace-details">
                {namespaces.map((ns) => (
                    <div key={ns.name} className="namespace-card">
                        <h3>{ns.name}</h3>
                        <p className="description">{ns.description}</p>
                        <div className="namespace-info">
                            <div className="info-item">
                                <span className="label">Rate Limit:</span>
                                <span className="value">{ns.rate_limit_qps} QPS</span>
                            </div>
                            <div className="info-item">
                                <span className="label">Default TTL:</span>
                                <span className="value">{ns.default_ttl}s</span>
                            </div>
                            <div className="info-item">
                                <span className="label">Memory Limit:</span>
                                <span className="value">{ns.memory_limit} MB</span>
                            </div>
                            <div className="info-item">
                                <span className="label">Max Keys:</span>
                                <span className="value">{ns.max_keys.toLocaleString()}</span>
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            <div className="charts-grid">
                <div className="chart-container">
                    <BarChart data={memoryData} title="Memory Usage by Namespace (MB)" color="#10b981" />
                </div>
                <div className="chart-container">
                    <BarChart data={qpsData} title="QPS by Namespace" color="#f59e0b" />
                </div>
            </div>
        </div>
    );
}

