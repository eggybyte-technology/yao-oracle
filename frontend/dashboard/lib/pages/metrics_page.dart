/// Real-time metrics visualization page
library;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/app_state.dart';
import '../widgets/metrics_chart.dart';
import '../widgets/loading_widget.dart' as loading;

/// Real-time metrics page with live charts
class MetricsPage extends StatefulWidget {
  const MetricsPage({super.key});

  @override
  State<MetricsPage> createState() => _MetricsPageState();
}

class _MetricsPageState extends State<MetricsPage> {
  // Store historical data for charts (max 30 data points = 2.5 minutes at 5s interval)
  final List<MetricsDataPoint> _qpsHistory = [];
  final List<MetricsDataPoint> _hitRateHistory = [];
  final List<MetricsDataPoint> _memoryHistory = [];
  final List<MetricsDataPoint> _latencyHistory = [];

  static const int _maxDataPoints = 30;

  @override
  Widget build(BuildContext context) {
    return Consumer<AppState>(
      builder: (context, appState, child) {
        // Update historical data when new metrics arrive
        if (appState.overview != null) {
          final now = DateTime.now();
          final metrics = appState.overview!.metrics;

          _addDataPoint(_qpsHistory, now, metrics.qps);
          _addDataPoint(_hitRateHistory, now, metrics.hitRate * 100);
          _addDataPoint(_memoryHistory, now, metrics.memoryUsedMb);
          _addDataPoint(_latencyHistory, now, metrics.latencyMs);
        }

        if (appState.isLoading && appState.overview == null) {
          return const loading.LoadingWidget(message: 'Loading metrics...');
        }

        if (appState.error != null) {
          return loading.ErrorWidget(
            message: appState.error!,
            onRetry: () => appState.loadAll(),
          );
        }

        return RefreshIndicator(
          onRefresh: () => appState.loadAll(),
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Page header
                Row(
                  children: [
                    Icon(
                      Icons.show_chart,
                      size: 28,
                      color: Theme.of(context).colorScheme.primary,
                    ),
                    const SizedBox(width: 12),
                    Text(
                      'Real-Time Metrics',
                      style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    const Spacer(),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: appState.isStreamConnected
                            ? Colors.green.withOpacity(0.2)
                            : Colors.grey.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 8,
                            height: 8,
                            decoration: BoxDecoration(
                              color: appState.isStreamConnected
                                  ? Colors.green
                                  : Colors.grey,
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            appState.isStreamConnected
                                ? 'LIVE STREAM'
                                : 'OFFLINE',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: appState.isStreamConnected
                                  ? Colors.green.shade700
                                  : Colors.grey.shade700,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                // Charts
                MetricsChart(
                  title: 'Queries Per Second (QPS)',
                  dataPoints: _qpsHistory,
                  yAxisLabel: 'QPS',
                  lineColor: Colors.blue,
                ),
                const SizedBox(height: 16),
                MetricsChart(
                  title: 'Hit Rate',
                  dataPoints: _hitRateHistory,
                  yAxisLabel: 'Hit Rate %',
                  lineColor: Colors.green,
                  maxY: 100,
                  showPercentage: true,
                ),
                const SizedBox(height: 16),
                MetricsChart(
                  title: 'Memory Usage',
                  dataPoints: _memoryHistory,
                  yAxisLabel: 'Memory (MB)',
                  lineColor: Colors.purple,
                ),
                const SizedBox(height: 16),
                MetricsChart(
                  title: 'Average Latency',
                  dataPoints: _latencyHistory,
                  yAxisLabel: 'Latency (ms)',
                  lineColor: Colors.orange,
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  void _addDataPoint(List<MetricsDataPoint> history, DateTime timestamp, double value) {
    history.add(MetricsDataPoint(timestamp: timestamp, value: value));
    if (history.length > _maxDataPoints) {
      history.removeAt(0);
    }
  }
}

