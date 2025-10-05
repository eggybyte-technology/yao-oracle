/* ======================================================================
   Yao-Oracle Dashboard - API Client
   ====================================================================== */

const API = {
    /**
     * Make authenticated API request
     * @param {string} endpoint - API endpoint
     * @param {Object} options - Fetch options
     * @returns {Promise<any>} Response data
     */
    async request(endpoint, options = {}) {
        const url = `${CONFIG.API_BASE}${endpoint}`;
        const token = Auth.getToken();

        const config = {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
                ...options.headers
            },
            signal: AbortSignal.timeout(CONFIG.REQUEST_TIMEOUT)
        };

        try {
            const response = await fetch(url, config);

            // Handle 401 Unauthorized
            if (response.status === 401) {
                Auth.logout();
                throw new Error('Session expired. Please login again.');
            }

            // Handle other errors
            if (!response.ok) {
                const error = await response.json().catch(() => ({}));
                throw new Error(error.message || `HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            throw error;
        }
    },

    /**
     * Fetch cluster status
     * @returns {Promise<Object>} Cluster status data
     */
    async fetchClusterStatus() {
        return this.request('/dashboard/cluster-status');
    },

    /**
     * Fetch namespaces
     * @returns {Promise<Object>} Namespaces data
     */
    async fetchNamespaces() {
        return this.request('/dashboard/namespaces');
    },

    /**
     * Fetch proxies
     * @returns {Promise<Object>} Proxies data
     */
    async fetchProxies() {
        return this.request('/dashboard/proxies');
    },

    /**
     * Fetch nodes
     * @returns {Promise<Object>} Nodes data
     */
    async fetchNodes() {
        return this.request('/dashboard/nodes');
    },

    /**
     * Fetch namespace stats
     * @param {string} namespace - Namespace name
     * @returns {Promise<Object>} Namespace stats
     */
    async fetchNamespaceStats(namespace) {
        return this.request(`/dashboard/namespace/${encodeURIComponent(namespace)}/stats`);
    },

    /**
     * Fetch node stats
     * @param {string} nodeId - Node ID
     * @returns {Promise<Object>} Node stats
     */
    async fetchNodeStats(nodeId) {
        return this.request(`/dashboard/node/${encodeURIComponent(nodeId)}/stats`);
    },

    /**
     * Fetch cache keys for a namespace
     * @param {string} namespace - Namespace name
     * @param {string} apikey - API key
     * @param {Object} params - Query parameters
     * @returns {Promise<Object>} Cache keys data
     */
    async fetchCacheKeys(namespace, apikey, params = {}) {
        const queryParams = new URLSearchParams({
            namespace,
            apikey,
            page: params.page || 1,
            limit: params.limit || CONFIG.ITEMS_PER_PAGE,
            ...(params.prefix && { prefix: params.prefix })
        });

        return this.request(`/cache/keys?${queryParams}`);
    },

    /**
     * Fetch cache key value
     * @param {string} namespace - Namespace name
     * @param {string} apikey - API key
     * @param {string} key - Cache key
     * @returns {Promise<Object>} Key value data
     */
    async fetchCacheValue(namespace, apikey, key) {
        const queryParams = new URLSearchParams({
            namespace,
            apikey,
            key
        });

        return this.request(`/cache/value?${queryParams}`);
    }
};
