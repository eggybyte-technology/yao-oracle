/// gRPC client for Yao-Oracle Dashboard service
library;

import 'dart:async';
import 'package:grpc/grpc.dart';
import '../generated/yao/oracle/v1/dashboard.pbgrpc.dart';
import '../models/metrics.dart' as metrics;

/// gRPC client for communicating with Yao-Oracle Admin service
class GrpcClient {
  final String host;
  final int port;
  late ClientChannel _channel;
  late DashboardServiceClient _client;
  StreamSubscription<ClusterMetrics>? _metricsSubscription;
  final StreamController<Map<String, dynamic>> _metricsController =
      StreamController<Map<String, dynamic>>.broadcast();

  GrpcClient({required this.host, required this.port}) {
    _initializeChannel();
  }

  /// Initialize gRPC channel and client
  void _initializeChannel() {
    print('🔌 Initializing gRPC connection to $host:$port');
    _channel = ClientChannel(
      host,
      port: port,
      options: const ChannelOptions(
        credentials: ChannelCredentials.insecure(),
        // Configure timeouts and keepalive for better connection stability
        connectionTimeout: Duration(seconds: 10),
        idleTimeout: Duration(minutes: 5),
      ),
    );
    _client = DashboardServiceClient(_channel);
    print('✅ gRPC client initialized');
  }

  /// Subscribe to real-time cluster metrics stream
  ///
  /// Parameters:
  ///   - namespace: Optional namespace filter. Empty string subscribes to global metrics.
  ///
  /// Returns a stream of metrics updates
  Stream<Map<String, dynamic>> streamMetrics({String namespace = ''}) {
    // Cancel existing subscription if any
    _metricsSubscription?.cancel();

    final request = SubscribeRequest()..namespace = namespace;

    final namespaceStr = namespace.isEmpty ? 'all' : namespace;
    print('📊 Subscribing to metrics stream (namespace: $namespaceStr)');

    try {
      final stream = _client.streamMetrics(request);

      _metricsSubscription = stream.listen(
        (clusterMetrics) {
          // Convert protobuf ClusterMetrics to Map for backwards compatibility
          final data = _convertClusterMetrics(clusterMetrics);
          print(
            '✅ Received metrics update: QPS=${data['global']['qps']?.toStringAsFixed(1)}, '
            'Nodes=${data['nodes']?.length ?? 0}, '
            'Namespaces=${data['namespaces']?.length ?? 0}',
          );
          _metricsController.add(data);
        },
        onError: (error) {
          print('❌ gRPC stream error: $error');
          _metricsController.addError(error);
          // Attempt to reconnect after error
          Future.delayed(const Duration(seconds: 5), () {
            print('🔄 Attempting to reconnect...');
            streamMetrics(namespace: namespace);
          });
        },
        onDone: () {
          print('👋 gRPC stream closed by server');
        },
      );

      print('✅ Metrics stream subscription active');
    } catch (e) {
      print('❌ Failed to start metrics stream: $e');
      _metricsController.addError(e);
    }

    return _metricsController.stream;
  }

  /// Convert protobuf ClusterMetrics to Map
  Map<String, dynamic> _convertClusterMetrics(ClusterMetrics metrics) {
    return {
      'timestamp': metrics.timestamp,
      'global': {
        'qps': metrics.global.qps,
        'latency_ms': metrics.global.latencyMs,
        'hit_rate': metrics.global.hitRate,
        'memory_used_mb': metrics.global.memoryUsedMb,
        'health_score': metrics.global.healthScore,
        'total_keys': metrics.global.totalKeys,
        'total_proxies': metrics.global.totalProxies,
        'total_nodes': metrics.global.totalNodes,
        'healthy_nodes': metrics.global.healthyNodes,
      },
      'namespaces': metrics.namespaces
          .map(
            (ns) => {
              'name': ns.name,
              'qps': ns.qps,
              'hit_rate': ns.hitRate,
              'ttl_avg': ns.ttlAvg,
              'keys': ns.keys,
              'memory_used_mb': ns.memoryUsedMb,
              'api_key': ns.apiKey,
              'description': ns.description,
              'max_memory_mb': ns.maxMemoryMb,
              'default_ttl': ns.defaultTtl,
              'rate_limit_qps': ns.rateLimitQps,
            },
          )
          .toList(),
      'nodes': metrics.nodes
          .map(
            (node) => {
              'id': node.id,
              'ip': node.ip,
              'namespace': node.namespace,
              'memory_used_mb': node.memoryUsedMb,
              'hit_rate': node.hitRate,
              'latency_ms': node.latencyMs,
              'key_count': node.keyCount,
              'healthy': node.healthy,
              'uptime_seconds': node.uptimeSeconds,
              'qps': node.qps,
            },
          )
          .toList(),
    };
  }

  /// Query a specific cache entry
  Future<CacheQueryResponse> queryCache({
    required String namespace,
    required String key,
  }) async {
    final request = CacheQueryRequest()
      ..namespace = namespace
      ..key = key;

    try {
      final response = await _client.queryCache(request);
      return response;
    } on GrpcError catch (e) {
      print('gRPC error querying cache: ${e.message}');
      rethrow;
    }
  }

