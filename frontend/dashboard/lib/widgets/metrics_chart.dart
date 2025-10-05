/// Real-time metrics chart widget using fl_chart
library;

import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/intl.dart';

/// Real-time line chart for displaying metrics over time
class MetricsChart extends StatelessWidget {
  final String title;
  final List<MetricsDataPoint> dataPoints;
  final String yAxisLabel;
  final Color lineColor;
  final double? maxY;
  final bool showPercentage;

  const MetricsChart({
    super.key,
    required this.title,
    required this.dataPoints,
    required this.yAxisLabel,
    this.lineColor = Colors.blue,
    this.maxY,
    this.showPercentage = false,
  });

  @override
  Widget build(BuildContext context) {
    if (dataPoints.isEmpty) {
      return _EmptyChart(title: title);
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                Text(
                  yAxisLabel,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: lineColor,
                        fontWeight: FontWeight.w500,
                      ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            SizedBox(
              height: 200,
              child: LineChart(
                _buildChartData(),
                duration: const Duration(milliseconds: 250),
              ),
            ),
          ],
        ),
      ),
    );
  }

  LineChartData _buildChartData() {
    final spots = dataPoints
        .asMap()
        .entries
        .map((entry) => FlSpot(entry.key.toDouble(), entry.value.value))
        .toList();

    final maxYValue = maxY ??
        (dataPoints.isEmpty
            ? 100
            : dataPoints
                    .map((e) => e.value)
                    .reduce((a, b) => a > b ? a : b) *
                1.2);

    return LineChartData(
      gridData: FlGridData(
        show: true,
        drawVerticalLine: false,
        horizontalInterval: maxYValue / 5,
        getDrawingHorizontalLine: (value) {
          return FlLine(
            color: Colors.grey.withOpacity(0.2),
            strokeWidth: 1,
          );
        },
      ),
      titlesData: FlTitlesData(
        leftTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            reservedSize: 45,
            getTitlesWidget: (value, meta) {
              if (value == meta.max || value == meta.min) {
                return const SizedBox.shrink();
              }
              final formattedValue = showPercentage
                  ? '${value.toStringAsFixed(0)}%'
                  : NumberFormat.compact().format(value);
              return Text(
                formattedValue,
                style: const TextStyle(fontSize: 10),
              );
            },
          ),
        ),
        rightTitles: const AxisTitles(
          sideTitles: SideTitles(showTitles: false),
        ),
        topTitles: const AxisTitles(
          sideTitles: SideTitles(showTitles: false),
        ),
        bottomTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            reservedSize: 30,
            interval: dataPoints.length > 6 ? 2 : 1,
            getTitlesWidget: (value, meta) {
              final index = value.toInt();
              if (index < 0 || index >= dataPoints.length) {
                return const SizedBox.shrink();
              }
              final timestamp = dataPoints[index].timestamp;
              final time = DateFormat('HH:mm:ss').format(timestamp);
              return Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  time,
                  style: const TextStyle(fontSize: 9),
                ),
              );
            },
          ),
        ),
      ),
      borderData: FlBorderData(show: false),
      minX: 0,
      maxX: (dataPoints.length - 1).toDouble(),
      minY: 0,
      maxY: maxYValue,
      lineBarsData: [
        LineChartBarData(
          spots: spots,
          isCurved: true,
          color: lineColor,
          barWidth: 3,
          isStrokeCapRound: true,
          dotData: FlDotData(
            show: true,
            getDotPainter: (spot, percent, barData, index) {
              return FlDotCirclePainter(
                radius: 3,
                color: lineColor,
                strokeWidth: 1,
                strokeColor: Colors.white,
              );
            },
          ),
          belowBarData: BarAreaData(
            show: true,
            color: lineColor.withOpacity(0.1),
          ),
        ),
      ],
      lineTouchData: LineTouchData(
        touchTooltipData: LineTouchTooltipData(
          getTooltipItems: (touchedSpots) {
            return touchedSpots.map((spot) {
              final index = spot.x.toInt();
              if (index < 0 || index >= dataPoints.length) {
                return null;
              }
              final dataPoint = dataPoints[index];
              final timeStr = DateFormat('HH:mm:ss').format(dataPoint.timestamp);
              final valueStr = showPercentage
                  ? '${spot.y.toStringAsFixed(1)}%'
                  : spot.y.toStringAsFixed(1);
              return LineTooltipItem(
                '$timeStr\n$valueStr',
                const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              );
            }).toList();
          },
        ),
      ),
    );
  }
}

/// Empty state for chart when no data is available
class _EmptyChart extends StatelessWidget {
  final String title;

  const _EmptyChart({required this.title});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              height: 200,
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.show_chart,
                      size: 48,
                      color: Colors.grey.withOpacity(0.5),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'No data available',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Colors.grey,
                          ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Data point for metrics chart
class MetricsDataPoint {
  final DateTime timestamp;
  final double value;

  const MetricsDataPoint({
    required this.timestamp,
    required this.value,
  });
}

