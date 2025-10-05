/// Global application state management using Provider
library;

import 'package:flutter/foundation.dart';
import 'grpc_client.dart';
import '../models/metrics.dart' as metrics;

/// Global application state
class AppState extends ChangeNotifier {
  final GrpcClient grpcClient;

  metrics.ClusterOverview? _overview;
  List<metrics.ProxyMetrics> _proxies = [];
  List<metrics.NodeMetrics> _nodes = [];
  List<metrics.NamespaceStats> _namespaces = [];
  List<metrics.TimeSeriesPoint> _clusterTimeseries = [];

  bool _isLoading = false;
  String? _error;
  bool _isStreamConnected = false;

  AppState({required this.grpcClient}) {
    _connectStream();
  }

  // Getters
  metrics.ClusterOverview? get overview => _overview;
  List<metrics.ProxyMetrics> get proxies => _proxies;
  List<metrics.NodeMetrics> get nodes => _nodes;
  List<metrics.NamespaceStats> get namespaces => _namespaces;
  List<metrics.TimeSeriesPoint> get clusterTimeseries => _clusterTimeseries;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isStreamConnected => _isStreamConnected;

  /// Connect to gRPC stream for real-time updates
  void _connectStream() {
    if (kDebugMode) {
      print('🔄 Connecting to gRPC metrics stream...');
    }

    try {
      grpcClient.streamMetrics().listen(
        (metrics) {
          if (!_isStreamConnected) {
            _isStreamConnected = true;
            if (kDebugMode) {
              print('✅ gRPC stream connected successfully');
            }
            notifyListeners();
          }

          _handleMetricsUpdate(metrics);
        },
        onError: (error) {
          if (kDebugMode) {
            print('❌ gRPC stream error: $error');
          }
          _isStreamConnected = false;
          _error = 'Connection error: $error';
          notifyListeners();

          // Attempt to reconnect after 5 seconds
          Future.delayed(const Duration(seconds: 5), () {
            if (kDebugMode) {
              print('🔄 Attempting to reconnect...');
            }
            _connectStream();
          });
        },
        onDone: () {
          if (kDebugMode) {
            print('👋 gRPC stream closed');
          }
          _isStreamConnected = false;
          notifyListeners();
        },
      );
    } catch (e) {
      if (kDebugMode) {
        print('❌ Failed to connect gRPC stream: $e');
      }
      _error = 'Failed to connect: $e';
      _isStreamConnected = false;
      notifyListeners();
    }
  }

  /// Handle incoming metrics updates from gRPC stream
  void _handleMetricsUpdate(Map<String, dynamic> metricsData) {
    // Update overview from global stats
    final global = metricsData['global'] as Map<String, dynamic>;
    _overview = metrics.ClusterOverview(
      proxies: global['total_proxies'] ?? 0,
      nodes: global['total_nodes'] ?? 0,
      metrics: metrics.ClusterMetricsData(
        qps: (global['qps'] ?? 0.0).toDouble(),
        latencyMs: (global['latency_ms'] ?? 0.0).toDouble(),
        hitRate: (global['hit_rate'] ?? 0.0).toDouble(),
        memoryUsedMb: (global['memory_used_mb'] ?? 0.0).toDouble(),
        healthScore: (global['health_score'] ?? 0.0).toDouble(),
        totalKeys: global['total_keys'] ?? 0,
      ),
      lastUpdated: DateTime.fromMillisecondsSinceEpoch(
        (metricsData['timestamp'] ?? 0) * 1000,
      ),
    );

    // Update nodes from node stats
    final nodesList = metricsData['nodes'] as List;
    _nodes = nodesList
        .map(
          (node) => metrics.NodeMetrics(
            id: node['id'] ?? '',
            address: node['id'] ?? '',
            healthy: node['healthy'] ?? false,
            totalKeys: node['key_count'] ?? 0,
            memoryUsed: ((node['memory_used_mb'] ?? 0.0) * 1024 * 1024).toInt(),
            memoryMax: (512 * 1024 * 1024), // Default max
            hitRate: (node['hit_rate'] ?? 0.0).toDouble(),
            uptime: node['uptime_seconds'] ?? 0,
          ),
        )
        .toList();

    // Update namespaces from namespace stats
    final namespacesList = metricsData['namespaces'] as List;
    _namespaces = namespacesList
        .map(
          (ns) => metrics.NamespaceStats(
            name: ns['name'] ?? '',
            description: ns['description'] ?? '',
            keyCount: ns['keys'] ?? 0,
            hitRate: (ns['hit_rate'] ?? 0.0).toDouble(),
            maxMemory: (ns['max_memory_mb'] ?? 0) * 1024 * 1024,
            defaultTTL: ns['default_ttl'] ?? 0,
            rateLimit: ns['rate_limit_qps'] ?? 0,
            qps: (ns['qps'] ?? 0.0).toDouble(),
            memoryUsedMb: (ns['memory_used_mb'] ?? 0.0).toDouble(),
          ),
        )
        .toList();

    notifyListeners();
  }

  /// Load all data
  Future<void> loadAll() async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      await Future.wait([
        loadOverview(),
        loadProxies(),
        loadNodes(),
        loadNamespaces(),
        loadClusterTimeseries(),
      ]);
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  /// Load cluster overview
  Future<void> loadOverview() async {
    try {
      _overview = await grpcClient.getOverview();
      notifyListeners();
    } catch (e) {
      if (kDebugMode) {
        print('Failed to load overview: $e');
      }
      rethrow;
    }
  }

  /// Load all proxies
  Future<void> loadProxies() async {
    try {
      _proxies = await grpcClient.getProxies();
      notifyListeners();
    } catch (e) {
      if (kDebugMode) {
        print('Failed to load proxies: $e');
      }
      rethrow;
    }
  }

  /// Load all nodes
  Future<void> loadNodes() async {
    try {
      _nodes = await grpcClient.getNodes();
      notifyListeners();
    } catch (e) {
      if (kDebugMode) {
        print('Failed to load nodes: $e');
      }
      rethrow;
    }
  }

  /// Load all namespaces
  Future<void> loadNamespaces() async {
    try {
      _namespaces = await grpcClient.getNamespaces();
      notifyListeners();
    } catch (e) {
      if (kDebugMode) {
        print('Failed to load namespaces: $e');
      }
      rethrow;
    }
  }

  /// Load cluster timeseries
  Future<void> loadClusterTimeseries() async {
    try {
      _clusterTimeseries = await grpcClient.getClusterTimeseries();
      notifyListeners();
    } catch (e) {
      if (kDebugMode) {
        print('Failed to load cluster timeseries: $e');
      }
      rethrow;
    }
  }

  @override
  void dispose() {
    grpcClient.dispose();
    super.dispose();
  }
}
