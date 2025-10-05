// cache-explorer.js - Cache data browsing and querying functionality
//
// This module provides:
// - Namespace listing and selection
// - Cache key listing with pagination
// - Key value querying and display
// - Search and filtering capabilities

/**
 * CacheExplorer manages the cache data browsing interface.
 */
class CacheExplorer {
    constructor() {
        this.currentNamespace = null;
        this.currentApikey = null;
        this.currentPage = 1;
        this.keysPerPage = 50;
        this.totalKeys = 0;
        this.searchPrefix = '';
        this.namespaces = [];
    }

    /**
     * init initializes the cache explorer UI and event listeners.
     */
    async init() {
        console.log('[CacheExplorer] Initializing...');

        // Load namespaces
        await this.loadNamespaces();

        // Setup event listeners
        this.setupEventListeners();

        console.log('[CacheExplorer] Initialized successfully');
    }

    /**
     * setupEventListeners registers event handlers for UI interactions.
     */
    setupEventListeners() {
        // Namespace selector
        const namespaceSelect = document.getElementById('namespace-select');
        if (namespaceSelect) {
            namespaceSelect.addEventListener('change', (e) => {
                this.selectNamespace(e.target.value);
            });
        }

        // Search input
        const searchInput = document.getElementById('key-search-input');
        if (searchInput) {
            // Debounce search to avoid too many requests
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    this.searchPrefix = e.target.value;
                    this.currentPage = 1;
                    this.loadCacheKeys();
                }, 500);
            });
        }

        // Pagination buttons
        const prevBtn = document.getElementById('pagination-prev');
        const nextBtn = document.getElementById('pagination-next');

        if (prevBtn) {
            prevBtn.addEventListener('click', () => {
                if (this.currentPage > 1) {
                    this.currentPage--;
                    this.loadCacheKeys();
                }
            });
        }

        if (nextBtn) {
            nextBtn.addEventListener('click', () => {
                const totalPages = Math.ceil(this.totalKeys / this.keysPerPage);
                if (this.currentPage < totalPages) {
                    this.currentPage++;
                    this.loadCacheKeys();
                }
            });
        }

        // Refresh button
        const refreshBtn = document.getElementById('cache-refresh-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadCacheKeys();
            });
        }
    }

    /**
     * loadNamespaces fetches the list of configured namespaces.
     */
    async loadNamespaces() {
        try {
            const data = await apiRequest('/cache/namespaces');

            if (!data || !data.namespaces) {
                showError('Failed to load namespaces');
                return;
            }

            this.namespaces = data.namespaces;
            this.renderNamespaceSelector();

        } catch (error) {
            console.error('[CacheExplorer] Error loading namespaces:', error);
            showError('Failed to load namespaces: ' + error.message);
        }
    }

    /**
     * renderNamespaceSelector populates the namespace dropdown.
     */
    renderNamespaceSelector() {
        const select = document.getElementById('namespace-select');
        if (!select) return;

        // Clear existing options
        select.innerHTML = '<option value="">Select a namespace...</option>';

        // Add namespace options
        this.namespaces.forEach(ns => {
            const option = document.createElement('option');
            option.value = ns.name;
            option.textContent = `${ns.name} - ${ns.description} (${formatNumber(ns.key_count || 0)} keys)`;
            option.dataset.apikey = ns.apikey;
            option.dataset.keyCount = ns.key_count || 0;
            option.dataset.memoryMb = ns.memory_mb || 0;
            select.appendChild(option);
        });
    }

    /**
     * selectNamespace handles namespace selection change.
     * 
     * @param {string} namespaceName - Selected namespace name
     */
    async selectNamespace(namespaceName) {
        if (!namespaceName) {
            this.currentNamespace = null;
            this.currentApikey = null;
            this.clearKeyList();
            return;
        }

        const select = document.getElementById('namespace-select');
        const selectedOption = select.options[select.selectedIndex];

        this.currentNamespace = namespaceName;
        this.currentApikey = selectedOption.dataset.apikey;
        this.currentPage = 1;
        this.searchPrefix = '';

        // Clear search input
        const searchInput = document.getElementById('key-search-input');
        if (searchInput) {
            searchInput.value = '';
        }

        // Load keys for this namespace
        await this.loadCacheKeys();
    }

    /**
     * loadCacheKeys fetches cache keys for the selected namespace.
     */
    async loadCacheKeys() {
        if (!this.currentNamespace || !this.currentApikey) {
            return;
        }

        // Show loading state
        this.showLoadingState();

        try {
            const params = new URLSearchParams({
                namespace: this.currentNamespace,
                apikey: this.currentApikey,
                page: this.currentPage.toString(),
                limit: this.keysPerPage.toString()
            });

            if (this.searchPrefix) {
                params.append('prefix', this.searchPrefix);
            }

            const data = await apiRequest(`/cache/keys?${params.toString()}`);

            if (!data) {
                showError('Failed to load cache keys');
                return;
            }

            this.totalKeys = data.total || 0;
            this.renderKeyList(data.keys || []);
            this.renderPagination(data.page || 1, data.total_pages || 1);

        } catch (error) {
            console.error('[CacheExplorer] Error loading cache keys:', error);
            showError('Failed to load cache keys: ' + error.message);
            this.clearKeyList();
        }
    }

    /**
     * renderKeyList displays the list of cache keys.
     * 
     * @param {array} keys - Array of key objects
     */
    renderKeyList(keys) {
        const tbody = document.getElementById('key-list-tbody');
        if (!tbody) return;

        if (keys.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="4" class="text-center">
                        No keys found${this.searchPrefix ? ` matching "${this.searchPrefix}"` : ''}
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = keys.map(key => {
            const ttlDisplay = key.ttl === -1 ? '∞' : `${key.ttl}s`;
            const lastModified = key.last_modified ?
                formatTimeAgo(new Date(key.last_modified)) : 'N/A';

            return `
                <tr class="key-row" data-key="${escapeHtml(key.key)}">
                    <td class="key-name" title="${escapeHtml(key.key)}">
                        ${escapeHtml(key.key)}
                    </td>
                    <td>${ttlDisplay}</td>
                    <td>${formatBytes(key.size_bytes || 0)}</td>
                    <td>${lastModified}</td>
                </tr>
            `;
        }).join('');

        // Add click handlers to rows
        tbody.querySelectorAll('.key-row').forEach(row => {
            row.addEventListener('click', () => {
                const key = row.dataset.key;
                this.queryKeyValue(key);
            });
        });
    }

    /**
     * renderPagination updates pagination controls.
     * 
     * @param {number} currentPage - Current page number
     * @param {number} totalPages - Total number of pages
     */
    renderPagination(currentPage, totalPages) {
        const pageInfo = document.getElementById('pagination-info');
        const prevBtn = document.getElementById('pagination-prev');
        const nextBtn = document.getElementById('pagination-next');

        if (pageInfo) {
            pageInfo.textContent = `Page ${currentPage} of ${totalPages}`;
        }

        if (prevBtn) {
            prevBtn.disabled = currentPage <= 1;
        }

        if (nextBtn) {
            nextBtn.disabled = currentPage >= totalPages;
        }
    }

    /**
     * queryKeyValue fetches and displays the value for a specific key.
     * 
     * @param {string} key - Cache key to query
     */
    async queryKeyValue(key) {
        if (!this.currentNamespace || !this.currentApikey) {
            return;
        }

        // Show loading state in detail viewer
        this.showKeyDetailLoading(key);

        try {
            const params = new URLSearchParams({
                namespace: this.currentNamespace,
                apikey: this.currentApikey,
                key: key
            });

            const data = await apiRequest(`/cache/value?${params.toString()}`);

            if (!data) {
                showError('Failed to load key value');
                return;
            }

            this.renderKeyDetail(data);

        } catch (error) {
            console.error('[CacheExplorer] Error querying key value:', error);
            showError('Failed to load key value: ' + error.message);
            this.clearKeyDetail();
        }
    }

    /**
     * renderKeyDetail displays the value and metadata for a key.
     * 
     * @param {object} data - Key value data
     */
    renderKeyDetail(data) {
        const container = document.getElementById('key-detail-viewer');
        if (!container) return;

        // Format value based on type
        let formattedValue = data.value;
        if (data.value_type === 'json') {
            try {
                const jsonObj = JSON.parse(data.value);
                formattedValue = JSON.stringify(jsonObj, null, 2);
            } catch (e) {
                formattedValue = data.value;
            }
        }

        const ttlDisplay = data.ttl === -1 ? '∞' : `${data.ttl} seconds`;
        const createdAt = data.created_at ? new Date(data.created_at).toLocaleString() : 'N/A';
        const lastAccessed = data.last_accessed ? new Date(data.last_accessed).toLocaleString() : 'N/A';

        container.innerHTML = `
            <div class="key-detail-header">
                <h3>Key: ${escapeHtml(data.key)}</h3>
                <button class="btn-close" onclick="cacheExplorer.clearKeyDetail()">×</button>
            </div>
            <div class="key-detail-metadata">
                <div class="metadata-item">
                    <span class="metadata-label">TTL:</span>
                    <span class="metadata-value">${ttlDisplay}</span>
                </div>
                <div class="metadata-item">
                    <span class="metadata-label">Size:</span>
                    <span class="metadata-value">${formatBytes(data.size_bytes)}</span>
                </div>
                <div class="metadata-item">
                    <span class="metadata-label">Type:</span>
                    <span class="metadata-value">${data.value_type}</span>
                </div>
                <div class="metadata-item">
                    <span class="metadata-label">Created:</span>
                    <span class="metadata-value">${createdAt}</span>
                </div>
                <div class="metadata-item">
                    <span class="metadata-label">Last Accessed:</span>
                    <span class="metadata-value">${lastAccessed}</span>
                </div>
            </div>
            <div class="key-detail-value">
                <h4>Value:</h4>
                <pre><code class="${data.value_type === 'json' ? 'language-json' : ''}">${escapeHtml(formattedValue)}</code></pre>
            </div>
        `;

        container.classList.add('visible');
    }

    /**
     * showKeyDetailLoading shows loading state in key detail viewer.
     * 
     * @param {string} key - Key being loaded
     */
    showKeyDetailLoading(key) {
        const container = document.getElementById('key-detail-viewer');
        if (!container) return;

        container.innerHTML = `
            <div class="key-detail-header">
                <h3>Key: ${escapeHtml(key)}</h3>
            </div>
            <div class="loading-spinner">
                <div class="spinner"></div>
                <p>Loading value...</p>
            </div>
        `;
        container.classList.add('visible');
    }

    /**
     * clearKeyDetail hides the key detail viewer.
     */
    clearKeyDetail() {
        const container = document.getElementById('key-detail-viewer');
        if (container) {
            container.classList.remove('visible');
            container.innerHTML = '';
        }
    }

    /**
     * showLoadingState shows loading state in key list.
     */
    showLoadingState() {
        const tbody = document.getElementById('key-list-tbody');
        if (tbody) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="4" class="text-center">
                        <div class="loading-spinner">
                            <div class="spinner"></div>
                            <span>Loading keys...</span>
                        </div>
                    </td>
                </tr>
            `;
        }
    }

    /**
     * clearKeyList clears the key list table.
     */
    clearKeyList() {
        const tbody = document.getElementById('key-list-tbody');
        if (tbody) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="4" class="text-center">
                        Select a namespace to view cache keys
                    </td>
                </tr>
            `;
        }

        const pageInfo = document.getElementById('pagination-info');
        if (pageInfo) {
            pageInfo.textContent = 'Page 0 of 0';
        }

        this.clearKeyDetail();
    }
}

/**
 * formatTimeAgo formats a date to a relative time string.
 * 
 * @param {Date} date - Date to format
 * @returns {string} Relative time string (e.g., "2m ago", "1h ago")
 */
function formatTimeAgo(date) {
    const seconds = Math.floor((new Date() - date) / 1000);

    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
}

/**
 * escapeHtml escapes HTML special characters to prevent XSS.
 * 
 * @param {string} text - Text to escape
 * @returns {string} Escaped text
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Create global cache explorer instance
const cacheExplorer = new CacheExplorer();

