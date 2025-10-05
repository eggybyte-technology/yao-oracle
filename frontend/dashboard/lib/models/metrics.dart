/// Data models for Yao-Oracle metrics aligned with protobuf definitions
library;

/// QPS breakdown by operation type
class QPSBreakdown {
  final int get;
  final int set;
  final int delete;

  const QPSBreakdown({
    required this.get,
    required this.set,
    required this.delete,
  });

  factory QPSBreakdown.fromJson(Map<String, dynamic> json) {
    return QPSBreakdown(
      get: json['get'] ?? 0,
      set: json['set'] ?? 0,
      delete: json['delete'] ?? 0,
    );
  }

  int get total => get + set + delete;
}

/// Latency statistics at different percentiles
class LatencyStats {
  final double p50;
  final double p90;
  final double p99;

  const LatencyStats({required this.p50, required this.p90, required this.p99});

  factory LatencyStats.fromJson(Map<String, dynamic> json) {
    return LatencyStats(
      p50: (json['p50'] ?? 0.0).toDouble(),
      p90: (json['p90'] ?? 0.0).toDouble(),
      p99: (json['p99'] ?? 0.0).toDouble(),
    );
  }
}

/// Hot key information
class HotKey {
  final String key;
  final int frequency;

  const HotKey({required this.key, required this.frequency});

  factory HotKey.fromJson(Map<String, dynamic> json) {
    return HotKey(key: json['key'] ?? '', frequency: json['frequency'] ?? 0);
  }
}

/// Proxy instance metrics
class ProxyMetrics {
  final String id;
  final String ip;
  final int uptime;
  final QPSBreakdown qps;
  final LatencyStats latency;
  final double errorRate;
  final int connections;
  final List<String> namespaces;
  final String status;

  const ProxyMetrics({
    required this.id,
    required this.ip,
    required this.uptime,
    required this.qps,
    required this.latency,
    required this.errorRate,
    required this.connections,
    required this.namespaces,
    required this.status,
  });

  factory ProxyMetrics.fromJson(Map<String, dynamic> json) {
    return ProxyMetrics(
      id: json['id'] ?? '',
      ip: json['ip'] ?? '',
      uptime: json['uptime'] ?? 0,
      qps: QPSBreakdown.fromJson(json['qps'] ?? {}),
      latency: LatencyStats.fromJson(json['latency'] ?? {}),
      errorRate: (json['error_rate'] ?? 0.0).toDouble(),
      connections: json['connections'] ?? 0,
      namespaces: List<String>.from(json['namespaces'] ?? []),
      status: json['status'] ?? 'unknown',
    );
  }

  bool get isHealthy => status == 'healthy';
}

/// Cache node metrics (simplified version for UI)
class NodeMetrics {
  final String id;
  final String address;
  final bool healthy;
  final int totalKeys;
  final int memoryUsed;
  final int memoryMax;
  final double hitRate;
  final int uptime;

  const NodeMetrics({
    required this.id,
    required this.address,
    required this.healthy,
    required this.totalKeys,
    required this.memoryUsed,
    required this.memoryMax,
    required this.hitRate,
    required this.uptime,
  });

  factory NodeMetrics.fromJson(Map<String, dynamic> json) {
    return NodeMetrics(
      id: json['id'] ?? '',
      address: json['address'] ?? '',
      healthy: json['healthy'] ?? false,
      totalKeys: json['total_keys'] ?? 0,
      memoryUsed: json['memory_used'] ?? 0,
      memoryMax: json['memory_max'] ?? 0,
      hitRate: (json['hit_rate'] ?? 0.0).toDouble(),
      uptime: json['uptime'] ?? 0,
    );
  }

  bool get isHealthy => healthy;
  String get ip => address;

  double get memoryUsagePercent =>
      memoryMax > 0 ? (memoryUsed / memoryMax) * 100 : 0;
}

/// Namespace statistics (aligned with protobuf NamespaceStats)
class NamespaceStats {
  final String name;
  final String description;
  final int keyCount;
  final double hitRate;
  final int maxMemory;
  final int defaultTTL;
  final int rateLimit;
  final double qps;
  final double memoryUsedMb;

  const NamespaceStats({
    required this.name,
    required this.description,
    required this.keyCount,
    required this.hitRate,
    required this.maxMemory,
    required this.defaultTTL,
    required this.rateLimit,
    required this.qps,
    required this.memoryUsedMb,
  });

