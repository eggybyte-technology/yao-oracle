/// HTTP API client for Yao-Oracle Admin service
library;

import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';
import '../models/metrics.dart';

/// API client for communicating with mock-admin service
class ApiClient {
  final String baseUrl;
  final String wsUrl;
  WebSocketChannel? _wsChannel;
  StreamController<Map<String, dynamic>>? _wsController;

  ApiClient({required this.baseUrl, required this.wsUrl});

  /// Get cluster overview
  Future<ClusterOverview> getOverview() async {
    final response = await http.get(Uri.parse('$baseUrl/overview'));
    if (response.statusCode == 200) {
      return ClusterOverview.fromJson(json.decode(response.body));
    }
    throw Exception('Failed to load overview: ${response.statusCode}');
  }

  /// Get cluster time series data
  Future<List<TimeSeriesPoint>> getClusterTimeseries() async {
    final response = await http.get(Uri.parse('$baseUrl/cluster/timeseries'));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return (data['metrics'] as List)
          .map((e) => TimeSeriesPoint.fromJson(e))
          .toList();
    }
    throw Exception(
      'Failed to load cluster timeseries: ${response.statusCode}',
    );
  }

  /// Get all proxy instances
  Future<List<ProxyMetrics>> getProxies() async {
    final response = await http.get(Uri.parse('$baseUrl/proxies'));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return (data['proxies'] as List)
          .map((e) => ProxyMetrics.fromJson(e))
          .toList();
    }
    throw Exception('Failed to load proxies: ${response.statusCode}');
  }

  /// Get proxy details by ID
  Future<ProxyMetrics> getProxy(String id) async {
    final response = await http.get(Uri.parse('$baseUrl/proxies/$id'));
    if (response.statusCode == 200) {
      return ProxyMetrics.fromJson(json.decode(response.body));
    }
    throw Exception('Failed to load proxy $id: ${response.statusCode}');
  }

  /// Get proxy time series data
  Future<List<TimeSeriesPoint>> getProxyTimeseries(String id) async {
    final response = await http.get(
      Uri.parse('$baseUrl/proxies/$id/timeseries'),
    );
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return (data['metrics'] as List)
          .map((e) => TimeSeriesPoint.fromJson(e))
          .toList();
    }
    throw Exception('Failed to load proxy timeseries: ${response.statusCode}');
  }

  /// Get all cache nodes
  Future<List<NodeMetrics>> getNodes() async {
    final response = await http.get(Uri.parse('$baseUrl/nodes'));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return (data['nodes'] as List)
          .map((e) => NodeMetrics.fromJson(e))
          .toList();
    }
    throw Exception('Failed to load nodes: ${response.statusCode}');
  }

  /// Get node details by ID
  Future<NodeMetrics> getNode(String id) async {
    final response = await http.get(Uri.parse('$baseUrl/nodes/$id'));
    if (response.statusCode == 200) {
      return NodeMetrics.fromJson(json.decode(response.body));
    }
    throw Exception('Failed to load node $id: ${response.statusCode}');
  }

  /// Get node time series data
  Future<List<TimeSeriesPoint>> getNodeTimeseries(String id) async {
    final response = await http.get(Uri.parse('$baseUrl/nodes/$id/timeseries'));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return (data['metrics'] as List)
          .map((e) => TimeSeriesPoint.fromJson(e))
          .toList();
    }
    throw Exception('Failed to load node timeseries: ${response.statusCode}');
  }

  /// Get all namespaces
  Future<List<NamespaceStats>> getNamespaces() async {
    final response = await http.get(Uri.parse('$baseUrl/namespaces'));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return (data['namespaces'] as List)
          .map((e) => NamespaceStats.fromJson(e))
          .toList();
    }
    throw Exception('Failed to load namespaces: ${response.statusCode}');
  }

  /// Get namespace details by name
  Future<NamespaceStats> getNamespace(String name) async {
    final response = await http.get(Uri.parse('$baseUrl/namespaces/$name'));
    if (response.statusCode == 200) {
      return NamespaceStats.fromJson(json.decode(response.body));
    }
    throw Exception('Failed to load namespace $name: ${response.statusCode}');
  }

  /// Query cache entries with filters
  Future<Map<String, dynamic>> queryCacheEntries({
    String? namespace,
    String? key,
    int page = 1,
    int pageSize = 20,
  }) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'page_size': pageSize.toString(),
    };
    if (namespace != null && namespace.isNotEmpty) {
      queryParams['namespace'] = namespace;
    }
    if (key != null && key.isNotEmpty) {
      queryParams['key'] = key;
    }

    final uri = Uri.parse(
      '$baseUrl/cache',
    ).replace(queryParameters: queryParams);
    final response = await http.get(uri);

    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return {
        'entries': (data['entries'] as List)
            .map((e) => CacheEntry.fromJson(e))
            .toList(),
        'total': data['total'],
        'page': data['page'],
        'page_size': data['page_size'],
      };
    }
    throw Exception('Failed to query cache: ${response.statusCode}');
  }

  /// Connect to WebSocket for real-time updates
  Stream<Map<String, dynamic>> connectWebSocket() {
    _wsController = StreamController<Map<String, dynamic>>.broadcast();

    try {
      _wsChannel = WebSocketChannel.connect(Uri.parse(wsUrl));

      _wsChannel!.stream.listen(
        (message) {
          try {
            final data = json.decode(message as String);
            _wsController!.add(data);
          } catch (e) {
            print('Error parsing WebSocket message: $e');
          }
        },
        onError: (error) {
          print('WebSocket error: $error');
          _wsController!.addError(error);
        },
        onDone: () {
          print('WebSocket connection closed');
          _wsController!.close();
        },
      );
    } catch (e) {
      print('Failed to connect WebSocket: $e');
      _wsController!.addError(e);
    }

    return _wsController!.stream;
  }

  /// Close WebSocket connection
  void closeWebSocket() {
    _wsChannel?.sink.close();
    _wsController?.close();
    _wsChannel = null;
    _wsController = null;
  }

  /// Dispose resources
  void dispose() {
    closeWebSocket();
  }
}
