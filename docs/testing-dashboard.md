# Dashboard Testing Guide

This guide explains how to test the Yao-Oracle Dashboard using the mock Admin service.

## Quick Start

Run the one-click test script:

```bash
# Using the script directly
bash scripts/test-dashboard.sh

# Or using Make
make test-dashboard
```

This script will:
1. Check dependencies (Go, Node.js, npm)
2. **Automatically create/update** `.env.development` with correct configuration
3. Install npm packages if needed
4. Start Mock Admin service on port 8081
5. Start Dashboard dev server on port 5173
6. Display access URLs and test commands

**Note**: The script will automatically configure the WebSocket URL to `ws://localhost:8081/ws` (not `/api/ws`)

## Architecture

```
┌─────────────────┐          ┌──────────────────┐
│   Dashboard     │◄────────►│   Mock Admin     │
│  (React/Vite)   │   REST   │   (Go/Gin)       │
│  localhost:5173 │          │  localhost:8081  │
└─────────────────┘          └──────────────────┘
         ▲                            │
         │        WebSocket           │
         └────────────────────────────┘
            Real-time updates
```

### Components

**Dashboard (Frontend)**
- Technology: React + TypeScript + Vite
- Port: 5173
- API Client: `dashboard/src/api/client.ts`
- WebSocket: `dashboard/src/api/websocket.ts`
- Environment: `dashboard/.env.development`

**Mock Admin (Backend)**
- Technology: Go + Gin
- Port: 8081
- Source: `cmd/mock-admin/main.go`
- Features:
  - REST API for metrics queries
  - WebSocket for real-time updates (2s interval)
  - Dynamic mock data generation
  - CORS enabled for localhost

## API Endpoints Mapping

### REST API (Mock Admin → Dashboard)

| Dashboard Client Function | Mock Admin Endpoint | Description |
|---------------------------|---------------------|-------------|
| `fetchHealth()` | `GET /api/health` | Health check |
| `fetchOverview()` | `GET /api/overview` | Cluster overview metrics |
| `fetchClusterTimeseries()` | `GET /api/cluster/timeseries` | Cluster-wide time-series data |
| `fetchProxies()` | `GET /api/proxies` | List all proxy instances |
| `fetchProxyDetails(id)` | `GET /api/proxies/:id` | Get proxy details by ID |
| `fetchProxyTimeseries(id)` | `GET /api/proxies/:id/timeseries` | Proxy time-series data |
| `fetchNodes()` | `GET /api/nodes` | List all cache nodes |
| `fetchNodeDetails(id)` | `GET /api/nodes/:id` | Get node details by ID |
| `fetchNodeTimeseries(id)` | `GET /api/nodes/:id/timeseries` | Node time-series data |
| `fetchNamespaces()` | `GET /api/namespaces` | List all namespaces |
| `fetchNamespaceDetails(name)` | `GET /api/namespaces/:name` | Get namespace details by name |
| `fetchCacheEntries(params)` | `GET /api/cache` | Query cache entries with filters |

### WebSocket (Real-time Updates)

**Connection:**
- URL: `ws://localhost:8081/ws`
- Client: `dashboard/src/api/websocket.ts`
- Auto-reconnect: Yes (exponential backoff)

**Message Types:**
```typescript
{
  "type": "overview_update",
  "data": {
    // ClusterMetrics object
  }
}
```

**Update Frequency:**
- Mock Admin broadcasts updates every 2 seconds
- Dashboard receives and updates UI in real-time

## Environment Configuration

### Dashboard `.env.development`

**Automatically created/updated** by `test-dashboard.sh` script:

```bash
VITE_ADMIN_URL=http://localhost:8081/api
VITE_ADMIN_WS_URL=ws://localhost:8081/ws
VITE_APP_TITLE=Yao-Oracle Dashboard (Dev)
VITE_LOG_LEVEL=debug
```

### Mock Admin Environment

```bash
PORT=8081  # HTTP server port (default)
```

## Testing Checklist

### ✅ REST API Tests

Test each endpoint with curl:

