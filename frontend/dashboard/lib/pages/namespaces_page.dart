/// Namespaces monitoring page
library;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import '../core/app_state.dart';
import '../widgets/loading_widget.dart' as loading;

/// Namespaces page showing all business namespaces and their statistics
class NamespacesPage extends StatelessWidget {
  const NamespacesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<AppState>(
      builder: (context, appState, child) {
        if (appState.isLoading && appState.namespaces.isEmpty) {
          return const loading.LoadingWidget(message: 'Loading namespaces...');
        }

        if (appState.error != null) {
          return loading.ErrorWidget(
            message: appState.error!,
            onRetry: () => appState.loadNamespaces(),
          );
        }

        if (appState.namespaces.isEmpty) {
          return const loading.EmptyStateWidget(
            message: 'No namespaces configured',
            icon: Icons.folder,
          );
        }

        return RefreshIndicator(
          onRefresh: () => appState.loadNamespaces(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'Business Namespaces',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 16),
              ...appState.namespaces.map((ns) => _NamespaceCard(namespace: ns)),
            ],
          ),
        );
      },
    );
  }
}

class _NamespaceCard extends StatelessWidget {
  final dynamic namespace;

  const _NamespaceCard({required this.namespace});

  @override
  Widget build(BuildContext context) {
    final memoryMaxMB = (namespace.maxMemory / (1024 * 1024)).toInt();
    
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
                  Icons.folder,
                  color: Theme.of(context).colorScheme.primary,
                  size: 28,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        namespace.name,
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      if (namespace.description.isNotEmpty)
                        Text(
                          namespace.description,
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                          ),
                        ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // Resource usage
            _buildResourceUsage(
              context,
              'Memory',
              namespace.memoryUsedMb,
              memoryMaxMB.toDouble(),
              'MB',
              namespace.memoryUsagePercent,
            ),
            const SizedBox(height: 16),

            // Metrics
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                _MetricChip(
                  label: 'QPS',
                  value: namespace.qps.toStringAsFixed(1),
                  icon: Icons.speed,
                  color: Colors.blue,
                ),
                _MetricChip(
                  label: 'Total Keys',
                  value: NumberFormat.compact().format(namespace.keyCount),
                  icon: Icons.key,
                  color: Colors.purple,
                ),
                _MetricChip(
                  label: 'Hit Rate',
                  value: '${(namespace.hitRate * 100).toStringAsFixed(1)}%',
                  icon: Icons.analytics,
                  color: Colors.green,
                ),
                _MetricChip(
                  label: 'Rate Limit',
                  value: '${namespace.rateLimit} QPS',
                  icon: Icons.speed_outlined,
                  color: Colors.orange,
                ),
                _MetricChip(
                  label: 'Default TTL',
                  value: '${namespace.defaultTTL}s',
                  icon: Icons.timer,
                  color: Colors.teal,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildResourceUsage(
    BuildContext context,
    String label,
    double used,
    double max,
    String unit,
    double percent,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(label, style: Theme.of(context).textTheme.bodySmall),
            Text(
              '${used.toStringAsFixed(1)} / ${max.toStringAsFixed(0)} $unit',
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
            value: percent / 100,
            backgroundColor: Colors.grey.withOpacity(0.2),
            valueColor: AlwaysStoppedAnimation<Color>(
              _getColorForUsage(percent),
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
