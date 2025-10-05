# Yao-Oracle Dashboard (Flutter Web)

A modern, responsive web dashboard for monitoring the Yao-Oracle distributed KV cache system.

## Features

✨ **Modern UI**: Built with Flutter Web and Material Design 3
🔄 **Real-time Updates**: WebSocket integration for live metrics
📊 **Rich Visualizations**: Comprehensive metrics and statistics
📱 **Responsive Design**: Works on mobile, tablet, and desktop
🎨 **Dark Mode**: Automatic theme switching based on system preference

## Pages

- **Overview**: Cluster health, component status, and key metrics
- **Proxies**: Proxy instance monitoring with QPS, latency, and error rates
- **Nodes**: Cache node metrics including memory usage and hot keys
- **Namespaces**: Business namespace statistics and resource usage

## Prerequisites

- Flutter SDK 3.9.2 or higher
- Dart SDK (included with Flutter)
- Modern web browser (Chrome, Firefox, Safari, or Edge)

## Installation

```bash
# Navigate to dashboard directory
cd frontend/dashboard

# Install dependencies
flutter pub get
```

## Running the Dashboard

### Option 1: Integrated Testing (Recommended)

Use the provided script to start both mock-admin and dashboard:

```bash
# From project root
./scripts/test-dashboard.sh
```

This will:
1. Start the mock admin service on port 8081
2. Start the Flutter dashboard on port 8080
3. Open your browser automatically

### Option 2: Development Mode

Run dashboard separately (requires mock-admin running):

```bash
# Terminal 1: Start mock-admin
./scripts/run-mock-admin.sh

# Terminal 2: Start dashboard
./scripts/run-dashboard-dev.sh
```

### Option 3: Manual Flutter Run

```bash
cd frontend/dashboard

# Run with custom API endpoint
flutter run -d web-server --web-port 8080 \
  --dart-define=API_URL=http://localhost:8081/api \
  --dart-define=WS_URL=ws://localhost:8081/ws
```

## Configuration

The dashboard can be configured via Dart defines:

```bash
--dart-define=API_URL=http://your-admin-service:8081/api
--dart-define=WS_URL=ws://your-admin-service:8081/ws
```

## Building for Production

```bash
# Build web app
flutter build web --release \
  --dart-define=API_URL=https://your-production-admin/api \
  --dart-define=WS_URL=wss://your-production-admin/ws

# Output will be in build/web/
```

Serve the built files using any static file server (Nginx, Apache, or cloud storage).

## Development

### Project Structure

```
lib/
├── core/              # Core functionality
│   ├── api_client.dart    # HTTP & WebSocket client
│   └── app_state.dart     # Global state management
├── models/            # Data models
│   └── metrics.dart       # Metrics data structures
├── pages/             # Page widgets
│   ├── overview_page.dart
│   ├── proxies_page.dart
│   ├── nodes_page.dart
│   └── namespaces_page.dart
├── widgets/           # Reusable components
│   ├── metric_card.dart
│   └── loading_widget.dart
└── main.dart          # App entry point
```

### Code Style

- Follow Flutter/Dart style guide
- Use meaningful variable names
- Document public APIs
- Keep widgets focused and composable

### Testing

```bash
# Run unit tests
flutter test

# Run with coverage
flutter test --coverage

# Analyze code
flutter analyze
```

## Deployment

### Docker

```dockerfile
FROM nginx:alpine
COPY build/web /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Kubernetes

See `helm/yao-oracle/templates/dashboard/` for deployment manifests.

## Mock Admin API

The dashboard requires an admin service providing these endpoints:

- `GET /api/health` - Health check
- `GET /api/overview` - Cluster overview
- `GET /api/cluster/timeseries` - Cluster time-series data
- `GET /api/proxies` - List all proxies
- `GET /api/proxies/:id` - Proxy details
- `GET /api/proxies/:id/timeseries` - Proxy metrics over time
- `GET /api/nodes` - List all cache nodes
- `GET /api/nodes/:id` - Node details
- `GET /api/nodes/:id/timeseries` - Node metrics over time
- `GET /api/namespaces` - List all namespaces
- `GET /api/namespaces/:name` - Namespace details
- `GET /api/cache` - Query cache entries (with pagination)
- `WS /ws` - WebSocket for real-time updates

## Troubleshooting

### Port Already in Use

```bash
# Find and kill process using port 8080
lsof -ti:8080 | xargs kill -9
```

### WebSocket Connection Failed

- Ensure mock-admin is running
- Check browser console for error messages
- Verify WebSocket URL is correct

### Build Errors

```bash
# Clean and rebuild
flutter clean
flutter pub get
flutter run
```

## Contributing

1. Follow the project's code style
2. Write tests for new features
3. Update documentation
4. Submit pull request

## License

See project root LICENSE file.

## Related Documentation

- [Project Structure](../../.cursor/rules/project-structure.mdc)
- [Admin Service](../../.cursor/rules/admin.mdc)
- [Infrastructure](../../.cursor/rules/infrastructure.mdc)
