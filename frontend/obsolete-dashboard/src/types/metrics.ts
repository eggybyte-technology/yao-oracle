// Metrics data types matching Admin service API

export interface QPSBreakdown {
    get: number;
    set: number;
    delete: number;
}

export interface LatencyStats {
    p50: number;
    p90: number;
    p99: number;
}

export interface MemoryStats {
    used: number;
    max: number;
}

export interface HotKey {
    key: string;
    frequency: number;
}

export interface ProxyMetrics {
    id: string;
    ip: string;
    uptime: number;
    qps: QPSBreakdown;
    latency: LatencyStats;
    error_rate: number;
    connections: number;
    namespaces: string[];
    status: 'healthy' | 'degraded' | 'down';
}

export interface NodeMetrics {
    id: string;
    ip: string;
    uptime: number;
    memory: MemoryStats;
    key_count: number;
    hit_count: number;
    miss_count: number;
    hot_keys: HotKey[];
    status: 'healthy' | 'full' | 'down';
}

export interface NamespaceStats {
    name: string;
    description: string;
    api_key_masked: string;
    memory_used: number;
    memory_limit: number;
    key_count: number;
    max_keys: number;
    qps: QPSBreakdown;
    hit_ratio: number;
    rate_limit_qps: number;
    default_ttl: number;
}

export interface ClusterOverview {
    proxies: {
        total: number;
        healthy: number;
        unhealthy: number;
    };
    nodes: {
        total: number;
        healthy: number;
        unhealthy: number;
    };
    metrics: {
        total_qps: number;
        total_keys: number;
        hit_ratio: number;
        avg_latency_ms: number;
    };
    last_updated: string;
}

export interface TimeSeriesPoint {
    timestamp: string;
    qps?: QPSBreakdown;
    latency?: LatencyStats;
    memory?: MemoryStats;
    hit_ratio?: number;
}

export interface ProxyTimeseries {
    instance_id: string;
    metrics: TimeSeriesPoint[];
}

export interface NodeTimeseries {
    instance_id: string;
    metrics: TimeSeriesPoint[];
}

// WebSocket message types
export type WebSocketMessageType =
    | 'overview_update'
    | 'proxy_update'
    | 'node_update'
    | 'event';

export interface WebSocketMessage {
    type: WebSocketMessageType;
    data: any;
}

export interface ClusterEvent {
    severity: 'info' | 'warning' | 'error';
    message: string;
    timestamp: string;
}

// Cache data types
export interface CacheEntry {
    namespace: string;
    key: string;
    value: string;
    ttl: number; // Seconds remaining
    size: number; // Bytes
    created_at: string;
    accessed_at: string;
    access_count: number;
}

export interface CacheQueryResponse {
    entries: CacheEntry[];
    total: number;
    page: number;
    page_size: number;
}

