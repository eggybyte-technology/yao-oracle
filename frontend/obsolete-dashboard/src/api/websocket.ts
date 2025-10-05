// WebSocket client for real-time updates from Admin service

import type { WebSocketMessage } from '../types/metrics';

const WS_URL =
    import.meta.env.VITE_ADMIN_WS_URL ||
    (window.location.protocol === 'https:' ? 'wss://' : 'ws://') +
    window.location.host +
    '/ws';

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export class MetricsWebSocket {
    private ws: WebSocket | null = null;
    private reconnectTimeout: number | null = null;
    private messageHandler: ((msg: WebSocketMessage) => void) | null = null;
    private statusHandler: ((status: ConnectionStatus) => void) | null = null;
    private reconnectDelay = 5000;
    private maxReconnectDelay = 30000;
    private reconnectAttempts = 0;
    private isManualDisconnect = false;

    connect(
        onMessage: (msg: WebSocketMessage) => void,
        onStatusChange?: (status: ConnectionStatus) => void
    ): void {
        this.messageHandler = onMessage;
        this.statusHandler = onStatusChange || null;
        this.isManualDisconnect = false;
        this.reconnectAttempts = 0;
        this.connectInternal();
    }

    private connectInternal(): void {
        // Don't reconnect if manually disconnected
        if (this.isManualDisconnect) {
            return;
        }

        // Clear any existing connection
        if (this.ws && this.ws.readyState !== WebSocket.CLOSED) {
            this.ws.close();
        }

        this.updateStatus('connecting');

        try {
            console.log('[WebSocket] Connecting to', WS_URL);
            this.ws = new WebSocket(WS_URL);

            this.ws.onopen = () => {
                console.log('[WebSocket] ✅ Connected successfully');
                this.reconnectAttempts = 0;
                this.reconnectDelay = 5000;
                this.updateStatus('connected');

                if (this.reconnectTimeout) {
                    clearTimeout(this.reconnectTimeout);
                    this.reconnectTimeout = null;
                }
            };

            this.ws.onmessage = (event) => {
                try {
                    const message: WebSocketMessage = JSON.parse(event.data);
                    if (this.messageHandler) {
                        this.messageHandler(message);
                    }
                } catch (err) {
                    console.error('[WebSocket] Failed to parse message:', err);
                }
            };

            this.ws.onerror = (event) => {
                // Only log error if we're not in the middle of reconnecting
                if (this.reconnectAttempts === 0) {
                    console.error('[WebSocket] Connection error', event);
                }
                // Don't update status here - let onclose handle it
            };

            this.ws.onclose = (event) => {
                if (this.isManualDisconnect) {
                    console.log('[WebSocket] Connection closed manually');
                    this.updateStatus('disconnected');
                    return;
                }

                // Calculate exponential backoff
                this.reconnectAttempts++;
                const delay = Math.min(
                    this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1),
                    this.maxReconnectDelay
                );

                // Only log if it's not a normal closure
                if (event.code !== 1000 && event.code !== 1001) {
                    console.log(
                        `[WebSocket] Disconnected (code: ${event.code}, attempt ${this.reconnectAttempts}), ` +
                        `reconnecting in ${(delay / 1000).toFixed(1)}s...`
                    );
                }

                this.updateStatus('disconnected');

                this.reconnectTimeout = window.setTimeout(() => {
                    this.connectInternal();
                }, delay);
            };
        } catch (err) {
            console.error('[WebSocket] Connection failed:', err);
            this.updateStatus('error');

            const delay = Math.min(
                this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts),
                this.maxReconnectDelay
            );

            this.reconnectTimeout = window.setTimeout(() => {
                this.connectInternal();
            }, delay);
        }
    }

    private updateStatus(status: ConnectionStatus): void {
        if (this.statusHandler) {
            this.statusHandler(status);
        }
    }

    disconnect(): void {
        this.isManualDisconnect = true;

        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        this.messageHandler = null;
        this.statusHandler = null;
        this.updateStatus('disconnected');
    }

    send(message: Record<string, unknown>): void {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.warn('[WebSocket] Cannot send message: connection not open');
        }
    }

    getStatus(): ConnectionStatus {
        if (!this.ws) return 'disconnected';

        switch (this.ws.readyState) {
            case WebSocket.CONNECTING:
                return 'connecting';
            case WebSocket.OPEN:
                return 'connected';
            case WebSocket.CLOSING:
            case WebSocket.CLOSED:
                return 'disconnected';
            default:
                return 'disconnected';
        }
    }
}