  /// Update API key for a namespace (Secret management)
  Future<SecretUpdateResponse> manageSecret({
    required String namespace,
    required String newApiKey,
  }) async {
    final request = SecretUpdateRequest()
      ..namespace = namespace
      ..newApiKey = newApiKey;

    try {
      final response = await _client.manageSecret(request);
      return response;
    } on GrpcError catch (e) {
      print('gRPC error updating secret: ${e.message}');
      rethrow;
    }
  }

  /// Get configuration for all namespaces
  Future<ConfigResponse> getConfig() async {
    final request = ConfigRequest();

    try {
      final response = await _client.getConfig(request);
      return response;
    } on GrpcError catch (e) {
      print('gRPC error getting config: ${e.message}');
      rethrow;
    }
  }

  /// Get cluster overview (derived from metrics stream)
  ///
  /// This method provides backwards compatibility with the old HTTP API
  Future<metrics.ClusterOverview> getOverview() async {
    // Subscribe to metrics stream and wait for first message
    final completer = Completer<metrics.ClusterOverview>();
    StreamSubscription? subscription;

    subscription = streamMetrics().listen(
      (data) {
        if (!completer.isCompleted) {
          final overview = metrics.ClusterOverview(
            proxies: data['global']['total_proxies'] ?? 0,
            nodes: data['global']['total_nodes'] ?? 0,
            metrics: metrics.ClusterMetricsData(
              qps: data['global']['qps'],
              latencyMs: data['global']['latency_ms'],
              hitRate: data['global']['hit_rate'],
              memoryUsedMb: data['global']['memory_used_mb'],
              healthScore: data['global']['health_score'],
              totalKeys: data['global']['total_keys'],
            ),
            lastUpdated: DateTime.fromMillisecondsSinceEpoch(
              data['timestamp'] * 1000,
            ),
          );
          completer.complete(overview);
          subscription?.cancel();
        }
      },
      onError: (error) {
        if (!completer.isCompleted) {
          completer.completeError(error);
          subscription?.cancel();
        }
      },
    );

    // Set timeout
    Future.delayed(const Duration(seconds: 10), () {
      if (!completer.isCompleted) {
        subscription?.cancel();
        completer.completeError(
          TimeoutException('Timeout waiting for metrics'),
        );
      }
    });

    return completer.future;
  }

  /// Get all proxies (derived from metrics stream)
  Future<List<metrics.ProxyMetrics>> getProxies() async {
    // For now, return empty list as we aggregate at cluster level
    // Can be implemented if we add per-proxy metrics to proto
    return [];
  }

  /// Get all cache nodes (derived from metrics stream)
  Future<List<metrics.NodeMetrics>> getNodes() async {
    final completer = Completer<List<metrics.NodeMetrics>>();
    StreamSubscription? subscription;

    subscription = streamMetrics().listen(
      (data) {
        if (!completer.isCompleted) {
          final nodesList = (data['nodes'] as List)
              .map(
                (node) => metrics.NodeMetrics(
                  id: node['id'],
                  address: node['id'],
                  healthy: node['healthy'],
                  totalKeys: node['key_count'],
                  memoryUsed: (node['memory_used_mb'] * 1024 * 1024).toInt(),
                  memoryMax: (512 * 1024 * 1024), // Default max
                  hitRate: node['hit_rate'],
                  uptime: node['uptime_seconds'],
                ),
              )
              .toList();
          completer.complete(nodesList);
          subscription?.cancel();
        }
      },
      onError: (error) {
        if (!completer.isCompleted) {
          completer.completeError(error);
          subscription?.cancel();
        }
      },
    );

    Future.delayed(const Duration(seconds: 10), () {
      if (!completer.isCompleted) {
        subscription?.cancel();
        completer.completeError(
          TimeoutException('Timeout waiting for metrics'),
        );
      }
    });

    return completer.future;
  }

  /// Get all namespaces (derived from metrics stream)
  Future<List<metrics.NamespaceStats>> getNamespaces() async {
    final completer = Completer<List<metrics.NamespaceStats>>();
    StreamSubscription? subscription;

    subscription = streamMetrics().listen(
      (data) {
        if (!completer.isCompleted) {
          final namespacesList = (data['namespaces'] as List)
              .map(
                (ns) => metrics.NamespaceStats(
                  name: ns['name'],
                  description: ns['description'] ?? '',
                  keyCount: ns['keys'],
                  hitRate: ns['hit_rate'],
                  maxMemory: ns['max_memory_mb'] * 1024 * 1024,
                  defaultTTL: ns['default_ttl'],
                  rateLimit: ns['rate_limit_qps'],
                  qps: ns['qps'],
                  memoryUsedMb: ns['memory_used_mb'],
                ),
              )
              .toList();
          completer.complete(namespacesList);
          subscription?.cancel();
        }
      },
      onError: (error) {
        if (!completer.isCompleted) {
          completer.completeError(error);
          subscription?.cancel();
        }
      },
    );

    Future.delayed(const Duration(seconds: 10), () {
      if (!completer.isCompleted) {
        subscription?.cancel();
        completer.completeError(
          TimeoutException('Timeout waiting for metrics'),
        );
      }
    });

    return completer.future;
  }

  /// Get cluster time series data (mock for now)
  Future<List<metrics.TimeSeriesPoint>> getClusterTimeseries() async {
    // In the future, this could be added to the proto definition
    // For now, return empty list
    return [];
  }

  /// Close gRPC channel and cleanup
  Future<void> dispose() async {
    await _metricsSubscription?.cancel();
    await _metricsController.close();
    await _channel.shutdown();
  }
}