```bash
# Health check
curl http://localhost:8081/api/health

# Overview
curl http://localhost:8081/api/overview | jq .

# Cluster timeseries
curl http://localhost:8081/api/cluster/timeseries | jq .

# Proxies
curl http://localhost:8081/api/proxies | jq .
curl http://localhost:8081/api/proxies/proxy-0 | jq .
curl http://localhost:8081/api/proxies/proxy-0/timeseries | jq .

# Nodes
curl http://localhost:8081/api/nodes | jq .
curl http://localhost:8081/api/nodes/node-0 | jq .
curl http://localhost:8081/api/nodes/node-0/timeseries | jq .

# Namespaces
curl http://localhost:8081/api/namespaces | jq .
curl http://localhost:8081/api/namespaces/game-app | jq .

# Cache query
curl 'http://localhost:8081/api/cache?namespace=game-app' | jq .
curl 'http://localhost:8081/api/cache?namespace=game-app&key=user:' | jq .
curl 'http://localhost:8081/api/cache?page=2&page_size=10' | jq .
```

### ✅ Dashboard UI Tests

Open http://localhost:5173 and verify:

**Overview Page:**
- [ ] Cluster metrics displayed (QPS, Latency, Hit Rate, Memory)
- [ ] Gauges show correct percentages
- [ ] Time-series chart displays data
- [ ] WebSocket connection status shown
- [ ] Metrics update in real-time (every 2s)

**Proxies Page:**
- [ ] Proxy instances list displayed
- [ ] Instance cards show status, QPS, latency
- [ ] Click instance to view details
- [ ] QPS breakdown chart visible
- [ ] Latency chart visible

**Nodes Page:**
- [ ] Node instances list displayed
- [ ] Node cards show status, memory, hit rate
- [ ] Click node to view details
- [ ] Memory usage chart visible
- [ ] Hot keys table visible

**Namespaces Page:**
- [ ] Namespace list displayed
- [ ] Click namespace to view details
- [ ] Metrics displayed correctly

**Cache Query Page:**
- [ ] Filter by namespace works
- [ ] Search by key works
- [ ] Pagination works
- [ ] Entry details displayed correctly

### ✅ WebSocket Tests

Test WebSocket connection:

```bash
# Using websocat (install: brew install websocat)
websocat ws://localhost:8081/ws

# You should see real-time updates every 2 seconds:
# {"type":"overview_update","data":{...}}
```

In Dashboard:
- [ ] Connection status shows "Connected" (green)
- [ ] Metrics update without page refresh
- [ ] Disconnection triggers reconnect automatically

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 8081
lsof -i :8081

# Kill the process
kill <PID>

# Or use script's auto-check (will fail if port is busy)
bash scripts/test-dashboard.sh
```

### Dashboard Not Loading

```bash
# Check if vite server is running
lsof -i :5173

# Check dashboard logs
tail -f /tmp/yao-dashboard.log

# Manually start dashboard
cd dashboard
npm install
npm run dev
```

### API Connection Failed

```bash
# Check if mock-admin is running
curl http://localhost:8081/api/health

# Check mock-admin logs
tail -f /tmp/yao-mock-admin.log

# Manually start mock-admin
go run cmd/mock-admin/main.go
```

### CORS Errors

Mock Admin allows requests from:
- http://localhost:5173 (Vite default)
- http://localhost:3000 (Alternative)

If using different port, update `cmd/mock-admin/main.go`:
```go
AllowOrigins: []string{"http://localhost:5173", "http://localhost:YOUR_PORT"},
```

## Manual Testing (Without Script)

### Terminal 1: Start Mock Admin

```bash
cd /path/to/yao-oracle
go run cmd/mock-admin/main.go
```

Wait for:
```
[SUCCESS] 🚀 Mock Admin Service Started Successfully!
```

### Terminal 2: Start Dashboard

```bash
cd /path/to/yao-oracle/dashboard

# Create .env.development if not exists
cat > .env.development << EOF
VITE_ADMIN_URL=http://localhost:8081/api
VITE_ADMIN_WS_URL=ws://localhost:8081/ws
VITE_APP_TITLE=Yao-Oracle Dashboard (Dev)
VITE_LOG_LEVEL=debug
EOF

# Install dependencies
npm install

