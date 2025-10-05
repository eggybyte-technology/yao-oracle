/* ======================================================================
   Yao-Oracle Dashboard - Chart.js Management
   ====================================================================== */

/**
 * Chart Manager - Manages Chart.js instances lifecycle
 */
const ChartManager = {
    charts: {},

    /**
     * Create a new chart instance
     * @param {string} canvasId - Canvas element ID
     * @param {object} config - Chart.js configuration
     * @returns {Chart} Chart instance
     */
    create(canvasId, config) {
        // Destroy existing chart if any
        this.destroy(canvasId);

        const ctx = document.getElementById(canvasId);
        if (!ctx) {
            console.error(`[ChartManager] Canvas element not found: ${canvasId}`);
            return null;
        }

        try {
            this.charts[canvasId] = new Chart(ctx, config);
            if (CONFIG.DEBUG) {
                console.log(`[ChartManager] Created chart: ${canvasId}`);
            }
            return this.charts[canvasId];
        } catch (error) {
            console.error(`[ChartManager] Failed to create chart ${canvasId}:`, error);
            return null;
        }
    },

    /**
     * Update existing chart with new data
     * @param {string} canvasId - Canvas element ID
     * @param {object} newData - New chart data
     */
    update(canvasId, newData) {
        const chart = this.charts[canvasId];
        if (!chart) {
            console.warn(`[ChartManager] Chart not found for update: ${canvasId}`);
            return;
        }

        chart.data = newData;
        chart.update('none'); // Update without animation for smoother refresh
    },

    /**
     * Destroy chart instance
     * @param {string} canvasId - Canvas element ID
     */
    destroy(canvasId) {
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
            delete this.charts[canvasId];
            if (CONFIG.DEBUG) {
                console.log(`[ChartManager] Destroyed chart: ${canvasId}`);
            }
        }
    },

    /**
     * Destroy all chart instances
     */
    destroyAll() {
        Object.keys(this.charts).forEach(canvasId => this.destroy(canvasId));
    },

    /**
     * Get chart instance
     * @param {string} canvasId - Canvas element ID
     * @returns {Chart} Chart instance
     */
    get(canvasId) {
        return this.charts[canvasId];
    }
};

/**
 * Chart Presets - Reusable chart configurations
 */
