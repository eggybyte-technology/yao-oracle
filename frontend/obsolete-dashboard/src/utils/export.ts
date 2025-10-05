// Data export utilities for CSV, JSON, and other formats

/**
 * Export data to CSV format
 */
export function exportToCSV<T extends Record<string, unknown>>(
    data: T[],
    filename: string,
    headers?: string[]
): void {
    if (data.length === 0) {
        console.warn('No data to export');
        return;
    }

    // Use provided headers or extract from first object
    const keys = headers || Object.keys(data[0]);

    // Create CSV header row
    const csvHeader = keys.map(escapeCSVValue).join(',');

    // Create CSV data rows
    const csvRows = data.map((row) =>
        keys
            .map((key) => {
                const value = row[key];
                return escapeCSVValue(
                    value !== null && value !== undefined ? String(value) : ''
                );
            })
            .join(',')
    );

    // Combine header and rows
    const csv = [csvHeader, ...csvRows].join('\n');

    // Download file
    downloadFile(csv, filename, 'text/csv;charset=utf-8;');
}

/**
 * Export data to JSON format
 */
export function exportToJSON<T>(
    data: T,
    filename: string,
    pretty = true
): void {
    const json = pretty
        ? JSON.stringify(data, null, 2)
        : JSON.stringify(data);

    downloadFile(json, filename, 'application/json;charset=utf-8;');
}

/**
 * Export chart as image (PNG)
 * Requires ECharts instance
 */
export function exportChartAsImage(
    chartInstance: { getDataURL: () => string },
    filename: string
): void {
    try {
        const dataURL = chartInstance.getDataURL({
            type: 'png',
            pixelRatio: 2,
            backgroundColor: '#0a0e1a',
        });

        downloadFile(dataURL, filename, null, true);
    } catch (error) {
        console.error('Failed to export chart:', error);
        alert('Failed to export chart. Please try again.');
    }
}

/**
 * Escape CSV values (handle quotes, commas, newlines)
 */
function escapeCSVValue(value: string): string {
    // If value contains comma, quote, or newline, wrap in quotes and escape quotes
    if (value.includes(',') || value.includes('"') || value.includes('\n')) {
        return `"${value.replace(/"/g, '""')}"`;
    }
    return value;
}

/**
 * Download file helper
 */
function downloadFile(
    content: string,
    filename: string,
    mimeType: string | null,
    isDataURL = false
): void {
    const blob = isDataURL
        ? dataURLtoBlob(content)
        : new Blob([content], { type: mimeType || 'text/plain' });

    const url = isDataURL ? content : URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);

    // Clean up object URL (except for data URLs)
    if (!isDataURL) {
        setTimeout(() => URL.revokeObjectURL(url), 100);
    }
}

/**
 * Convert data URL to Blob
 */
function dataURLtoBlob(dataURL: string): Blob {
    const arr = dataURL.split(',');
    const mime = arr[0].match(/:(.*?);/)?.[1] || 'image/png';
    const bstr = atob(arr[1]);
    let n = bstr.length;
    const u8arr = new Uint8Array(n);
    while (n--) {
        u8arr[n] = bstr.charCodeAt(n);
    }
    return new Blob([u8arr], { type: mime });
}

/**
 * Format timestamp for filename
 */
export function getTimestampedFilename(baseName: string, extension: string): string {
    const now = new Date();
    const timestamp = now
        .toISOString()
        .replace(/[:.]/g, '-')
        .slice(0, 19);
    return `${baseName}_${timestamp}.${extension}`;
}

/**
 * Copy data to clipboard as JSON
 */
export async function copyToClipboard<T>(data: T): Promise<boolean> {
    try {
        const json = JSON.stringify(data, null, 2);
        await navigator.clipboard.writeText(json);
        return true;
    } catch (error) {
        console.error('Failed to copy to clipboard:', error);
        return false;
    }
}

/**
 * Export metrics summary as markdown
 */
export function exportAsMarkdown(
    title: string,
    sections: Array<{ heading: string; content: string | string[] }>
): void {
    let markdown = `# ${title}\n\n`;
    markdown += `Generated: ${new Date().toLocaleString()}\n\n`;
    markdown += '---\n\n';

    for (const section of sections) {
        markdown += `## ${section.heading}\n\n`;
        if (Array.isArray(section.content)) {
            markdown += section.content.join('\n') + '\n\n';
        } else {
            markdown += section.content + '\n\n';
        }
    }

    const filename = getTimestampedFilename(
        title.toLowerCase().replace(/\s+/g, '_'),
        'md'
    );
    downloadFile(markdown, filename, 'text/markdown;charset=utf-8;');
}

