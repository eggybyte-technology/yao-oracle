// Reusable metric card component

import type React from 'react';

interface MetricCardProps {
    title: string;
    value: string | number;
    unit?: string;
    subtitle?: string;
    icon?: React.ReactNode;
    color?: string;
}

export function MetricCard({
    title,
    value,
    unit,
    subtitle,
    icon,
    color = '#3b82f6',
}: MetricCardProps) {
    return (
        <div className="metric-card" style={{ borderLeftColor: color }}>
            <div className="metric-card-header">
                {icon && <div className="metric-card-icon">{icon}</div>}
                <h3 className="metric-card-title">{title}</h3>
            </div>
            <div className="metric-card-value">
                <span className="value">{value}</span>
                {unit && <span className="unit">{unit}</span>}
            </div>
            {subtitle && <div className="metric-card-subtitle">{subtitle}</div>}
        </div>
    );
}

