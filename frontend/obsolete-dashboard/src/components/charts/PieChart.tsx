// Pie chart component using ECharts

import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';

interface PieChartProps {
    data: Array<{ name: string; value: number }>;
    title?: string;
}

export function PieChart({ data, title }: PieChartProps) {
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

        const option: echarts.EChartsOption = {
            animation: false,
            backgroundColor: 'transparent',
            title: title ? {
                text: title,
                left: 'center',
                top: 10,
                textStyle: {
                    fontSize: 18,
                    fontWeight: 700,
                    color: '#f0f4f8',
                    fontFamily: 'Inter',
                }
            } : undefined,
            tooltip: {
                trigger: 'item',
                formatter: '{a} <br/>{b}: {c} ({d}%)',
                backgroundColor: 'rgba(15, 20, 25, 0.95)',
                borderColor: 'rgba(0, 245, 255, 0.3)',
                borderWidth: 1,
                textStyle: {
                    color: '#f0f4f8',
                    fontSize: 14,
                    fontFamily: 'Inter',
                },
            },
            legend: {
                orient: 'vertical',
                left: 'left',
                top: 'middle',
                textStyle: {
                    fontSize: 14,
                    color: '#b4c0d3',
                    fontWeight: 600,
                    fontFamily: 'Inter',
                },
                itemWidth: 20,
                itemHeight: 14,
                itemGap: 16,
            },
            color: ['#00f5ff', '#a855f7', '#10b981', '#fbbf24', '#f43f5e', '#06b6d4'],
            series: [
                {
                    name: title || 'Distribution',
                    type: 'pie',
                    radius: ['45%', '75%'],
                    center: ['60%', '50%'],
                    data: data,
                    label: {
                        formatter: '{b}\n{d}%',
                        fontSize: 13,
                        fontWeight: 700,
                        color: '#f0f4f8',
                        fontFamily: 'Inter',
                    },
                    labelLine: {
                        lineStyle: {
                            color: 'rgba(180, 192, 211, 0.5)',
                            width: 2,
                        },
                    },
                    itemStyle: {
                        borderRadius: 8,
                        borderColor: 'rgba(10, 14, 26, 0.8)',
                        borderWidth: 3,
                        shadowBlur: 15,
                        shadowColor: 'rgba(0, 0, 0, 0.5)',
                    },
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 25,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 245, 255, 0.6)',
                        },
                        label: {
                            fontSize: 15,
                            fontWeight: 800,
                        },
                    },
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
    }, [data, title]);

    return <div ref={chartRef} style={{ width: '100%', height: '400px' }} />;
}