  factory NamespaceStats.fromJson(Map<String, dynamic> json) {
    return NamespaceStats(
      name: json['name'] ?? '',
      description: json['description'] ?? '',
      keyCount: json['key_count'] ?? 0,
      hitRate: (json['hit_rate'] ?? 0.0).toDouble(),
      maxMemory: json['max_memory'] ?? 0,
      defaultTTL: json['default_ttl'] ?? 0,
      rateLimit: json['rate_limit'] ?? 0,
      qps: (json['qps'] ?? 0.0).toDouble(),
      memoryUsedMb: (json['memory_used_mb'] ?? 0.0).toDouble(),
    );
  }

  double get memoryUsagePercent =>
      maxMemory > 0 ? (memoryUsedMb * 1024 * 1024 / maxMemory) * 100 : 0;
}

/// Cache entry details
class CacheEntry {
  final String namespace;
  final String key;
  final String value;
  final int ttl;
  final int size;
  final String createdAt;
  final String accessedAt;
  final int accessCount;

  const CacheEntry({
    required this.namespace,
    required this.key,
    required this.value,
    required this.ttl,
    required this.size,
    required this.createdAt,
    required this.accessedAt,
    required this.accessCount,
  });

  factory CacheEntry.fromJson(Map<String, dynamic> json) {
    return CacheEntry(
      namespace: json['namespace'] ?? '',
      key: json['key'] ?? '',
      value: json['value'] ?? '',
      ttl: json['ttl'] ?? 0,
      size: json['size'] ?? 0,
      createdAt: json['created_at'] ?? '',
      accessedAt: json['accessed_at'] ?? '',
      accessCount: json['access_count'] ?? 0,
    );
  }
}

/// Cluster metrics data
class ClusterMetricsData {
  final double qps;
  final double latencyMs;
  final double hitRate;
  final double memoryUsedMb;
  final double healthScore;
  final int totalKeys;

  const ClusterMetricsData({
    required this.qps,
    required this.latencyMs,
    required this.hitRate,
    required this.memoryUsedMb,
    required this.healthScore,
    required this.totalKeys,
  });

  factory ClusterMetricsData.fromJson(Map<String, dynamic> json) {
    return ClusterMetricsData(
      qps: (json['qps'] ?? 0.0).toDouble(),
      latencyMs: (json['latency_ms'] ?? 0.0).toDouble(),
      hitRate: (json['hit_rate'] ?? 0.0).toDouble(),
      memoryUsedMb: (json['memory_used_mb'] ?? 0.0).toDouble(),
      healthScore: (json['health_score'] ?? 0.0).toDouble(),
      totalKeys: json['total_keys'] ?? 0,
    );
  }

  // Backwards compatibility properties
  int get totalQPS => qps.toInt();
  double get hitRatio => hitRate;
  double get avgLatencyMS => latencyMs;
}

/// Component health information
class ComponentHealth {
  final int total;
  final int healthy;
  final int unhealthy;

  const ComponentHealth({
    required this.total,
    required this.healthy,
    required this.unhealthy,
  });

  factory ComponentHealth.fromJson(Map<String, dynamic> json) {
    return ComponentHealth(
      total: json['total'] ?? 0,
      healthy: json['healthy'] ?? 0,
      unhealthy: json['unhealthy'] ?? 0,
    );
  }

  double get healthPercent => total > 0 ? (healthy / total) * 100 : 0;
}

/// Cluster overview information (simplified for UI)
class ClusterOverview {
  final int proxies;
  final int nodes;
  final ClusterMetricsData metrics;
  final DateTime lastUpdated;

  const ClusterOverview({
    required this.proxies,
    required this.nodes,
    required this.metrics,
    required this.lastUpdated,
  });

  factory ClusterOverview.fromJson(Map<String, dynamic> json) {
    return ClusterOverview(
      proxies: json['proxies'] ?? 0,
      nodes: json['nodes'] ?? 0,
      metrics: ClusterMetricsData.fromJson(json['metrics'] ?? {}),
      lastUpdated: json['last_updated'] != null
          ? DateTime.parse(json['last_updated'])
          : DateTime.now(),
    );
  }
}

/// Time series data point
class TimeSeriesPoint {
  final String timestamp;
  final QPSBreakdown? qps;
  final LatencyStats? latency;
  final double? hitRatio;

  const TimeSeriesPoint({
    required this.timestamp,
    this.qps,
    this.latency,
    this.hitRatio,
  });

  factory TimeSeriesPoint.fromJson(Map<String, dynamic> json) {
    return TimeSeriesPoint(
      timestamp: json['timestamp'] ?? '',
      qps: json['qps'] != null ? QPSBreakdown.fromJson(json['qps']) : null,
      latency: json['latency'] != null
          ? LatencyStats.fromJson(json['latency'])
          : null,
      hitRatio: json['hit_ratio']?.toDouble(),
    );
  }
}
