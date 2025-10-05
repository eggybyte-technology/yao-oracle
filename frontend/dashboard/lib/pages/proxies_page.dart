/// Proxies monitoring page
library;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/app_state.dart';
import '../widgets/metric_card.dart';
import '../widgets/loading_widget.dart' as loading;

/// Proxies page showing all proxy instances and their metrics
class ProxiesPage extends StatelessWidget {
  const ProxiesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<AppState>(
      builder: (context, appState, child) {
        if (appState.isLoading && appState.proxies.isEmpty) {
          return const loading.LoadingWidget(message: 'Loading proxies...');
        }

        if (appState.error != null) {
          return loading.ErrorWidget(
            message: appState.error!,
            onRetry: () => appState.loadProxies(),
          );
        }

        if (appState.proxies.isEmpty) {
          return const loading.EmptyStateWidget(
            message: 'No proxy instances found',
            icon: Icons.router,
          );
        }

        return RefreshIndicator(
          onRefresh: () => appState.loadProxies(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'Proxy Instances',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 16),
              ...appState.proxies.map((proxy) => _ProxyCard(proxy: proxy)),
            ],
          ),
        );
      },
    );
  }
}

class _ProxyCard extends StatelessWidget {
  final dynamic proxy;

  const _ProxyCard({required this.proxy});

  @override
  Widget build(BuildContext context) {
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
                  Icons.router,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        proxy.id,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      Text(
                        proxy.ip,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
                StatusBadge(status: proxy.status, isHealthy: proxy.isHealthy),
              ],
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                _MetricChip(
                  label: 'QPS',
                  value: proxy.qps.total.toString(),
                  icon: Icons.speed,
                ),
                _MetricChip(
                  label: 'Latency P50',
                  value: '${proxy.latency.p50.toStringAsFixed(2)}ms',
                  icon: Icons.timer,
                ),
                _MetricChip(
                  label: 'Error Rate',
                  value: '${(proxy.errorRate * 100).toStringAsFixed(3)}%',
                  icon: Icons.error_outline,
                ),
                _MetricChip(
                  label: 'Connections',
                  value: proxy.connections.toString(),
                  icon: Icons.link,
                ),
                _MetricChip(
                  label: 'Uptime',
                  value: _formatUptime(proxy.uptime),
                  icon: Icons.access_time,
                ),
              ],
            ),
            if (proxy.namespaces.isNotEmpty) ...[
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: proxy.namespaces
                    .map<Widget>(
                      (ns) => Chip(
                        label: Text(ns),
                        labelStyle: const TextStyle(fontSize: 12),
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                      ),
                    )
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
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

  const _MetricChip({
    required this.label,
    required this.value,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: Theme.of(context).colorScheme.primary),
          const SizedBox(width: 6),
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
                ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.bold),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
