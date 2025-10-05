/// Nodes monitoring page
library;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import '../core/app_state.dart';
import '../widgets/metric_card.dart';
import '../widgets/loading_widget.dart' as loading;

/// Nodes page showing all cache nodes and their metrics
class NodesPage extends StatelessWidget {
  const NodesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<AppState>(
      builder: (context, appState, child) {
        if (appState.isLoading && appState.nodes.isEmpty) {
          return const loading.LoadingWidget(message: 'Loading nodes...');
        }

        if (appState.error != null) {
          return loading.ErrorWidget(
            message: appState.error!,
            onRetry: () => appState.loadNodes(),
          );
        }

        if (appState.nodes.isEmpty) {
          return const loading.EmptyStateWidget(
            message: 'No cache nodes found',
            icon: Icons.storage,
          );
        }

        return RefreshIndicator(
          onRefresh: () => appState.loadNodes(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'Cache Nodes',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 16),
              ...appState.nodes.map((node) => _NodeCard(node: node)),
            ],
          ),
        );
      },
    );
  }
}

class _NodeCard extends StatelessWidget {
  final dynamic node;

  const _NodeCard({required this.node});

  @override
  Widget build(BuildContext context) {
    final memoryUsedMB = (node.memoryUsed / (1024 * 1024)).toInt();
    final memoryMaxMB = (node.memoryMax / (1024 * 1024)).toInt();
    
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.storage,
                  color: node.healthy ? Colors.green : Colors.red,
                  size: 28,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        node.id,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      Text(
                        node.address,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: node.healthy 
                        ? Colors.green.withOpacity(0.2)
                        : Colors.red.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        decoration: BoxDecoration(
                          color: node.healthy ? Colors.green : Colors.red,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 6),
                      Text(
                        node.healthy ? 'HEALTHY' : 'UNHEALTHY',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: node.healthy ? Colors.green.shade700 : Colors.red.shade700,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // Memory usage
            _buildMemoryUsage(context, memoryUsedMB, memoryMaxMB, node.memoryUsagePercent),
            const SizedBox(height: 16),

            // Key metrics
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                _MetricChip(
                  label: 'Total Keys',
                  value: NumberFormat.compact().format(node.totalKeys),
                  icon: Icons.key,
                  color: Colors.blue,
                ),
                _MetricChip(
                  label: 'Hit Rate',
                  value: '${(node.hitRate * 100).toStringAsFixed(1)}%',
                  icon: Icons.analytics,
                  color: Colors.green,
                ),
                _MetricChip(
                  label: 'Memory',
                  value: '${memoryUsedMB} MB',
                  icon: Icons.memory,
                  color: Colors.purple,
                ),
                _MetricChip(
                  label: 'Uptime',
                  value: _formatUptime(node.uptime),
                  icon: Icons.access_time,
                  color: Colors.orange,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMemoryUsage(BuildContext context, int usedMB, int maxMB, double usagePercent) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text('Memory Usage', style: Theme.of(context).textTheme.bodySmall),
            Text(
              '$usedMB MB / $maxMB MB',
              style: Theme.of(
                context,
              ).textTheme.bodySmall?.copyWith(fontWeight: FontWeight.bold),
            ),
          ],
        ),
        const SizedBox(height: 8),
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: usagePercent / 100,
            backgroundColor: Colors.grey.withOpacity(0.2),
            valueColor: AlwaysStoppedAnimation<Color>(
              _getColorForUsage(usagePercent),
            ),
            minHeight: 8,
          ),
        ),
      ],
    );
  }

  Color _getColorForUsage(double percent) {
    if (percent < 60) return Colors.green;
    if (percent < 80) return Colors.orange;
    return Colors.red;
  }

  String _formatUptime(int seconds) {
    final duration = Duration(seconds: seconds);
    if (duration.inDays > 0) {
      return '${duration.inDays}d ${duration.inHours % 24}h';
    } else if (duration.inHours > 0) {
      return '${duration.inHours}h ${duration.inMinutes % 60}m';
    } else {
      return '${duration.inMinutes}m';
    }
  }
}

class _MetricChip extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;
  final Color color;

  const _MetricChip({
    required this.label,
    required this.value,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withOpacity(0.3), width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 18, color: color),
          const SizedBox(width: 8),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                label,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(
                    context,
                  ).colorScheme.onSurface.withOpacity(0.6),
                ),
              ),
              Text(
                value,
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: color,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
