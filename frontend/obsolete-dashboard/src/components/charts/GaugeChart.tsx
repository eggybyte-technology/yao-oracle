// Gauge chart for hit ratio using ECharts

import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';

interface GaugeChartProps {
    value: number; // 0-1 range
    title?: string;
}

export function GaugeChart({ value, title = 'Hit Ratio' }: GaugeChartProps) {
    const chartRef = useRef<HTMLDivElement>(null);
    const chartInstance = useRef<echarts.ECharts | null>(null);

    useEffect(() => {
        if (!chartRef.current) return;

        if (!chartInstance.current) {
            chartInstance.current = echarts.init(chartRef.current);
        }

        const chart = chartInstance.current;

        // Resize handler
        let resizeTimeout: number;
        const handleResize = () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = window.setTimeout(() => chart.resize(), 100);
        };

        window.addEventListener('resize', handleResize);
        handleResize();

        const percentage = (value * 100).toFixed(1);

        const option: echarts.EChartsOption = {
            animation: false,
            backgroundColor: 'transparent',
            series: [
                {
                    type: 'gauge',
                    startAngle: 180,
                    endAngle: 0,
                    min: 0,
                    max: 100,
                    splitNumber: 10,
                    radius: '85%',
                    axisLine: {
                        lineStyle: {
                            width: 32,
                            color: [
                                [0.5, '#f43f5e'],
                                [0.75, '#fbbf24'],
                                [0.9, '#10b981'],
                                [1, '#00f5ff'],
                            ],
                            shadowColor: 'rgba(0, 245, 255, 0.5)',
                            shadowBlur: 20,
                        },
                    },
                    pointer: {
                        itemStyle: {
                            color: '#00f5ff',
                            shadowColor: 'rgba(0, 245, 255, 0.8)',
                            shadowBlur: 15,
                        },
                        length: '70%',
                        width: 6,
                    },
                    axisTick: {
                        distance: -35,
                        length: 10,
                        lineStyle: {
                            color: 'rgba(255, 255, 255, 0.6)',
                            width: 2,
                        },
                    },
                    splitLine: {
                        distance: -40,
                        length: 18,
                        lineStyle: {
                            color: 'rgba(255, 255, 255, 0.8)',
                            width: 3,
                            shadowColor: 'rgba(0, 245, 255, 0.5)',
                            shadowBlur: 10,
                        },
                    },
                    axisLabel: {
                        color: '#b4c0d3',
                        distance: 50,
                        fontSize: 15,
                        fontWeight: 600,
                        fontFamily: 'Inter',
                    },
                    detail: {
                        valueAnimation: false,
                        formatter: '{value}%',
                        color: '#f0f4f8',
                        fontSize: 48,
                        fontWeight: 800,
                        fontFamily: 'JetBrains Mono',
                        offsetCenter: [0, '70%'],
                        shadowColor: 'rgba(0, 245, 255, 0.6)',
                        shadowBlur: 20,
                    },
                    title: {
                        offsetCenter: [0, '90%'],
                        fontSize: 16,
                        fontWeight: 700,
                        color: '#6b7a94',
                        fontFamily: 'Inter',
                    },
                    data: [
                        {
                            value: parseFloat(percentage),
                            name: title,
                        },
                    ],
                },
            ],
        };

        chart.setOption(option, { notMerge: true });

        return () => {
            window.removeEventListener('resize', handleResize);
            clearTimeout(resizeTimeout);
            chart.dispose();
            chartInstance.current = null;
        };
    }, [value, title]);

    return <div ref={chartRef} style={{ width: '100%', height: '300px' }} />;
}

