// QPS trend chart using ECharts

import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import type { TimeSeriesPoint } from '../../types/metrics';

interface QPSChartProps {
    data: TimeSeriesPoint[];
    title?: string;
}

export function QPSChart({ data, title = 'QPS Trend' }: QPSChartProps) {
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
        // Initial resize
        handleResize();

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
            legend: {
                data: ['GET', 'SET', 'DELETE'],
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
            },
            grid: {
                left: '3%',
                right: '4%',
                bottom: '14%',
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
                name: 'QPS',
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
            series: [
                {
                    name: 'GET',
                    type: 'line',
                    smooth: true,
                    data: data.map((d) => [d.timestamp, d.qps?.get || 0]),
                    itemStyle: { color: '#10b981' },
                    areaStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: 'rgba(16, 185, 129, 0.4)' },
                            { offset: 1, color: 'rgba(16, 185, 129, 0.05)' }
                        ])
                    },
                    lineStyle: { width: 3, shadowBlur: 10, shadowColor: 'rgba(16, 185, 129, 0.5)' },
                    symbolSize: 8,
                },
                {
                    name: 'SET',
                    type: 'line',
                    smooth: true,
                    data: data.map((d) => [d.timestamp, d.qps?.set || 0]),
                    itemStyle: { color: '#00f5ff' },
                    areaStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: 'rgba(0, 245, 255, 0.4)' },
                            { offset: 1, color: 'rgba(0, 245, 255, 0.05)' }
                        ])
                    },
                    lineStyle: { width: 3, shadowBlur: 10, shadowColor: 'rgba(0, 245, 255, 0.5)' },
                    symbolSize: 8,
                },
                {
                    name: 'DELETE',
                    type: 'line',
                    smooth: true,
                    data: data.map((d) => [d.timestamp, d.qps?.delete || 0]),
                    itemStyle: { color: '#f43f5e' },
                    areaStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: 'rgba(244, 63, 94, 0.4)' },
                            { offset: 1, color: 'rgba(244, 63, 94, 0.05)' }
                        ])
                    },
                    lineStyle: { width: 3, shadowBlur: 10, shadowColor: 'rgba(244, 63, 94, 0.5)' },
                    symbolSize: 8,
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