const ChartPresets = {
    /**
     * Create a line chart for time series data
     * @param {Array} labels - X-axis labels
     * @param {Array} data - Y-axis data points
     * @param {string} label - Dataset label
     * @param {string} color - Line color
     * @returns {object} Chart configuration
     */
    lineChart(labels, data, label, color = CONFIG.CHART_COLORS.primary) {
        return {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: label,
                    data: data,
                    borderColor: color,
                    backgroundColor: color + '20',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4,
                    pointRadius: 3,
                    pointHoverRadius: 5
                }]
            },
            options: {
                ...CONFIG.CHART_DEFAULTS,
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.05)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false
                    }
                }
            }
        };
    },

    /**
     * Create a gauge chart (doughnut semi-circle)
     * @param {number} value - Current value (0-100)
     * @param {string} label - Chart label
     * @returns {object} Chart configuration
     */
    gaugeChart(value, label) {
        const percentage = Math.min(100, Math.max(0, value));

        return {
            type: 'doughnut',
            data: {
                labels: ['Value', 'Remaining'],
                datasets: [{
                    data: [percentage, 100 - percentage],
                    backgroundColor: [
                        percentage >= 80 ? CONFIG.CHART_COLORS.success :
                            percentage >= 60 ? CONFIG.CHART_COLORS.primary :
                                percentage >= 40 ? CONFIG.CHART_COLORS.warning :
                                    CONFIG.CHART_COLORS.error,
                        'rgba(200, 200, 200, 0.1)'
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                ...CONFIG.CHART_DEFAULTS,
                circumference: 180,
                rotation: 270,
                cutout: '70%',
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            label: function (context) {
                                return context.label + ': ' + context.parsed + '%';
                            }
                        }
                    }
                }
            },
            plugins: [{
                id: 'centerText',
                beforeDraw: function (chart) {
                    const ctx = chart.ctx;
                    const width = chart.width;
                    const height = chart.height;

                    ctx.restore();
                    ctx.font = 'bold 24px sans-serif';
                    ctx.textBaseline = 'middle';
                    ctx.fillStyle = getComputedStyle(document.documentElement)
                        .getPropertyValue('--color-text');

                    const text = percentage.toFixed(1) + '%';
                    const textX = Math.round((width - ctx.measureText(text).width) / 2);
                    const textY = height / 1.4;

                    ctx.fillText(text, textX, textY);

                    ctx.font = '12px sans-serif';
                    ctx.fillStyle = getComputedStyle(document.documentElement)
                        .getPropertyValue('--color-text-secondary');
                    const labelX = Math.round((width - ctx.measureText(label).width) / 2);
                    const labelY = height / 1.15;
                    ctx.fillText(label, labelX, labelY);

                    ctx.save();
                }
            }]
        };
    },

    /**
     * Create a bar chart
     * @param {Array} labels - X-axis labels
     * @param {Array} data - Y-axis data points
     * @param {string} label - Dataset label
     * @param {string|Array} colors - Bar color(s)
     * @returns {object} Chart configuration
     */
    barChart(labels, data, label, colors = CONFIG.CHART_COLORS.primary) {
        const backgroundColor = Array.isArray(colors) ? colors :
            data.map(value => {
                if (value >= 80) return CONFIG.CHART_COLORS.success;
                if (value >= 60) return CONFIG.CHART_COLORS.primary;
                if (value >= 40) return CONFIG.CHART_COLORS.warning;
                return CONFIG.CHART_COLORS.error;
            });

        return {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: label,
                    data: data,
                    backgroundColor: backgroundColor,
                    borderRadius: 6,
                    borderWidth: 0
                }]
            },
            options: {
                ...CONFIG.CHART_DEFAULTS,
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.05)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            label: function (context) {
                                return context.dataset.label + ': ' + context.parsed.y;
                            }
                        }
                    }
                }
            }
        };
    },

    /**
     * Create a doughnut chart
     * @param {Array} labels - Segment labels
     * @param {Array} data - Segment values
     * @param {Array} colors - Segment colors
     * @returns {object} Chart configuration
     */
    doughnutChart(labels, data, colors) {
        return {
            type: 'doughnut',
            data: {
                labels: labels,
                datasets: [{
                    data: data,
                    backgroundColor: colors || [
                        CONFIG.CHART_COLORS.primary,
                        CONFIG.CHART_COLORS.success,
                        CONFIG.CHART_COLORS.warning,
                        CONFIG.CHART_COLORS.error,
                        CONFIG.CHART_COLORS.info
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                ...CONFIG.CHART_DEFAULTS,
                cutout: '60%',
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            padding: 15,
                            usePointStyle: true
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function (context) {
                                const label = context.label || '';
                                const value = context.parsed;
                                const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                const percentage = ((value / total) * 100).toFixed(1);
                                return `${label}: ${value} (${percentage}%)`;
                            }
                        }
                    }
                }
            }
        };
    },

    /**
     * Create a multi-line chart
     * @param {Array} labels - X-axis labels
     * @param {Array} datasets - Array of dataset objects {label, data, color}
     * @returns {object} Chart configuration
     */
    multiLineChart(labels, datasets) {
        const chartDatasets = datasets.map(ds => ({
            label: ds.label,
            data: ds.data,
            borderColor: ds.color || CONFIG.CHART_COLORS.primary,
            backgroundColor: (ds.color || CONFIG.CHART_COLORS.primary) + '10',
            borderWidth: 2,
            fill: false,
            tension: 0.4,
            pointRadius: 2,
            pointHoverRadius: 4
        }));

        return {
            type: 'line',
            data: {
                labels: labels,
                datasets: chartDatasets
            },
            options: {
                ...CONFIG.CHART_DEFAULTS,
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.05)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: true,
                        position: 'top',
                        labels: {
                            padding: 15,
                            usePointStyle: true
                        }
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false
                    }
                }
            }
        };
    }
};

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.ChartManager = ChartManager;
    window.ChartPresets = ChartPresets;
}
