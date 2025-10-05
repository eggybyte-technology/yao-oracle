// Cache query page with filtering by namespace and key

import { useState, useEffect, Fragment } from 'react';
import { fetchCacheEntries } from '../api/client';
import type { CacheEntry } from '../types/metrics';

export function CacheQuery() {
    const [entries, setEntries] = useState<CacheEntry[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(20);
    const [expandedKey, setExpandedKey] = useState<string | null>(null);

    // Filter states
    const [namespaceFilter, setNamespaceFilter] = useState('');
    const [keyFilter, setKeyFilter] = useState('');
    const [matchMode, setMatchMode] = useState<'exact' | 'prefix' | 'contains'>('contains');

    // Applied filters (triggered on search)
    const [appliedNamespace, setAppliedNamespace] = useState('');
    const [appliedKey, setAppliedKey] = useState('');

    // Sort state
    const [sortField, setSortField] = useState<'key' | 'ttl' | 'size' | 'access_count' | 'accessed_at'>('accessed_at');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

    useEffect(() => {
        const loadData = async () => {
            try {
                setLoading(true);
                const params: Parameters<typeof fetchCacheEntries>[0] = {
                    page,
                    page_size: pageSize,
                };
                if (appliedNamespace) params.namespace = appliedNamespace;
                if (appliedKey) params.key = appliedKey;

                const data = await fetchCacheEntries(params);
                setEntries(data.entries || []);
                setTotal(data.total);
                setError(null);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load cache data');
                setEntries([]);
            } finally {
                setLoading(false);
            }
        };

        loadData();
        const interval = setInterval(loadData, 10000); // Refresh every 10 seconds
        return () => clearInterval(interval);
    }, [page, pageSize, appliedNamespace, appliedKey]);

    const handleSearch = () => {
        setAppliedNamespace(namespaceFilter);
        setAppliedKey(keyFilter);
        setPage(1); // Reset to first page
    };

    const handleReset = () => {
        setNamespaceFilter('');
        setKeyFilter('');
        setAppliedNamespace('');
        setAppliedKey('');
        setPage(1);
    };

    const formatSize = (bytes: number): string => {
        if (bytes < 1024) return `${bytes} B`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
        return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
    };

    const formatTTL = (seconds: number): string => {
        if (seconds < 0) return 'Expired';
        if (seconds < 60) return `${seconds}s`;
        if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
        if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`;
        return `${Math.floor(seconds / 86400)}d`;
    };

    const formatTimestamp = (timestamp: string): string => {
        const date = new Date(timestamp);
        return date.toLocaleString();
    };

    const handleSort = (field: typeof sortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('desc');
        }
    };

    const getSortedEntries = () => {
        const sorted = [...entries].sort((a, b) => {
            let comparison = 0;
            switch (sortField) {
                case 'key':
                    comparison = a.key.localeCompare(b.key);
                    break;
                case 'ttl':
                    comparison = a.ttl - b.ttl;
                    break;
                case 'size':
                    comparison = a.size - b.size;
                    break;
                case 'access_count':
                    comparison = a.access_count - b.access_count;
                    break;
                case 'accessed_at':
                    comparison = new Date(a.accessed_at).getTime() - new Date(b.accessed_at).getTime();
                    break;
            }
            return sortDirection === 'asc' ? comparison : -comparison;
        });
        return sorted;
    };

    const exportToCSV = () => {
        const headers = ['Namespace', 'Key', 'Value', 'Size (bytes)', 'TTL (s)', 'Access Count', 'Created At', 'Last Access'];
        const rows = entries.map((entry) => [
            entry.namespace,
            entry.key,
            entry.value.replace(/"/g, '""'), // Escape quotes
            entry.size.toString(),
            entry.ttl.toString(),
            entry.access_count.toString(),
            entry.created_at,
            entry.accessed_at,
        ]);

        const csvContent = [
            headers.join(','),
            ...rows.map((row) => row.map((cell) => `"${cell}"`).join(',')),
        ].join('\n');

        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const link = document.createElement('a');
        const url = URL.createObjectURL(blob);
        link.setAttribute('href', url);
        link.setAttribute('download', `cache_entries_${new Date().toISOString()}.csv`);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    const totalPages = Math.ceil(total / pageSize);
    const sortedEntries = getSortedEntries();

    return (
        <div className="page">
            <h1>Cache Query</h1>

            {/* Filter Section */}
            <div className="filter-section">
                <div className="filter-inputs">
                    <div className="filter-group">
                        <label htmlFor="namespace-filter">Namespace</label>
                        <input
                            id="namespace-filter"
                            type="text"
                            placeholder="Filter by namespace..."
                            value={namespaceFilter}
                            onChange={(e) => setNamespaceFilter(e.target.value)}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') handleSearch();
                            }}
                            className="filter-input"
                        />
                    </div>

                    <div className="filter-group">
                        <label htmlFor="key-filter">Key</label>
                        <input
                            id="key-filter"
                            type="text"
                            placeholder="Filter by key pattern..."
                            value={keyFilter}
                            onChange={(e) => setKeyFilter(e.target.value)}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') handleSearch();
                            }}
                            className="filter-input"
                        />
                    </div>

                    <div className="filter-group">
                        <label htmlFor="match-mode">Match Mode</label>
                        <select
                            id="match-mode"
                            value={matchMode}
                            onChange={(e) => setMatchMode(e.target.value as typeof matchMode)}
                            className="filter-input"
                            style={{ width: '140px' }}
                        >
                            <option value="exact">Exact Match</option>
                            <option value="prefix">Prefix Match</option>
                            <option value="contains">Contains</option>
                        </select>
                    </div>

                    <div className="filter-actions">
                        <button onClick={handleSearch} className="btn btn-primary">
                            🔍 Search
                        </button>
                        <button onClick={handleReset} className="btn btn-secondary">
                            🔄 Reset
                        </button>
                        {entries.length > 0 && (
                            <button onClick={exportToCSV} className="btn btn-secondary">
                                📥 Export CSV
                            </button>
                        )}
                    </div>
                </div>

                <div className="filter-info">
                    <span>
                        Showing {entries.length} of {total} entries
                        {(appliedNamespace || appliedKey) && (
                            <span className="filter-applied">
                                {' '}
                                • Filtered by{' '}
                                {appliedNamespace && (
                                    <code>namespace: {appliedNamespace}</code>
                                )}
                                {appliedNamespace && appliedKey && ', '}
                                {appliedKey && <code>key: {appliedKey}</code>}
                            </span>
                        )}
                    </span>
                </div>
            </div>

            {/* Loading State */}
            {loading && <div className="loading">Loading cache data...</div>}

            {/* Error State */}
            {error && (
                <div className="error-message">
                    ⚠️ Error: {error}
                </div>
            )}

            {/* Data Table with expandable rows */}
            {!loading && !error && entries.length > 0 && (
                <>
                    <div className="table-container">
                        <table className="data-table">
                            <thead>
                                <tr>
                                    <th style={{ width: '36px' }}></th>
                                    <th>Namespace</th>
                                    <th
                                        onClick={() => handleSort('key')}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        Key {sortField === 'key' && (sortDirection === 'asc' ? '▲' : '▼')}
                                    </th>
                                    <th>Value Preview</th>
                                    <th
                                        onClick={() => handleSort('size')}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        Size {sortField === 'size' && (sortDirection === 'asc' ? '▲' : '▼')}
                                    </th>
                                    <th
                                        onClick={() => handleSort('ttl')}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        TTL {sortField === 'ttl' && (sortDirection === 'asc' ? '▲' : '▼')}
                                    </th>
                                    <th
                                        onClick={() => handleSort('access_count')}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        Access Count {sortField === 'access_count' && (sortDirection === 'asc' ? '▲' : '▼')}
                                    </th>
                                    <th>Created At</th>
                                    <th
                                        onClick={() => handleSort('accessed_at')}
                                        style={{ cursor: 'pointer', userSelect: 'none' }}
                                    >
                                        Last Access {sortField === 'accessed_at' && (sortDirection === 'asc' ? '▲' : '▼')}
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {sortedEntries.map((entry, idx) => {
                                    const rowId = `${entry.namespace}-${entry.key}-${idx}`;
                                    const isExpanded = expandedKey === rowId;
                                    return (
                                        <Fragment key={rowId}>
                                            <tr
                                                className={isExpanded ? 'selected' : ''}
                                                onClick={() => setExpandedKey(isExpanded ? null : rowId)}
                                                style={{ cursor: 'pointer' }}
                                            >
                                                <td>
                                                    <span className="expander">{isExpanded ? '▾' : '▸'}</span>
                                                </td>
                                                <td>
                                                    <span className="badge badge-namespace">
                                                        {entry.namespace}
                                                    </span>
                                                </td>
                                                <td>
                                                    <code className="key-code">{entry.key}</code>
                                                </td>
                                                <td>
                                                    <div className="value-preview">
                                                        {entry.value.length > 100
                                                            ? `${entry.value.substring(0, 100)}...`
                                                            : entry.value}
                                                    </div>
                                                </td>
                                                <td>{formatSize(entry.size)}</td>
                                                <td>
                                                    <span
                                                        className={`ttl-badge ${entry.ttl < 60 ? 'ttl-warning' : ''}`}
                                                    >
                                                        {formatTTL(entry.ttl)}
                                                    </span>
                                                </td>
                                                <td className="text-center">{entry.access_count}</td>
                                                <td className="text-small">
                                                    {formatTimestamp(entry.created_at)}
                                                </td>
                                                <td className="text-small">
                                                    {formatTimestamp(entry.accessed_at)}
                                                </td>
                                            </tr>
                                            {isExpanded && (
                                                <tr className="details-row">
                                                    <td colSpan={9}>
                                                        <div className="details-section">
                                                            <h3>Entry Details</h3>
                                                            <div className="detail-info">
                                                                <div className="info-item">
                                                                    <span className="label">Namespace</span>
                                                                    <span className="value">{entry.namespace}</span>
                                                                </div>
                                                                <div className="info-item">
                                                                    <span className="label">Key</span>
                                                                    <span className="value">{entry.key}</span>
                                                                </div>
                                                                <div className="info-item">
                                                                    <span className="label">Size</span>
                                                                    <span className="value">{formatSize(entry.size)}</span>
                                                                </div>
                                                                <div className="info-item">
                                                                    <span className="label">TTL</span>
                                                                    <span className="value">{formatTTL(entry.ttl)}</span>
                                                                </div>
                                                                <div className="info-item">
                                                                    <span className="label">Access Count</span>
                                                                    <span className="value">{entry.access_count}</span>
                                                                </div>
                                                                <div className="info-item">
                                                                    <span className="label">Created At</span>
                                                                    <span className="value">{formatTimestamp(entry.created_at)}</span>
                                                                </div>
                                                                <div className="info-item">
                                                                    <span className="label">Last Access</span>
                                                                    <span className="value">{formatTimestamp(entry.accessed_at)}</span>
                                                                </div>
                                                            </div>

                                                            <h3>Value</h3>
                                                            <pre className="value-block"><code>{entry.value}</code></pre>
                                                        </div>
                                                    </td>
                                                </tr>
                                            )}
                                        </Fragment>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <div className="pagination">
                            <button
                                onClick={() => setPage((p) => Math.max(1, p - 1))}
                                disabled={page === 1}
                                className="btn btn-secondary"
                            >
                                ← Previous
                            </button>
                            <span className="page-info">
                                Page {page} of {totalPages}
                            </span>
                            <button
                                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                                disabled={page === totalPages}
                                className="btn btn-secondary"
                            >
                                Next →
                            </button>
                        </div>
                    )}
                </>
            )}

            {/* Empty State */}
            {!loading && !error && entries.length === 0 && (
                <div className="empty-state">
                    <div className="empty-icon">📦</div>
                    <h3>No cache entries found</h3>
                    <p>
                        {appliedNamespace || appliedKey
                            ? 'Try adjusting your filters or search criteria.'
                            : 'The cache is currently empty.'}
                    </p>
                </div>
            )}
        </div>
    );
}

