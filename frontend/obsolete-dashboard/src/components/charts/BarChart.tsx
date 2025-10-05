// Bar chart component using ECharts

import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';

interface BarChartProps {
    data: Array<{ name: string; value: number }>;
    title?: string;
    color?: string;
    horizontal?: boolean;
}

export function BarChart({
    data,
    title,
    color = '#3b82f6',
    horizontal = false,
}: BarChartProps) {
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
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow',
                    shadowStyle: {
                        color: 'rgba(0, 245, 255, 0.1)',
                    },
                },
                backgroundColor: 'rgba(15, 20, 25, 0.95)',
                borderColor: 'rgba(0, 245, 255, 0.3)',
                borderWidth: 1,
                textStyle: {
                    color: '#f0f4f8',
                    fontSize: 14,
                    fontFamily: 'Inter',
                },
            },
            grid: {
                left: '3%',
                right: '4%',
                bottom: '3%',
                top: title ? '18%' : '3%',
                containLabel: true,
            },
            xAxis: horizontal
                ? {
                    type: 'value',
                    axisLabel: {
                        fontSize: 13,
                        color: '#b4c0d3',
                        fontWeight: 600,
                        fontFamily: 'Inter',
                    },
                    axisLine: {
                        lineStyle: { color: 'rgba(255, 255, 255, 0.1)' }
                    },
                    splitLine: {
                        lineStyle: { color: 'rgba(255, 255, 255, 0.05)' }
                    }
                }
                : {
                    type: 'category',
                    data: data.map((d) => d.name),
                    axisLabel: {
                        fontSize: 13,
                        color: '#b4c0d3',
                        fontWeight: 600,
                        fontFamily: 'Inter',
                    },
                    axisLine: {
                        lineStyle: { color: 'rgba(255, 255, 255, 0.1)' }
                    }
                },
            yAxis: horizontal
                ? {
                    type: 'category',
                    data: data.map((d) => d.name),
                    axisLabel: {
                        fontSize: 13,
                        color: '#b4c0d3',
                        fontWeight: 600,
                        fontFamily: 'Inter',
                    },
                    axisLine: {
                        lineStyle: { color: 'rgba(255, 255, 255, 0.1)' }
                    }
                }
                : {
                    type: 'value',
                    axisLabel: {
                        fontSize: 13,
                        color: '#b4c0d3',
                        fontWeight: 600,
                        fontFamily: 'Inter',
                    },
                    axisLine: {
                        lineStyle: { color: 'rgba(255, 255, 255, 0.1)' }
                    },
                    splitLine: {
                        lineStyle: { color: 'rgba(255, 255, 255, 0.05)' }
                    }
                },
            series: [
                {
                    type: 'bar',
                    data: data.map((d) => d.value),
                    itemStyle: {
                        color: new echarts.graphic.LinearGradient(
                            0, 0, 0, 1,
                            [
                                { offset: 0, color: color },
                                { offset: 1, color: color + '80' }
                            ]
                        ),
                        borderRadius: horizontal ? [0, 6, 6, 0] : [6, 6, 0, 0],
                        shadowBlur: 15,
                        shadowColor: color + '40',
                    },
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 25,
                            shadowColor: color + '80',
                        }
                    },
                    label: {
                        show: true,
                        position: horizontal ? 'right' : 'top',
                        fontWeight: 700,
                        fontSize: 13,
                        color: '#f0f4f8',
                        fontFamily: 'JetBrains Mono',
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
    }, [data, title, color, horizontal]);

    return <div ref={chartRef} style={{ width: '100%', height: '400px' }} />;
}

