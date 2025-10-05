// State management for metrics data using Zustand

import { create } from 'zustand';
import type {
    ClusterOverview,
    ProxyMetrics,
    NodeMetrics,
    NamespaceStats,
    ClusterEvent,
} from '../types/metrics';
import type { ConnectionStatus } from '../api/websocket';

interface MetricsState {
    overview: ClusterOverview | null;
    proxies: ProxyMetrics[];
    nodes: NodeMetrics[];
    namespaces: NamespaceStats[];
    events: ClusterEvent[];
    loading: boolean;
    error: string | null;
    wsStatus: ConnectionStatus;

    // Actions
    setOverview: (overview: ClusterOverview) => void;
    updateOverview: (update: Partial<ClusterOverview>) => void;
    setProxies: (proxies: ProxyMetrics[]) => void;
    updateProxy: (proxy: ProxyMetrics) => void;
    setNodes: (nodes: NodeMetrics[]) => void;
    updateNode: (node: NodeMetrics) => void;
    setNamespaces: (namespaces: NamespaceStats[]) => void;
    addEvent: (event: ClusterEvent) => void;
    setLoading: (loading: boolean) => void;
    setError: (error: string | null) => void;
    setWsStatus: (status: ConnectionStatus) => void;
}

export const useMetricsStore = create<MetricsState>((set) => ({
    overview: null,
    proxies: [],
    nodes: [],
    namespaces: [],
    events: [],
    loading: false,
    error: null,
    wsStatus: 'disconnected',

    setOverview: (overview) => set({ overview }),

    updateOverview: (update) =>
        set((state) => ({
            overview: state.overview ? { ...state.overview, ...update } : null,
        })),

    setProxies: (proxies) => set({ proxies }),

    updateProxy: (proxy) =>
        set((state) => {
            const index = state.proxies.findIndex((p) => p.id === proxy.id);
            if (index >= 0) {
                const newProxies = [...state.proxies];
                newProxies[index] = proxy;
                return { proxies: newProxies };
            } else {
                return { proxies: [...state.proxies, proxy] };
            }
        }),

    setNodes: (nodes) => set({ nodes }),

    updateNode: (node) =>
        set((state) => {
            const index = state.nodes.findIndex((n) => n.id === node.id);
            if (index >= 0) {
                const newNodes = [...state.nodes];
                newNodes[index] = node;
                return { nodes: newNodes };
            } else {
                return { nodes: [...state.nodes, node] };
            }
        }),

    setNamespaces: (namespaces) => set({ namespaces }),

    addEvent: (event) =>
        set((state) => ({
            events: [event, ...state.events].slice(0, 50), // Keep last 50 events
        })),

    setLoading: (loading) => set({ loading }),

    setError: (error) => set({ error }),

    setWsStatus: (status) => set({ wsStatus: status }),
}));

