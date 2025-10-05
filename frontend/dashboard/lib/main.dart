/// Yao-Oracle Dashboard - Main entry point
library;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'core/grpc_client.dart';
import 'core/app_state.dart';
import 'pages/overview_page.dart';
import 'pages/metrics_page.dart';
import 'pages/proxies_page.dart';
import 'pages/nodes_page.dart';
import 'pages/namespaces_page.dart';

void main() {
  runApp(const YaoOracleApp());
}

class YaoOracleApp extends StatelessWidget {
  const YaoOracleApp({super.key});

  @override
  Widget build(BuildContext context) {
    // gRPC configuration (can be overridden via environment variables)
    const grpcHost = String.fromEnvironment(
      'GRPC_HOST',
      defaultValue: 'localhost',
    );
    const grpcPortStr = String.fromEnvironment(
      'GRPC_PORT',
      defaultValue: '9090',
    );
    final grpcPort = int.tryParse(grpcPortStr) ?? 9090;

    final grpcClient = GrpcClient(host: grpcHost, port: grpcPort);

    return ChangeNotifierProvider(
      create: (_) => AppState(grpcClient: grpcClient)..loadAll(),
      child: MaterialApp(
        title: 'Yao-Oracle Dashboard',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(
            seedColor: Colors.blue,
            brightness: Brightness.light,
          ),
          useMaterial3: true,
          cardTheme: const CardThemeData(
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.all(Radius.circular(12)),
            ),
          ),
        ),
        darkTheme: ThemeData(
          colorScheme: ColorScheme.fromSeed(
            seedColor: Colors.blue,
            brightness: Brightness.dark,
          ),
          useMaterial3: true,
          cardTheme: const CardThemeData(
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.all(Radius.circular(12)),
            ),
          ),
        ),
        themeMode: ThemeMode.system,
        home: const DashboardHome(),
      ),
    );
  }
}

class DashboardHome extends StatefulWidget {
  const DashboardHome({super.key});

  @override
  State<DashboardHome> createState() => _DashboardHomeState();
}

class _DashboardHomeState extends State<DashboardHome> {
  int _selectedIndex = 0;

  final List<_NavigationItem> _navItems = [
    _NavigationItem(
      icon: Icons.dashboard,
      label: 'Overview',
      page: const OverviewPage(),
    ),
    _NavigationItem(
      icon: Icons.show_chart,
      label: 'Metrics',
      page: const MetricsPage(),
    ),
    _NavigationItem(
      icon: Icons.router,
      label: 'Proxies',
      page: const ProxiesPage(),
    ),
    _NavigationItem(
      icon: Icons.storage,
      label: 'Nodes',
      page: const NodesPage(),
    ),
    _NavigationItem(
      icon: Icons.folder,
      label: 'Namespaces',
      page: const NamespacesPage(),
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            // Logo from external URL
            Image.network(
              'https://yao-verse.eggybyte.com/favicon.png',
              height: 32,
              width: 32,
              errorBuilder: (context, error, stackTrace) {
                return Icon(
                  Icons.cloud,
                  size: 32,
                  color: Theme.of(context).colorScheme.primary,
                );
              },
            ),
            const SizedBox(width: 12),
            const Flexible(
              child: Text(
                'Yao-Oracle Dashboard',
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const Spacer(),
            Consumer<AppState>(
              builder: (context, appState, child) {
                return Flexible(
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (appState.isStreamConnected)
                        Container(
                          width: 8,
                          height: 8,
                          decoration: const BoxDecoration(
                            color: Colors.green,
                            shape: BoxShape.circle,
                          ),
                        ),
                      if (!appState.isStreamConnected)
                        Container(
                          width: 8,
                          height: 8,
                          decoration: const BoxDecoration(
                            color: Colors.grey,
                            shape: BoxShape.circle,
                          ),
                        ),
                      const SizedBox(width: 8),
                      Text(
                        appState.isStreamConnected ? 'Live' : 'Offline',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                );
              },
            ),
            const SizedBox(width: 16),
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () {
                context.read<AppState>().loadAll();
              },
              tooltip: 'Refresh',
            ),
          ],
        ),
      ),
      body: Row(
        children: [
          // Navigation rail for desktop/tablet
          if (MediaQuery.of(context).size.width >= 640)
            NavigationRail(
              selectedIndex: _selectedIndex,
              onDestinationSelected: (index) {
                setState(() {
                  _selectedIndex = index;
                });
              },
              labelType: NavigationRailLabelType.all,
              destinations: _navItems
                  .map(
                    (item) => NavigationRailDestination(
                      icon: Icon(item.icon),
                      label: Text(item.label),
                    ),
                  )
                  .toList(),
            ),

          // Content area
          Expanded(child: _navItems[_selectedIndex].page),
        ],
      ),
      // Bottom navigation bar for mobile
      bottomNavigationBar: MediaQuery.of(context).size.width < 640
          ? NavigationBar(
              selectedIndex: _selectedIndex,
              onDestinationSelected: (index) {
                setState(() {
                  _selectedIndex = index;
                });
              },
              destinations: _navItems
                  .map(
                    (item) => NavigationDestination(
                      icon: Icon(item.icon),
                      label: item.label,
                    ),
                  )
                  .toList(),
            )
          : null,
    );
  }
}

class _NavigationItem {
  final IconData icon;
  final String label;
  final Widget page;

  _NavigationItem({
    required this.icon,
    required this.label,
    required this.page,
  });
}
