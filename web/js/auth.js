/* ======================================================================
   Yao-Oracle Dashboard - Authentication Module
   ====================================================================== */

const Auth = {
    /**
     * Check if user is authenticated
     * @returns {boolean} True if authenticated
     */
    isAuthenticated() {
        const token = this.getToken();
        return token !== null && token !== '';
    },

    /**
     * Get authentication token
     * @returns {string|null} Session token or null
     */
    getToken() {
        return localStorage.getItem(CONFIG.TOKEN_KEY);
    },

    /**
     * Store authentication token
     * @param {string} token - Session token
     */
    setToken(token) {
        localStorage.setItem(CONFIG.TOKEN_KEY, token);
    },

    /**
     * Clear authentication token
     */
    clearToken() {
        localStorage.removeItem(CONFIG.TOKEN_KEY);
    },

    /**
     * Login with password
     * @param {string} password - User password
     * @returns {Promise<Object>} Login response
     */
    async login(password) {
        // Test mode: Accept default password without backend
        if (CONFIG.TEST_MODE) {
            if (password === CONFIG.DEFAULT_PASSWORD) {
                const mockSessionId = 'test-session-' + Date.now();
                this.setToken(mockSessionId);
                console.log('[Auth] Test mode login successful');
                return { success: true };
            } else {
                throw new Error('Invalid password. Use default password: ' + CONFIG.DEFAULT_PASSWORD);
            }
        }

        // Production mode: Call backend API
        try {
            const response = await fetch(`${CONFIG.API_BASE}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ password })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Login failed');
            }

            const data = await response.json();

            if (data.success && data.session_id) {
                this.setToken(data.session_id);
                return { success: true };
            } else {
                throw new Error('Invalid response from server');
            }
        } catch (error) {
            console.error('[Auth] Login error:', error);
            throw error;
        }
    },

    /**
     * Logout current user
     * @returns {Promise<void>}
     */
    async logout() {
        const token = this.getToken();

        if (token) {
            try {
                await fetch(`${CONFIG.API_BASE}/auth/logout`, {
                    method: 'POST',
                    headers: {
                        'X-Session-ID': token
                    }
                });
            } catch (error) {
                console.error('[Auth] Logout error:', error);
            }
        }

        this.clearToken();
        window.location.href = '/login';
    },

    /**
     * Redirect to login if not authenticated
     */
    requireAuth() {
        if (!this.isAuthenticated()) {
            window.location.href = '/login';
        }
    },

    /**
     * Redirect to dashboard if already authenticated
     */
    redirectIfAuthenticated() {
        if (this.isAuthenticated()) {
            window.location.href = '/dashboard';
        }
    }
};

// Login form handler (if on login page)
if (document.getElementById('login-form')) {
    Auth.redirectIfAuthenticated();

    const form = document.getElementById('login-form');
    const passwordInput = document.getElementById('password');
    const loginButton = document.getElementById('login-button');
    const buttonText = document.getElementById('button-text');
    const errorMessage = document.getElementById('error-message');

    // Show test mode banner if in test mode
    if (CONFIG.TEST_MODE) {
        const banner = document.getElementById('test-mode-banner');
        const passwordDisplay = document.getElementById('default-password-display');
        if (banner && passwordDisplay) {
            banner.style.display = 'flex';
            passwordDisplay.textContent = CONFIG.DEFAULT_PASSWORD;
        }
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const password = passwordInput.value;

        // Disable form
        passwordInput.disabled = true;
        loginButton.disabled = true;
        buttonText.textContent = 'Signing in...';
        errorMessage.classList.remove('show');

        try {
            await Auth.login(password);
            window.location.href = '/index.html';
        } catch (error) {
            errorMessage.textContent = error.message;
            errorMessage.classList.add('show');

            // Re-enable form
            passwordInput.disabled = false;
            loginButton.disabled = false;
            buttonText.textContent = 'Sign In';
            passwordInput.focus();
        }
    });
}
