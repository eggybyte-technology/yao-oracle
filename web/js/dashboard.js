/* ======================================================================
   Yao-Oracle Dashboard - Main Logic
   ====================================================================== */

/**
 * Mock Data Generator for Testing
 */
const MockData = {
    // Generate mock overview data
    getOverview() {
        return {
            namespaces: 4,
            nodes: 6,
            keys: 125847,
            memory: '2.4 GB / 8 GB',
            memoryPercent: 30,
            requests: 1847362,
            hitRate: 94.7,
            hits: 1749123,
            misses: 98239,
            proxyStatus: 'healthy',
            healthyNodes: 6
        };
    },

    // Generate mock time series data
    getTimeSeries(points = 20) {
        const now = Date.now();
        const labels = [];
        const data = [];

        for (let i = points - 1; i >= 0; i--) {
            const time = new Date(now - i * 60000);
            labels.push(time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }));
            data.push(Math.random() * 50 + 50); // Random value between 50-100
        }

        return { labels, data };
    },

    // Generate mock QPS data
    getQPSData(points = 20) {
        const now = Date.now();
        const labels = [];
        const data = [];

        for (let i = points - 1; i >= 0; i--) {
            const time = new Date(now - i * 60000);
            labels.push(time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }));
            data.push(Math.floor(Math.random() * 5000 + 8000)); // 8000-13000 QPS
        }

        return { labels, data };
    },

    // Generate mock memory distribution data
    getMemoryDistribution() {
        return {
            labels: ['game-app', 'ads-service', 'user-cache', 'api-cache'],
            data: [512, 256, 384, 128]
        };
    },

    // Generate mock latency data
    getLatencyData(points = 20) {
        const now = Date.now();
        const labels = [];
        const data = [];

        for (let i = points - 1; i >= 0; i--) {
            const time = new Date(now - i * 60000);
            labels.push(time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }));
            data.push(Math.random() * 10 + 2); // 2-12ms latency
        }

        return { labels, data };
    },

    // Generate mock health data
    getHealthData() {
        return [
            {
                component: 'Proxy',
                status: 'healthy',
                instances: 3,
                health: '100%',
                uptime: '7d 14h',
                lastCheck: '10s ago'
            },
            {
                component: 'Cache Nodes',
                status: 'healthy',
                instances: 6,
                health: '100%',
                uptime: '7d 14h',
                lastCheck: '10s ago'
            },
            {
                component: 'Dashboard',
                status: 'healthy',
                instances: 2,
                health: '100%',
                uptime: '7d 14h',
                lastCheck: '5s ago'
            }
        ];
    },

    // Generate mock namespace data
    getNamespaces() {
        return [
            {
                name: 'game-app',
                description: 'Gaming application cache',
                keys: 45823,
                memory: '512 MB',
                hitRate: 96.2,
                qps: 5420,
                errorRate: 0.1,
                status: 'healthy'
            },
            {
                name: 'ads-service',
                description: 'Advertisement service cache',
                keys: 28941,
                memory: '256 MB',
                hitRate: 93.8,
                qps: 3210,
                errorRate: 0.2,
                status: 'healthy'
            },
            {
                name: 'user-cache',
                description: 'User profile cache',
                keys: 38765,
                memory: '384 MB',
                hitRate: 94.5,
                qps: 4180,
                errorRate: 0.15,
                status: 'healthy'
            },
            {
                name: 'api-cache',
                description: 'API response cache',
                keys: 12318,
                memory: '128 MB',
                hitRate: 92.1,
                qps: 1890,
                errorRate: 0.3,
                status: 'warning'
            }
        ];
    },

    // Generate mock proxy data
    getProxies() {
        return [
            {
                name: 'proxy-0',
                ip: '10.244.1.23',
                status: 'healthy',
                uptime: '7d 14h 32m',
                qps: 4820,
                latency: '3.2ms',
                errors: 12,
                namespaces: 4
            },
            {
                name: 'proxy-1',
                ip: '10.244.1.24',
                status: 'healthy',
                uptime: '7d 14h 32m',
                qps: 4650,
                latency: '3.5ms',
                errors: 8,
                namespaces: 4
            },
            {
                name: 'proxy-2',
                ip: '10.244.1.25',
                status: 'healthy',
                uptime: '7d 14h 32m',
                qps: 5050,
                latency: '3.1ms',
                errors: 15,
                namespaces: 4
            }
        ];
    },

    // Generate mock node data
    getNodes() {
        return [
            {
                name: 'node-0',
                ip: '10.244.2.10',
                status: 'healthy',
                keys: 21847,
                memory: '412 MB / 1024 MB',
                memoryPercent: 40,
                hits: 324850,
                misses: 18234,
                hitRate: 94.7,
                uptime: '7d 14h 32m'
            },
            {
                name: 'node-1',
                ip: '10.244.2.11',
                status: 'healthy',
                keys: 19523,
                memory: '385 MB / 1024 MB',
                memoryPercent: 38,
                hits: 298451,
                misses: 16892,
                hitRate: 94.6,
                uptime: '7d 14h 32m'
            },
            {
                name: 'node-2',
                ip: '10.244.2.12',
                status: 'healthy',
                keys: 23184,
                memory: '441 MB / 1024 MB',
                memoryPercent: 43,
                hits: 345123,
                misses: 19874,
                hitRate: 94.6,
                uptime: '7d 14h 32m'
            },
            {
                name: 'node-3',
                ip: '10.244.2.13',
                status: 'healthy',
                keys: 20456,
                memory: '398 MB / 1024 MB',
                memoryPercent: 39,
                hits: 312456,
                misses: 17234,
                hitRate: 94.8,
                uptime: '7d 14h 32m'
            },
            {
                name: 'node-4',
                ip: '10.244.2.14',
                status: 'healthy',
                keys: 22198,
                memory: '427 MB / 1024 MB',
                memoryPercent: 42,
                hits: 332851,
                misses: 18675,
                hitRate: 94.7,
                uptime: '7d 14h 32m'
            },
            {
                name: 'node-5',
                ip: '10.244.2.15',
                status: 'healthy',
                keys: 18639,
                memory: '361 MB / 1024 MB',
                memoryPercent: 35,
                hits: 285392,
                misses: 16330,
                hitRate: 94.6,
                uptime: '7d 14h 32m'
            }
        ];
    }
};