# Start dev server
npm run dev
```

### Terminal 3: Test APIs

```bash
# Run test commands
curl http://localhost:8081/api/health
curl http://localhost:8081/api/overview | jq .
```

## Mock Data

Mock Admin generates realistic data:

**Proxies:**
- 2 instances: proxy-0, proxy-1
- QPS: 500-700 per proxy
- Latency: 1-8ms
- Status: healthy

**Nodes:**
- 3 instances: node-0, node-1, node-2
- Memory: 700-900MB used / 1024MB max
- Hit Rate: 85-95%
- Status: healthy

**Namespaces:**
- game-app (512MB, 100k keys, 100 QPS)
- ads-service (256MB, 50k keys, 50 QPS)
- user-api (512MB, 100k keys, 80 QPS)

**Cache Entries:**
- 100 entries with realistic keys and values
- Key prefixes: user:, session:, config:, leaderboard:, ad:, profile:, stats:
- JSON values matching key types

**Time-Series:**
- 12 data points (last hour, 5-minute intervals)
- Dynamic values with realistic fluctuations

## Next Steps

After verifying the dashboard with mock data:

1. **Implement Real Admin Service:**
   - Follow same API contract
   - Replace mock data with real metrics collection
   - Implement same WebSocket message format

2. **Integration Testing:**
   - Connect dashboard to real Admin service
   - Verify metrics accuracy
   - Test with real cluster data

3. **Performance Testing:**
   - Test with many proxies/nodes
   - Test WebSocket with many clients
   - Monitor memory usage

4. **Production Deployment:**
   - Build dashboard: `npm run build`
   - Deploy as static site (Nginx)
   - Configure environment variables for production

## Troubleshooting

### Dashboard Shows "Loading..." Forever

**Symptoms:**
- Dashboard shows "Loading cluster overview..." indefinitely
- WebSocket connection errors in console
- No data displayed

**Fixes Applied (2025-10-04):**

1. **Fixed WebSocket URL Configuration**
   - Changed from `ws://localhost:8081/api/ws` to `ws://localhost:8081/ws`
   - The test script now automatically creates correct `.env.development`
   - Solution: Always run `make test-dashboard` to ensure correct configuration

2. **Improved Loading State Handling**
   - Added proper loading state checks in Overview component
   - Now shows clear error messages when data loading fails
   - Dashboard will retry loading automatically every 5 seconds

3. **Enhanced WebSocket Error Handling**
   - Reduced excessive error logging during connection attempts
   - Improved reconnection logic with exponential backoff
   - Clear success indicator when connected: `[WebSocket] ✅ Connected successfully`

**How to Verify Fix:**
```bash
# 1. Stop any running services
# 2. Run the test script
make test-dashboard

# 3. Check browser console - you should see:
# [WebSocket] Connecting to ws://localhost:8081/ws
# [WebSocket] ✅ Connected successfully

# 4. Overview page should load data within 1-2 seconds
```

### WebSocket Connection Errors

**Error:** `WebSocket connection to 'ws://localhost:8081/ws' failed`

**Common Causes:**
1. Mock Admin not running
2. Wrong WebSocket URL in `.env.development`
3. Port 8081 already in use

**Solutions:**
```bash
# Check if Mock Admin is running
curl http://localhost:8081/api/health

# Check port availability
lsof -i :8081

# Recreate environment file
cd dashboard
rm .env.development
cd ..
make test-dashboard
```

### Port Already in Use

**Error:** `Port 8081 is already in use`

**Solution:**
```bash
# Find and kill the process using port 8081
lsof -i :8081
kill <PID>

# Or use a different port (requires manual setup)
PORT=8082 go run cmd/mock-admin/main.go
```

### npm Install Fails

**Error:** `npm install` hangs or fails

**Solution:**
```bash
cd dashboard
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

### CORS Errors

**Error:** `Access-Control-Allow-Origin` errors in browser console

**Note:** Mock Admin has CORS properly configured for `localhost:5173` and `localhost:3000`. If you see CORS errors:

1. Check if you're accessing from the correct port
2. Verify Mock Admin is running
3. Check browser DevTools Network tab for the actual request

## Resources

- Dashboard source: `dashboard/src/`
- Mock Admin source: `cmd/mock-admin/main.go`
- API client: `dashboard/src/api/client.ts`
- WebSocket client: `dashboard/src/api/websocket.ts`
- Test script: `scripts/test-dashboard.sh`
