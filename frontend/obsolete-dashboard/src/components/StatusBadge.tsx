// Status indicator badge component

type Status = 'healthy' | 'degraded' | 'down' | 'full' | 'warning' | 'info';

interface StatusBadgeProps {
    status: Status;
    label?: string;
}

const statusConfig: Record<
    Status,
    { color: string; bgColor: string; borderColor: string; emoji: string; label: string }
> = {
    healthy: {
        color: '#10b981',
        bgColor: 'rgba(16, 185, 129, 0.12)',
        borderColor: 'rgba(16, 185, 129, 0.4)',
        emoji: '✓',
        label: 'Healthy'
    },
    degraded: {
        color: '#fbbf24',
        bgColor: 'rgba(251, 191, 36, 0.12)',
        borderColor: 'rgba(251, 191, 36, 0.4)',
        emoji: '⚠',
        label: 'Degraded'
    },
    down: {
        color: '#f43f5e',
        bgColor: 'rgba(244, 63, 94, 0.12)',
        borderColor: 'rgba(244, 63, 94, 0.4)',
        emoji: '✕',
        label: 'Down'
    },
    full: {
        color: '#fbbf24',
        bgColor: 'rgba(251, 191, 36, 0.12)',
        borderColor: 'rgba(251, 191, 36, 0.4)',
        emoji: '⚠',
        label: 'Full'
    },
    warning: {
        color: '#fbbf24',
        bgColor: 'rgba(251, 191, 36, 0.12)',
        borderColor: 'rgba(251, 191, 36, 0.4)',
        emoji: '⚠',
        label: 'Warning'
    },
    info: {
        color: '#00f5ff',
        bgColor: 'rgba(0, 245, 255, 0.12)',
        borderColor: 'rgba(0, 245, 255, 0.4)',
        emoji: 'ℹ',
        label: 'Info'
    },
};

export function StatusBadge({ status, label }: StatusBadgeProps) {
    const config = statusConfig[status] || statusConfig.info;

    return (
        <span
            className="status-badge"
            style={{
                color: config.color,
                backgroundColor: config.bgColor,
                borderColor: config.borderColor,
                fontWeight: 600,
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.4rem',
                padding: '0.4rem 0.9rem',
                borderRadius: '8px',
                fontSize: '0.8125rem',
                border: '1px solid',
                backdropFilter: 'blur(10px)',
                textTransform: 'capitalize',
                letterSpacing: '0.3px',
            }}
        >
            <span className="status-emoji" style={{ fontSize: '0.875rem', lineHeight: 1 }}>
                {config.emoji}
            </span>
            {label || config.label}
        </span>
    );
}

