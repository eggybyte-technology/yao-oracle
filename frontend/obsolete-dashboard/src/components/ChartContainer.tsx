// Chart container with responsive behavior and export capabilities

import { ReactNode, useEffect, useRef, useState } from 'react';
import './ChartContainer.css';

interface ChartContainerProps {
    title: string;
    children: ReactNode;
    onExport?: () => void;
    onFullscreen?: () => void;
    showExportButton?: boolean;
    showFullscreenButton?: boolean;
    height?: string;
    className?: string;
}

export function ChartContainer({
    title,
    children,
    onExport,
    onFullscreen,
    showExportButton = false,
    showFullscreenButton = false,
    height = '400px',
    className = '',
}: ChartContainerProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const [isFullscreen, setIsFullscreen] = useState(false);

    useEffect(() => {
        const handleFullscreenChange = () => {
            setIsFullscreen(!!document.fullscreenElement);
        };

        document.addEventListener('fullscreenchange', handleFullscreenChange);
        return () => {
            document.removeEventListener('fullscreenchange', handleFullscreenChange);
        };
    }, []);

    const handleFullscreen = async () => {
        if (!containerRef.current) return;

        try {
            if (!isFullscreen) {
                await containerRef.current.requestFullscreen();
                if (onFullscreen) onFullscreen();
            } else {
                await document.exitFullscreen();
            }
        } catch (error) {
            console.error('Failed to toggle fullscreen:', error);
        }
    };

    return (
        <div
            ref={containerRef}
            className={`chart-container ${className} ${isFullscreen ? 'fullscreen' : ''}`}
            style={{ height }}
        >
            <div className="chart-header">
                <h3 className="chart-title">{title}</h3>
                <div className="chart-actions">
                    {showExportButton && onExport && (
                        <button
                            className="chart-action-button"
                            onClick={onExport}
                            title="Export chart"
                        >
                            📥
                        </button>
                    )}
                    {showFullscreenButton && (
                        <button
                            className="chart-action-button"
                            onClick={handleFullscreen}
                            title={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
                        >
                            {isFullscreen ? '🗗' : '⛶'}
                        </button>
                    )}
                </div>
            </div>
            <div className="chart-body">
                {children}
            </div>
        </div>
    );
}

// Hook for responsive chart resizing
export function useChartResize(
    chartInstance: { resize: () => void } | null,
    dependencies: unknown[] = []
): void {
    useEffect(() => {
        if (!chartInstance) return;

        const handleResize = () => {
            if (chartInstance && chartInstance.resize) {
                chartInstance.resize();
            }
        };

        // Resize on mount
        handleResize();

        // Resize on window resize with debounce
        let timeoutId: number;
        const debouncedResize = () => {
            clearTimeout(timeoutId);
            timeoutId = window.setTimeout(handleResize, 100);
        };

        window.addEventListener('resize', debouncedResize);
        return () => {
            window.removeEventListener('resize', debouncedResize);
            clearTimeout(timeoutId);
        };
    }, [chartInstance, ...dependencies]);
}

