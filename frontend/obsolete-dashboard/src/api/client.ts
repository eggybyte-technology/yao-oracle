// REST API client for Admin service

import type {
    ClusterOverview,
    ProxyMetrics,
    NodeMetrics,
    NamespaceStats,
    ProxyTimeseries,
    NodeTimeseries,
} from '../types/metrics';

const API_BASE = import.meta.env.VITE_ADMIN_URL || '/api';

async function fetchJSON<T>(path: string): Promise<T> {
    const response = await fetch(`${API_BASE}${path}`);
    if (!response.ok) {
        throw new Error(`API error: ${response.statusText}`);
    }
    return response.json();
}

export async function fetchOverview(): Promise<ClusterOverview> {
    return fetchJSON<ClusterOverview>('/overview');
}

export async function fetchClusterTimeseries(): Promise<{
    metrics: import('../types/metrics').TimeSeriesPoint[];
}> {
    return fetchJSON<{ metrics: import('../types/metrics').TimeSeriesPoint[] }>(
        '/cluster/timeseries'
    );
}

export async function fetchProxies(): Promise<{ proxies: ProxyMetrics[] }> {
    return fetchJSON<{ proxies: ProxyMetrics[] }>('/proxies');
}

export async function fetchProxyDetails(id: string): Promise<ProxyMetrics> {
    return fetchJSON<ProxyMetrics>(`/proxies/${id}`);
}

export async function fetchProxyTimeseries(
    id: string,
): Promise<ProxyTimeseries> {
    return fetchJSON<ProxyTimeseries>(`/proxies/${id}/timeseries`);
}

export async function fetchNodes(): Promise<{ nodes: NodeMetrics[] }> {
    return fetchJSON<{ nodes: NodeMetrics[] }>('/nodes');
}

export async function fetchNodeDetails(id: string): Promise<NodeMetrics> {
    return fetchJSON<NodeMetrics>(`/nodes/${id}`);
}

export async function fetchNodeTimeseries(id: string): Promise<NodeTimeseries> {
    return fetchJSON<NodeTimeseries>(`/nodes/${id}/timeseries`);
}

export async function fetchNamespaces(): Promise<{
    namespaces: NamespaceStats[];
}> {
    return fetchJSON<{ namespaces: NamespaceStats[] }>('/namespaces');
}

export async function fetchNamespaceDetails(
    name: string,
): Promise<NamespaceStats> {
    return fetchJSON<NamespaceStats>(`/namespaces/${name}`);
}

export async function fetchHealth(): Promise<{ status: string }> {
    return fetchJSON<{ status: string }>('/health');
}

// Cache query API
export async function fetchCacheEntries(params?: {
    namespace?: string;
    key?: string;
    page?: number;
    page_size?: number;
}): Promise<import('../types/metrics').CacheQueryResponse> {
    const queryParams = new URLSearchParams();
    if (params?.namespace) queryParams.append('namespace', params.namespace);
    if (params?.key) queryParams.append('key', params.key);
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.page_size) queryParams.append('page_size', params.page_size.toString());

    const query = queryParams.toString();
    return fetchJSON<import('../types/metrics').CacheQueryResponse>(
        `/cache${query ? `?${query}` : ''}`
    );
}

