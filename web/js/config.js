/* ======================================================================
   Yao-Oracle Dashboard - Configuration
   ====================================================================== */

const CONFIG = {
    // API Configuration
    API_BASE: '/api',

    // Test Mode (set to false in production)
    TEST_MODE: true,
    DEFAULT_PASSWORD: 'admin123',  // Default password for testing

    // Refresh Intervals (in milliseconds)
    REFRESH_INTERVAL: 5000,
    METRICS_REFRESH_INTERVAL: 3000,

    // Storage Keys
    TOKEN_KEY: 'yao-oracle-session',
    THEME_KEY: 'yao-oracle-theme',

    // UI Configuration
    CHART_COLORS: {
        primary: '#667eea',
        success: '#48bb78',
        warning: '#ed8936',
        error: '#f56565',
        info: '#4299e1',
        gradient: ['#667eea', '#764ba2']
    },

    // Chart Options
    CHART_DEFAULTS: {
        responsive: true,
        maintainAspectRatio: false,
        animation: {
            duration: 750
        }
    },

    // Pagination
    ITEMS_PER_PAGE: 50,

    // Timeouts
    REQUEST_TIMEOUT: 10000,
    SESSION_TIMEOUT: 1800000, // 30 minutes

    // Debug Mode
    DEBUG: false
};

// Log configuration in debug mode
if (CONFIG.DEBUG) {
    console.log('[Config] Configuration loaded:', CONFIG);
}
