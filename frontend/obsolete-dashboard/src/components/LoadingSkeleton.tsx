// Loading skeleton component for better UX during data loading

import './LoadingSkeleton.css';

interface SkeletonProps {
    type?: 'card' | 'chart' | 'table' | 'text';
    count?: number;
    height?: string;
    width?: string;
}

export function LoadingSkeleton({
    type = 'text',
    count = 1,
    height,
    width = '100%'
}: SkeletonProps) {
    const getSkeletonElement = () => {
        switch (type) {
            case 'card':
                return (
                    <div className="skeleton-card" style={{ width }}>
                        <div className="skeleton-header" />
                        <div className="skeleton-value" />
                        <div className="skeleton-text" style={{ width: '60%' }} />
                    </div>
                );
            case 'chart':
                return (
                    <div className="skeleton-chart" style={{ height: height || '400px', width }}>
                        <div className="skeleton-chart-title" />
                        <div className="skeleton-chart-body" />
                    </div>
                );
            case 'table':
                return (
                    <div className="skeleton-table" style={{ width }}>
                        <div className="skeleton-table-header" />
                        {[...Array(5)].map((_, i) => (
                            <div key={i} className="skeleton-table-row" />
                        ))}
                    </div>
                );
            case 'text':
            default:
                return (
                    <div
                        className="skeleton-text"
                        style={{ height: height || '1rem', width }}
                    />
                );
        }
    };

    return (
        <div className="skeleton-container">
            {[...Array(count)].map((_, i) => (
                <div key={i} className="skeleton-item">
                    {getSkeletonElement()}
                </div>
            ))}
        </div>
    );
}

// Grid layout for metric cards
export function MetricCardsSkeleton({ count = 4 }: { count?: number }) {
    return (
        <div className="metrics-grid">
            {[...Array(count)].map((_, i) => (
                <LoadingSkeleton key={i} type="card" />
            ))}
        </div>
    );
}

// Chart grid layout
export function ChartsGridSkeleton({ count = 3 }: { count?: number }) {
    return (
        <div className="charts-grid">
            {[...Array(count)].map((_, i) => (
                <LoadingSkeleton key={i} type="chart" />
            ))}
        </div>
    );
}

