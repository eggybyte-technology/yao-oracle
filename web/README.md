# Yao-Oracle Dashboard - Web Frontend

## 📁 File Structure

```
web/
├── login.html              # Login page (password authentication)
├── index.html              # Main dashboard page
├── css/
│   ├── common.css          # Shared styles, CSS variables, utilities
│   ├── login.css           # Login page styles
│   ├── dashboard.css       # Dashboard layout and components
│   └── cache-explorer.css  # Cache explorer styles
├── js/
│   ├── config.js           # Configuration constants
│   ├── auth.js             # Authentication logic
│   ├── api.js              # API client wrapper
│   ├── charts.js           # Chart.js management
│   ├── dashboard.js        # Main dashboard logic with mock data
│   └── cache-explorer.js   # Cache data browser
└── assets/                 # (Optional) Images and icons
```

## 🎨 Features

### Login Page (`login.html`)
- Clean, centered login form
- Password-based authentication
- Error message display
- Automatic redirect to dashboard on success
- Responsive design

### Dashboard (`index.html`)
- **Overview Tab**: System-level metrics with 4 interactive charts
  - Key metrics cards (Namespaces, Nodes, Keys, Memory)
  - Cache hit rate gauge chart
  - QPS time series chart
  - Memory distribution doughnut chart
  - Response time line chart
  - Cluster health status table
  
- **Namespaces Tab**: Business namespace statistics
  - Grid layout showing all namespaces
  - Per-namespace metrics (keys, memory, hit rate, QPS)
  - Status indicators

- **Proxies Tab**: Proxy instance monitoring
  - Table view of all proxy instances
  - Real-time QPS, latency, and error metrics
  - Health status indicators

- **Nodes Tab**: Cache node details
  - Card layout for each node
  - Memory usage progress bars
  - Hit/miss statistics
  - Uptime information

- **Cache Explorer Tab**: Browse cache data
  - Namespace selector dropdown
  - Paginated key listing
  - Search/filter by key prefix
  - Click to view key details (value, TTL, metadata)

### UI/UX Features
- Dark/Light theme toggle
- Auto-refresh every 5 seconds
- Manual refresh button
- Responsive design (mobile, tablet, desktop)
- Smooth animations and transitions
- Modern gradient design
- Loading states
- Alert/toast notifications

## 🧪 Testing with Mock Data

The dashboard currently uses **mock data** for testing purposes. This allows you to:
- View the complete UI without a backend
- Test all interactions and features
- Validate the design and user experience

### Test Mode Configuration

In `config.js`:
```javascript
TEST_MODE: true,              // Enable test mode (no backend required)
DEFAULT_PASSWORD: 'admin123', // Default password for testing
```

**Default Login Credentials:**
- Password: `admin123`

When `TEST_MODE` is enabled:
- ✅ A banner displays the default password on the login page
- ✅ Authentication works without a backend
- ✅ All data is simulated with realistic mock data
- ✅ Perfect for frontend development and UI testing

### Enable Mock Data Mode

In `dashboard.js`:
```javascript
useMockData: true, // Set to false when backend is ready
```

### Mock Data Includes:
- 4 namespaces (game-app, ads-service, user-cache, api-cache)
- 3 proxy instances
- 6 cache nodes
- Time series data for charts (QPS, hit rate, latency)
- Health status for all components

### Testing Checklist

1. **Login Page**
   - [ ] Load `/login.html`
   - [ ] Verify test mode banner is displayed with default password
   - [ ] Enter default password: `admin123`
   - [ ] Click "Sign In" button
   - [ ] Should redirect to dashboard (`/index.html`)

2. **Dashboard Overview**
   - [ ] Verify all 4 metric cards display data
   - [ ] Check hit rate gauge chart (should show ~94.7%)
   - [ ] Verify QPS line chart displays trend
   - [ ] Check memory distribution chart
   - [ ] Verify latency chart
   - [ ] Review health status table

3. **Navigation**
   - [ ] Click each tab in sidebar
   - [ ] Verify content changes
   - [ ] Check active state highlighting

4. **Namespaces Tab**
   - [ ] View all 4 namespace cards
   - [ ] Verify metrics display correctly

