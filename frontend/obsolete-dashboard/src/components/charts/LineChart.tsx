// Line chart component for time-series data using ECharts

import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';

interface DataPoint {
    timestamp: string;
    value: number;
}

interface Dataset {
    label: string;
    data: DataPoint[];
    color: string;
    fill?: boolean;
}

interface Props {
    datasets: Dataset[];
    title: string;
    yAxisLabel?: string;
    height?: number;
}

export function LineChart({ datasets, title, yAxisLabel, height = 300 }: Props) {
    const chartRef = useRef<HTMLDivElement>(null);
    const chartInstance = useRef<echarts.ECharts | null>(null);

    useEffect(() => {
        if (!chartRef.current) return;

        if (!chartInstance.current) {
            chartInstance.current = echarts.init(chartRef.current);
        }

        const chart = chartInstance.current;

        // Resize handler with debounce
        let resizeTimeout: number;
        const handleResize = () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = window.setTimeout(() => {
                chart.resize();
            }, 100);
        };

        window.addEventListener('resize', handleResize);
        handleResize();

        // Prepare series data
        const series = datasets.map((ds) => ({
            name: ds.label,
            type: 'line' as const,
            smooth: true,
            data: ds.data.map((d) => [d.timestamp, d.value]),
            itemStyle: { color: ds.color },
            areaStyle: ds.fill ? {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                    { offset: 0, color: ds.color + '66' },
                    { offset: 1, color: ds.color + '0d' }
                ])
            } : undefined,
            lineStyle: {
                width: 2.5,
                shadowBlur: 8,
                shadowColor: ds.color + '80'
            },
            symbolSize: 6,
            emphasis: {
                focus: 'series',
                itemStyle: {
                    shadowBlur: 12,
                    shadowColor: ds.color,
                }
            }
        }));

        const option: echarts.EChartsOption = {
            animation: false,
            backgroundColor: 'transparent',
            title: {
                text: title,
                left: 'center',
                top: 10,
                textStyle: {
                    fontSize: 18,
                    fontWeight: 700,
                    color: '#f0f4f8',
                    fontFamily: 'Inter',
                }
            },
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross',
                    crossStyle: {
                        color: 'rgba(0, 245, 255, 0.5)',
                    },
                    lineStyle: {
                        color: 'rgba(0, 245, 255, 0.3)',
                    }
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
            legend: datasets.length > 1 ? {
                data: datasets.map(ds => ds.label),
                bottom: 10,
                textStyle: {
                    fontSize: 13,
                    color: '#b4c0d3',
                    fontWeight: 600,
                    fontFamily: 'Inter',
                },
                itemWidth: 28,
                itemHeight: 14,
                itemGap: 20,
            } : undefined,
            grid: {
                left: '3%',
                right: '4%',
                bottom: datasets.length > 1 ? '14%' : '3%',
                top: '18%',
                containLabel: true,
            },
            xAxis: {
                type: 'time',
                axisLabel: {
                    fontSize: 12,
                    color: '#b4c0d3',
                    fontWeight: 600,
                    fontFamily: 'Inter',
                },
                axisLine: {
                    lineStyle: { color: 'rgba(255, 255, 255, 0.1)' }
                },
                splitLine: {
                    show: false,
                }
            },
            yAxis: {
                type: 'value',
                name: yAxisLabel,
                nameTextStyle: {
                    fontSize: 14,
                    fontWeight: 700,
                    color: '#b4c0d3',
                    fontFamily: 'Inter',
                },
                axisLabel: {
                    fontSize: 12,
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
            series: series,
        };

        chart.setOption(option, { notMerge: true });

        return () => {
            window.removeEventListener('resize', handleResize);
            clearTimeout(resizeTimeout);
            chart.dispose();
            chartInstance.current = null;
        };
    }, [datasets, title, yAxisLabel]);

    return <div ref={chartRef} style={{ width: '100%', height: `${height}px` }} />;
}

