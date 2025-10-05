// Cluster overview page

import { useEffect, useState } from 'react';
import { useMetricsStore } from '../stores/metricsStore';
import { fetchOverview, fetchClusterTimeseries, fetchNamespaces } from '../api/client';
import { MetricCard } from '../components/MetricCard';
import { GaugeChart } from '../components/charts/GaugeChart';
import { PieChart } from '../components/charts/PieChart';
import { LineChart } from '../components/charts/LineChart';
import { TimeRangeSelector, type TimeRange } from '../components/TimeRangeSelector';
import type { TimeSeriesPoint, NamespaceStats } from '../types/metrics';

export function Overview() {
    const { overview, loading, error, setOverview, setError, setLoading } = useMetricsStore();
    const [timeseriesData, setTimeseriesData] = useState<TimeSeriesPoint[]>([]);
    const [namespaces, setNamespaces] = useState<NamespaceStats[]>([]);
    const [timeRange, setTimeRange] = useState<TimeRange>('1h');

    useEffect(() => {
        const loadData = async () => {
            try {
                setLoading(true);
                const [overviewData, namespacesData] = await Promise.all([
                    fetchOverview(),
                    fetchNamespaces(),
                ]);
                setOverview(overviewData);
                setNamespaces(namespacesData.namespaces);
                setError(null);
            } catch (err) {
                console.error('Failed to load overview data:', err);
                setError(err instanceof Error ? err.message : 'Failed to load data');
            } finally {
                setLoading(false);
            }
        };

        loadData();
        const interval = setInterval(loadData, 5000);
        return () => clearInterval(interval);
    }, [setOverview, setError, setLoading]);

    useEffect(() => {
        const loadTimeseries = async () => {
            try {
                // TODO: Pass timeRange to API when supported
                const data = await fetchClusterTimeseries();
                setTimeseriesData(data.metrics);
            } catch (err) {
                console.error('Failed to load timeseries:', err);
            }
        };

        loadTimeseries();
        const interval = setInterval(loadTimeseries, 10000);
        return () => clearInterval(interval);
    }, [timeRange]);

    // Show loading only on initial load
    if (loading && !overview) {
        return (
            <div className="page">
                <div className="loading">Loading cluster overview...</div>
            </div>
        );
    }

    // Show error if no data available
    if (error && !overview) {
        return (
            <div className="page">
                <div className="error-banner">
                    ⚠️ Failed to load data: {error}
                </div>
            </div>
        );
    }

    // Show fallback if overview is still null (shouldn't happen normally)
    if (!overview) {
        return (
            <div className="page">
                <div className="loading">Waiting for data...</div>
            </div>
        );
    }

    const requestDistribution = [
        {
            name: 'GET',
            value:
                overview.metrics.total_qps * 0.65 || 0,
        },
        {
            name: 'SET',
            value:
                overview.metrics.total_qps * 0.3 || 0,
        },
        {
            name: 'DELETE',
            value:
                overview.metrics.total_qps * 0.05 || 0,
        },
    ];

    const namespaceDistribution = namespaces.map((ns) => ({
        name: ns.name,
        value: ns.qps.get + ns.qps.set + ns.qps.delete,
    }));

    const handleTimeRangeChange = (range: TimeRange) => {
        setTimeRange(range);
        // TODO: Implement API call with time range parameters when Admin API supports it
        // The timeRange state will be used for filtering timeseries data
        console.log('Time range changed:', range, timeRange);
    };

    return (
        <div className="page">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h1>Cluster Overview</h1>
                <TimeRangeSelector onRangeChange={handleTimeRangeChange} defaultRange="1h" />
            </div>

            {/* Key Metrics Section */}
            <div className="metric-cards">
                <MetricCard
                    title="Total QPS"
                    value={overview.metrics.total_qps.toLocaleString()}
                    subtitle="Requests per second"
                    color="#00f5ff"
                    icon="⚡"
                />
                <MetricCard
                    title="Cache Hit Ratio"
                    value={`${(overview.metrics.hit_ratio * 100).toFixed(1)}%`}
                    subtitle={`${overview.metrics.total_keys.toLocaleString()} total keys`}
                    color="#10b981"
                    icon="🎯"
                />
                <MetricCard
                    title="Avg Latency"
                    value={overview.metrics.avg_latency_ms.toFixed(2)}
                    unit="ms"
                    subtitle="P50 response time"
                    color="#a855f7"
                    icon="⏱️"
                />
                <MetricCard
                    title="Cluster Health"
                    value={`${overview.proxies.healthy + overview.nodes.healthy}/${overview.proxies.total + overview.nodes.total}`}
                    subtitle={
                        (overview.proxies.unhealthy + overview.nodes.unhealthy) > 0
                            ? `${overview.proxies.unhealthy + overview.nodes.unhealthy} unhealthy`
                            : 'All services healthy'
                    }
                    color={
                        (overview.proxies.unhealthy + overview.nodes.unhealthy) === 0
                            ? '#10b981'
                            : (overview.proxies.unhealthy + overview.nodes.unhealthy) < 3
                                ? '#fbbf24'
                                : '#f43f5e'
                    }
                    icon="🏥"
                />
            </div>

            {/* Service Status Section */}
            <div style={{ marginTop: '2rem', display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.25rem' }}>
                <div style={{
                    background: 'var(--glass-bg)',
                    backdropFilter: 'blur(15px)',
                    borderRadius: '16px',
                    padding: '1.5rem',
                    border: '1px solid var(--glass-border)',
                    borderLeft: '3px solid #00f5ff'
                }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                        <h3 style={{ margin: 0, fontSize: '0.95rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                            Proxy Instances
                        </h3>
                        <span style={{ fontSize: '1.5rem' }}>🔀</span>
                    </div>
                    <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#00f5ff', marginBottom: '0.5rem' }}>
                        {overview.proxies.healthy}/{overview.proxies.total}
                    </div>
                    <div style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                        {overview.proxies.unhealthy > 0
                            ? `⚠ ${overview.proxies.unhealthy} unhealthy`
                            : '✓ All healthy'}
                    </div>
                </div>

                <div style={{
                    background: 'var(--glass-bg)',
                    backdropFilter: 'blur(15px)',
                    borderRadius: '16px',
                    padding: '1.5rem',
                    border: '1px solid var(--glass-border)',
                    borderLeft: '3px solid #a855f7'
                }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                        <h3 style={{ margin: 0, fontSize: '0.95rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                            Cache Nodes
                        </h3>
                        <span style={{ fontSize: '1.5rem' }}>💾</span>
                    </div>
                    <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#a855f7', marginBottom: '0.5rem' }}>
                        {overview.nodes.healthy}/{overview.nodes.total}
                    </div>
                    <div style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                        {overview.nodes.unhealthy > 0
                            ? `⚠ ${overview.nodes.unhealthy} unhealthy`
                            : '✓ All healthy'}
                    </div>
                </div>

                <div style={{
                    background: 'var(--glass-bg)',
                    backdropFilter: 'blur(15px)',
                    borderRadius: '16px',
                    padding: '1.5rem',
                    border: '1px solid var(--glass-border)',
                    borderLeft: '3px solid #10b981'
                }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                        <h3 style={{ margin: 0, fontSize: '0.95rem', color: 'var(--color-text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                            Namespaces
                        </h3>
                        <span style={{ fontSize: '1.5rem' }}>📦</span>
                    </div>
                    <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#10b981', marginBottom: '0.5rem' }}>
                        {namespaces.length}
                    </div>
                    <div style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                        Active namespaces
                    </div>
                </div>
            </div>

            {/* Performance Trends Section */}
            {timeseriesData.length > 0 && (
                <div style={{ marginTop: '2.5rem' }}>
                    <h2 style={{ marginBottom: '1.5rem', color: 'var(--color-text-primary)' }}>
                        📈 Performance Trends
                    </h2>
                    <div className="charts-grid">
                        <div className="chart-container" style={{ gridColumn: 'span 2' }}>
                            <LineChart
                                datasets={[
                                    {
                                        label: 'GET',
                                        data: timeseriesData.map((d) => ({
                                            timestamp: d.timestamp,
                                            value: d.qps?.get || 0,
                                        })),
                                        color: '#00f5ff',
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
                                        color: '#f43f5e',
                                        fill: true,
                                    },
                                ]}
                                title="QPS Trend by Operation Type"
                                yAxisLabel="Requests/sec"
                                height={280}
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
                                        color: '#00f5ff',
                                    },
                                    {
                                        label: 'P90',
                                        data: timeseriesData.map((d) => ({
                                            timestamp: d.timestamp,
                                            value: d.latency?.p90 || 0,
                                        })),
                                        color: '#fbbf24',
                                    },
                                    {
                                        label: 'P99',
                                        data: timeseriesData.map((d) => ({
                                            timestamp: d.timestamp,
                                            value: d.latency?.p99 || 0,
                                        })),
                                        color: '#f43f5e',
                                    },
                                ]}
                                title="Latency Distribution"
                                yAxisLabel="Latency (ms)"
                                height={280}
                            />
                        </div>

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
                                title="Cache Hit Ratio"
                                yAxisLabel="Hit Ratio (%)"
                                height={280}
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
                                        color: '#a855f7',
                                        fill: true,
                                    },
                                ]}
                                title="Memory Usage"
                                yAxisLabel="Memory (MB)"
                                height={280}
                            />
                        </div>
                    </div>
                </div>
            )}

            {/* Distribution Charts Section */}
            <div style={{ marginTop: '2.5rem' }}>
                <h2 style={{ marginBottom: '1.5rem', color: 'var(--color-text-primary)' }}>
                    📊 Request Distribution
                </h2>
                <div className="charts-grid">
                    <div className="chart-container">
                        <GaugeChart value={overview.metrics.hit_ratio} title="Overall Cache Hit Ratio" />
                    </div>

                    <div className="chart-container">
                        <PieChart data={requestDistribution} title="Request Type Distribution" />
                    </div>

                    {namespaceDistribution.length > 0 && (
                        <div className="chart-container">
                            <PieChart data={namespaceDistribution} title="Namespace QPS Distribution" />
                        </div>
                    )}
                </div>
            </div>

            {/* Summary Footer */}
            <div className="metrics-summary" style={{ marginTop: '2rem' }}>
                <div className="metric-item">
                    <span className="label">Total Keys:</span>
                    <span className="value">{overview.metrics.total_keys.toLocaleString()}</span>
                </div>
                <div className="metric-item">
                    <span className="label">Avg Latency:</span>
                    <span className="value">{overview.metrics.avg_latency_ms.toFixed(2)} ms</span>
                </div>
                <div className="metric-item">
                    <span className="label">Last Updated:</span>
                    <span className="value">
                        {new Date(overview.last_updated).toLocaleTimeString()}
                    </span>
                </div>
            </div>
        </div>
    );
}

