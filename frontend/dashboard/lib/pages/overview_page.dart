/// Cluster overview page
library;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import '../core/app_state.dart';
import '../models/metrics.dart';
import '../widgets/metric_card.dart';
import '../widgets/loading_widget.dart' as loading;

/// Cluster overview page displaying key metrics
class OverviewPage extends StatelessWidget {
  const OverviewPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<AppState>(
      builder: (context, appState, child) {
        if (appState.isLoading && appState.overview == null) {
          return const loading.LoadingWidget(message: 'Loading overview...');
        }

        if (appState.error != null) {
          return loading.ErrorWidget(
            message: appState.error!,
            onRetry: () => appState.loadAll(),
          );
        }

        final overview = appState.overview;
        if (overview == null) {
          return const loading.EmptyStateWidget(message: 'No data available');
        }

        return RefreshIndicator(
          onRefresh: () => appState.loadAll(),
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header
                _buildHeader(context, overview, appState.isStreamConnected),
                const SizedBox(height: 24),

                // Component Health
                Text(
                  'Component Health',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 12),
                _buildComponentHealth(overview),
                const SizedBox(height: 24),

                // Cluster Metrics
                Text(
                  'Cluster Metrics',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 12),
                _buildClusterMetrics(overview),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildHeader(
    BuildContext context,
    ClusterOverview overview,
    bool isWsConnected,
  ) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.dashboard,
                  size: 32,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Yao-Oracle Cluster',
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                      Text(
                        'Last updated: ${_formatTime(overview.lastUpdated.toIso8601String())}',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Theme.of(
                            context,
                          ).colorScheme.onSurface.withOpacity(0.6),
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: isWsConnected
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
                          color: isWsConnected ? Colors.green : Colors.grey,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 6),
                      Text(
                        isWsConnected ? 'LIVE' : 'OFFLINE',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: isWsConnected
                              ? Colors.green.shade700
                              : Colors.grey.shade700,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildComponentHealth(ClusterOverview overview) {
    // Create component health from simple counts
    final proxiesHealth = ComponentHealth(
      total: overview.proxies,
      healthy: overview.proxies, // Assume all healthy for now
      unhealthy: 0,
    );
    final nodesHealth = ComponentHealth(
      total: overview.nodes,
      healthy: overview.nodes, // Assume all healthy for now
      unhealthy: 0,
    );

    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 600;
        return Wrap(
          spacing: 16,
          runSpacing: 16,
          children: [
            SizedBox(
              width: isWide
                  ? (constraints.maxWidth - 16) / 2
                  : constraints.maxWidth,
              child: _ComponentHealthCard(
                title: 'Proxy Instances',
                health: proxiesHealth,
                icon: Icons.router,
                color: Colors.blue,
              ),
            ),
            SizedBox(
              width: isWide
                  ? (constraints.maxWidth - 16) / 2
                  : constraints.maxWidth,
              child: _ComponentHealthCard(
                title: 'Cache Nodes',
                health: nodesHealth,
                icon: Icons.storage,
                color: Colors.purple,
              ),
            ),
          ],
        );
      },
    );
  }

  Widget _buildClusterMetrics(ClusterOverview overview) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 900;
        final cardWidth = isWide
            ? (constraints.maxWidth - 48) / 4
            : (constraints.maxWidth > 600
                  ? (constraints.maxWidth - 16) / 2
                  : constraints.maxWidth);

        return Wrap(
          spacing: 16,
          runSpacing: 16,
          children: [
            SizedBox(
              width: cardWidth,
              child: MetricCard(
                label: 'Total QPS',
                value: NumberFormat.compact().format(overview.metrics.totalQPS),
                icon: Icons.speed,
                color: Colors.green,
                subtitle: 'Queries per second',
              ),
            ),
            SizedBox(
              width: cardWidth,
              child: MetricCard(
                label: 'Total Keys',
                value: NumberFormat.compact().format(
                  overview.metrics.totalKeys,
                ),
                icon: Icons.key,
                color: Colors.orange,
                subtitle: 'Cached entries',
              ),
            ),
            SizedBox(
              width: cardWidth,
              child: MetricCard(
                label: 'Hit Ratio',
                value:
                    '${(overview.metrics.hitRatio * 100).toStringAsFixed(1)}%',
                icon: Icons.analytics,
                color: Colors.blue,
                subtitle: 'Cache efficiency',
              ),
            ),
            SizedBox(
              width: cardWidth,
              child: MetricCard(
                label: 'Avg Latency',
                value: '${overview.metrics.avgLatencyMS.toStringAsFixed(2)}ms',
                icon: Icons.timer,
                color: Colors.purple,
                subtitle: 'Response time',
              ),
            ),
          ],
        );
      },
    );
  }

  String _formatTime(String? timestamp) {
    if (timestamp == null || timestamp.isEmpty) return 'N/A';
    try {
      final dt = DateTime.parse(timestamp);
      return DateFormat('MMM dd, HH:mm:ss').format(dt.toLocal());
    } catch (e) {
      return 'N/A';
    }
  }
}

class _ComponentHealthCard extends StatelessWidget {
  final String title;
  final ComponentHealth health;
  final IconData icon;
  final Color color;

  const _ComponentHealthCard({
    required this.title,
    required this.health,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, color: color, size: 24),
                const SizedBox(width: 12),
                Text(title, style: Theme.of(context).textTheme.titleMedium),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _HealthStat(
                  label: 'Total',
                  value: health.total.toString(),
                  color: Colors.grey,
                ),
                _HealthStat(
                  label: 'Healthy',
                  value: health.healthy.toString(),
                  color: Colors.green,
                ),
                _HealthStat(
                  label: 'Unhealthy',
                  value: health.unhealthy.toString(),
                  color: Colors.red,
                ),
              ],
            ),
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: health.total > 0 ? health.healthy / health.total : 0,
                backgroundColor: Colors.red.withOpacity(0.2),
                valueColor: const AlwaysStoppedAnimation<Color>(Colors.green),
                minHeight: 8,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              '${health.healthPercent.toStringAsFixed(0)}% Healthy',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _HealthStat extends StatelessWidget {
  final String label;
  final String value;
  final Color color;

  const _HealthStat({
    required this.label,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(
          value,
          style: Theme.of(context).textTheme.headlineMedium?.copyWith(
            color: color,
            fontWeight: FontWeight.bold,
          ),
        ),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
            color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
          ),
        ),
      ],
    );
  }
}