5. **Proxies Tab**
   - [ ] View proxy table
   - [ ] Check all columns have data

6. **Nodes Tab**
   - [ ] View node cards
   - [ ] Check memory progress bars
   - [ ] Verify statistics

7. **Theme Toggle**
   - [ ] Click theme toggle button
   - [ ] Verify dark/light mode switch
   - [ ] Check charts update colors

8. **Responsive Design**
   - [ ] Test on mobile (< 768px)
   - [ ] Test on tablet (768px - 1024px)
   - [ ] Test on desktop (> 1024px)

9. **Auto-Refresh**
   - [ ] Wait 5 seconds
   - [ ] Verify data refreshes automatically

10. **Manual Refresh**
    - [ ] Click refresh button
    - [ ] See success notification

## 🔌 Connecting to Backend

When the backend is ready:

### Step 1: Disable Test Mode and Mock Data
```javascript
// In config.js
TEST_MODE: false,       // Disable test mode
DEFAULT_PASSWORD: '',   // Clear default password

// In dashboard.js
useMockData: false,     // Use real backend data
```

### Step 2: Configure API Endpoints

Ensure these endpoints are implemented:

```
POST /api/auth/login          - Login with password
POST /api/auth/logout         - Logout

GET  /api/dashboard/cluster-status  - Cluster overview
GET  /api/dashboard/namespaces      - List namespaces
GET  /api/dashboard/proxies         - List proxies
GET  /api/dashboard/nodes           - List nodes

GET  /api/cache/namespaces          - List namespaces for cache explorer
GET  /api/cache/keys                - List cache keys (paginated)
GET  /api/cache/value               - Get key value
```

### Step 3: Verify Authentication

The frontend expects:
- Login endpoint to return: `{ "success": true, "session_id": "..." }`
- Session ID stored in localStorage as `yao-oracle-session`
- All API requests include `Authorization: Bearer <session_id>` header

### Step 4: Test with Real Data

Follow the testing checklist above with real backend data.

## 🎨 Customization

### Changing Colors

Edit CSS variables in `css/common.css`:
```css
:root {
    --color-primary: #667eea;
    --color-success: #48bb78;
    --color-warning: #ed8936;
    --color-error: #f56565;
    /* ... */
}
```

### Adjusting Refresh Interval

Edit in `js/config.js`:
```javascript
const CONFIG = {
    REFRESH_INTERVAL: 5000, // milliseconds (5 seconds)
    // ...
};
```

### Modifying Charts

Chart presets are defined in `js/charts.js`:
- `ChartPresets.lineChart()` - Line charts
- `ChartPresets.gaugeChart()` - Gauge charts
- `ChartPresets.barChart()` - Bar charts
- `ChartPresets.doughnutChart()` - Doughnut charts

## 📦 Deployment

### With Go Backend

The web files are embedded in the Go binary using `embed`:

```go
//go:embed web
var webFS embed.FS

func main() {
    router := gin.Default()
    
    // Serve embedded static files
    webRoot, _ := fs.Sub(webFS, "web")
    router.StaticFS("/", http.FS(webRoot))
    
    // API routes
    api := router.Group("/api")
    // ...
}
```

### Standalone Testing

For local testing without a backend:
```bash
# Using Python
cd web
python3 -m http.server 8080

# Using Node.js
npx serve web -p 8080

# Then open: http://localhost:8080/login.html
```

## 🐛 Troubleshooting

### Charts Not Rendering
- Check browser console for errors
- Verify Chart.js CDN is accessible
- Ensure canvas elements have correct IDs

### Authentication Loop
- Clear localStorage: `localStorage.clear()`
- Check session token validity
- Verify backend /auth/login endpoint

### Data Not Loading
- Check browser console for API errors
- Verify `useMockData` setting in `dashboard.js`
- Check network tab for failed requests

### Styling Issues
- Clear browser cache
- Verify CSS files are loading
- Check for CSS syntax errors

## 📚 Dependencies

- **Chart.js 4.x**: Loaded from CDN in `index.html`
- **No build step required**: Pure HTML/CSS/JS
- **No npm packages needed**: Everything runs in the browser

## 📄 License

Copyright © 2025 Yao-Oracle. All rights reserved.