/**
 * Dashboard Controller
 */
const Dashboard = {
    refreshInterval: null,
    useMockData: true, // Set to false when backend is ready

    /**
     * Initialize dashboard
     */
    async init() {
        if (CONFIG.DEBUG) {
            console.log('[Dashboard] Initializing...');
        }

        // Check authentication
        if (!this.checkAuth()) {
            return;
        }

        // Setup event listeners
        this.setupEventListeners();

        // Load initial data
        await this.loadOverviewTab();

        // Hide loading state
        document.getElementById('loading-state').style.display = 'none';

        // Start auto-refresh
        this.startAutoRefresh();

        if (CONFIG.DEBUG) {
            console.log('[Dashboard] Initialized successfully');
        }
    },

    /**
     * Check authentication
     */
    checkAuth() {
        const token = localStorage.getItem(CONFIG.TOKEN_KEY);
        if (!token) {
            window.location.href = '/login.html';
            return false;
        }
        return true;
    },

    /**
     * Setup event listeners
     */
    setupEventListeners() {
        // Tab navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const tab = item.dataset.tab;
                this.switchTab(tab);
            });
        });

        // Refresh button
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.refreshCurrentTab();
        });

        // Logout button
        document.getElementById('logout-btn').addEventListener('click', () => {
            localStorage.removeItem(CONFIG.TOKEN_KEY);
            window.location.href = '/login.html';
        });

        // Theme toggle
        document.getElementById('theme-toggle').addEventListener('click', () => {
            this.toggleTheme();
        });
    },

    /**
     * Switch between tabs
     */
    async switchTab(tabName) {
        // Update navigation active state
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update tab panel active state
        document.querySelectorAll('.tab-panel').forEach(panel => {
            panel.classList.remove('active');
        });
        document.getElementById(`tab-${tabName}`).classList.add('active');

        // Load tab data
        await this.loadTabData(tabName);
    },

    /**
     * Load data for specific tab
     */
    async loadTabData(tabName) {
        switch (tabName) {
            case 'overview':
                await this.loadOverviewTab();
                break;
            case 'namespaces':
                await this.loadNamespacesTab();
                break;
            case 'proxies':
                await this.loadProxiesTab();
                break;
            case 'nodes':
                await this.loadNodesTab();
                break;
            case 'cache-explorer':
                // Handled by cache-explorer.js
                break;
        }
    },

    /**
     * Load overview tab
     */
    async loadOverviewTab() {
        const data = this.useMockData ? MockData.getOverview() : await this.fetchOverviewData();

        // Update metrics
        document.getElementById('metric-namespaces').textContent = data.namespaces;
        document.getElementById('metric-nodes').textContent = data.nodes;
        document.getElementById('metric-keys').textContent = this.formatNumber(data.keys);
        document.getElementById('metric-memory').textContent = data.memory;

        // Create charts
        this.createOverviewCharts();

        // Update health table
        this.updateHealthTable();
    },

    /**
     * Create overview charts
     */
    createOverviewCharts() {
        // Hit Rate Gauge
        const hitRateData = this.useMockData ? 94.7 : 0;
        ChartManager.create('chart-hitrate', ChartPresets.gaugeChart(hitRateData, 'Hit Rate'));

        // QPS Line Chart
        const qpsData = this.useMockData ? MockData.getQPSData() : { labels: [], data: [] };
        ChartManager.create('chart-qps', ChartPresets.lineChart(
            qpsData.labels,
            qpsData.data,
            'QPS',
            CONFIG.CHART_COLORS.primary
        ));

        // Memory Distribution
        const memData = this.useMockData ? MockData.getMemoryDistribution() : { labels: [], data: [] };
        ChartManager.create('chart-memory', ChartPresets.doughnutChart(
            memData.labels,
            memData.data,
            [
                CONFIG.CHART_COLORS.primary,
                CONFIG.CHART_COLORS.success,
                CONFIG.CHART_COLORS.warning,
                CONFIG.CHART_COLORS.info
            ]
        ));

        // Latency Chart
        const latencyData = this.useMockData ? MockData.getLatencyData() : { labels: [], data: [] };
        ChartManager.create('chart-latency', ChartPresets.lineChart(
            latencyData.labels,
            latencyData.data,
            'Latency (ms)',
            CONFIG.CHART_COLORS.success
        ));
    },

    /**
     * Update health table
     */
    updateHealthTable() {
        const healthData = this.useMockData ? MockData.getHealthData() : [];
        const tbody = document.getElementById('health-table-body');

        tbody.innerHTML = healthData.map(item => `
            <tr>
                <td>
                    <div class="table-cell-with-icon">
                        <span class="status-indicator status-${item.status}"></span>
                        <span>${item.component}</span>
                    </div>
                </td>
                <td><span class="badge badge-${item.status === 'healthy' ? 'success' : 'warning'}">${item.status}</span></td>
                <td>${item.instances}</td>
                <td>${item.health}</td>
                <td>${item.uptime}</td>
                <td class="text-muted">${item.lastCheck}</td>
            </tr>
        `).join('');
    },

    /**
     * Load namespaces tab
     */
    async loadNamespacesTab() {
        const namespaces = this.useMockData ? MockData.getNamespaces() : await this.fetchNamespaces();
        const container = document.getElementById('namespaces-content');

        container.innerHTML = `
            <div class="grid-2">
                ${namespaces.map(ns => `
                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">${ns.name}</h3>
                            <span class="badge badge-${ns.status === 'healthy' ? 'success' : 'warning'}">${ns.status}</span>
                        </div>
                        <div class="card-body">
                            <p class="text-muted">${ns.description}</p>
                            <div class="stats-grid">
                                <div class="stat-item">
                                    <div class="stat-label">Keys</div>
                                    <div class="stat-value">${this.formatNumber(ns.keys)}</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-label">Memory</div>
                                    <div class="stat-value">${ns.memory}</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-label">Hit Rate</div>
                                    <div class="stat-value text-success">${ns.hitRate}%</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-label">QPS</div>
                                    <div class="stat-value">${this.formatNumber(ns.qps)}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    },

    /**
     * Load proxies tab
     */
    async loadProxiesTab() {
        const proxies = this.useMockData ? MockData.getProxies() : await this.fetchProxies();
        const container = document.getElementById('proxies-content');

        container.innerHTML = `
            <div class="card">
                <div class="card-body">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>IP Address</th>
                                <th>Status</th>
                                <th>Uptime</th>
                                <th>QPS</th>
                                <th>Latency</th>
                                <th>Errors</th>
                                <th>Namespaces</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${proxies.map(proxy => `
                                <tr>
                                    <td class="font-semibold">${proxy.name}</td>
                                    <td class="font-mono text-muted">${proxy.ip}</td>
                                    <td><span class="badge badge-success">${proxy.status}</span></td>
                                    <td>${proxy.uptime}</td>
                                    <td>${this.formatNumber(proxy.qps)}</td>
                                    <td class="text-success">${proxy.latency}</td>
                                    <td>${proxy.errors}</td>
                                    <td>${proxy.namespaces}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    },

    /**
     * Load nodes tab
     */
    async loadNodesTab() {
        const nodes = this.useMockData ? MockData.getNodes() : await this.fetchNodes();
        const container = document.getElementById('nodes-content');

        container.innerHTML = `
            <div class="grid-2">
                ${nodes.map(node => `
                    <div class="card">
                        <div class="card-header">
                            <div>
                                <h3 class="card-title">${node.name}</h3>
                                <p class="text-muted font-mono" style="font-size: 0.875rem;">${node.ip}</p>
                            </div>
                            <span class="badge badge-success">${node.status}</span>
                        </div>
                        <div class="card-body">
                            <div class="progress-bar-container">
                                <div class="progress-bar-header">
                                    <span>Memory Usage</span>
                                    <span>${node.memory}</span>
                                </div>
                                <div class="progress-bar">
                                    <div class="progress-bar-fill" style="width: ${node.memoryPercent}%; background-color: ${node.memoryPercent >= 80 ? CONFIG.CHART_COLORS.error :
                node.memoryPercent >= 60 ? CONFIG.CHART_COLORS.warning :
                    CONFIG.CHART_COLORS.success
            }"></div>
                                </div>
                            </div>
                            <div class="stats-grid">
                                <div class="stat-item">
                                    <div class="stat-label">Keys</div>
                                    <div class="stat-value">${this.formatNumber(node.keys)}</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-label">Hit Rate</div>
                                    <div class="stat-value text-success">${node.hitRate}%</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-label">Hits</div>
                                    <div class="stat-value">${this.formatNumber(node.hits)}</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-label">Misses</div>
                                    <div class="stat-value">${this.formatNumber(node.misses)}</div>
                                </div>
                            </div>
                            <div class="text-muted" style="margin-top: 1rem; font-size: 0.875rem;">
                                Uptime: ${node.uptime}
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    },

    /**
     * Refresh current tab
     */
    async refreshCurrentTab() {
        const activeTab = document.querySelector('.nav-item.active');
        if (activeTab) {
            const tabName = activeTab.dataset.tab;
            await this.loadTabData(tabName);
            this.showAlert('Data refreshed successfully', 'success');
        }
    },

    /**
     * Start auto-refresh
     */
    startAutoRefresh() {
        this.refreshInterval = setInterval(() => {
            const activeTab = document.querySelector('.nav-item.active');
            if (activeTab) {
                const tabName = activeTab.dataset.tab;
                this.loadTabData(tabName);
            }
        }, CONFIG.REFRESH_INTERVAL);
    },

    /**
     * Stop auto-refresh
     */
    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    },

    /**
     * Toggle theme
     */
    toggleTheme() {
        const html = document.documentElement;
        const currentTheme = html.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

        html.setAttribute('data-theme', newTheme);
        localStorage.setItem(CONFIG.THEME_KEY, newTheme);

        // Update theme icons
        const sunIcon = document.querySelector('.icon-sun');
        const moonIcon = document.querySelector('.icon-moon');

        if (newTheme === 'dark') {
            sunIcon.style.display = 'block';
            moonIcon.style.display = 'none';
        } else {
            sunIcon.style.display = 'none';
            moonIcon.style.display = 'block';
        }

        // Recreate charts to update colors
        this.refreshCurrentTab();
    },

    /**
     * Show alert message
     */
    showAlert(message, type = 'info') {
        const container = document.getElementById('alert-container');
        const alert = document.createElement('div');
        alert.className = `alert alert-${type}`;
        alert.textContent = message;

        container.appendChild(alert);

        setTimeout(() => {
            alert.remove();
        }, 5000);
    },

    /**
     * Format number with commas
     */
    formatNumber(num) {
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    },

    /**
     * Fetch methods (to be implemented when backend is ready)
     */
    async fetchOverviewData() {
        // TODO: Implement API call
        return MockData.getOverview();
    },

    async fetchNamespaces() {
        // TODO: Implement API call
        return MockData.getNamespaces();
    },

    async fetchProxies() {
        // TODO: Implement API call
        return MockData.getProxies();
    },

    async fetchNodes() {
        // TODO: Implement API call
        return MockData.getNodes();
    }
};

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    Dashboard.init();
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    Dashboard.stopAutoRefresh();
    ChartManager.destroyAll();
});
