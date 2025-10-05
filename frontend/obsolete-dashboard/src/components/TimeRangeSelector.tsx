// Time range selector component for historical data queries

import { useState } from 'react';

export type TimeRange = '5m' | '15m' | '1h' | '6h' | '24h' | '7d' | 'custom';

interface TimeRangeSelectorProps {
    onRangeChange: (range: TimeRange, from?: Date, to?: Date) => void;
    defaultRange?: TimeRange;
}

export function TimeRangeSelector({ onRangeChange, defaultRange = '1h' }: TimeRangeSelectorProps) {
    const [selectedRange, setSelectedRange] = useState<TimeRange>(defaultRange);
    const [showCustom, setShowCustom] = useState(false);
    const [customFrom, setCustomFrom] = useState('');
    const [customTo, setCustomTo] = useState('');

    const timeRanges: { value: TimeRange; label: string }[] = [
        { value: '5m', label: 'Last 5 Minutes' },
        { value: '15m', label: 'Last 15 Minutes' },
        { value: '1h', label: 'Last 1 Hour' },
        { value: '6h', label: 'Last 6 Hours' },
        { value: '24h', label: 'Last 24 Hours' },
        { value: '7d', label: 'Last 7 Days' },
        { value: 'custom', label: 'Custom Range' },
    ];

    const handleRangeSelect = (range: TimeRange) => {
        setSelectedRange(range);
        if (range === 'custom') {
            setShowCustom(true);
        } else {
            setShowCustom(false);
            onRangeChange(range);
        }
    };

    const handleCustomApply = () => {
        if (customFrom && customTo) {
            const from = new Date(customFrom);
            const to = new Date(customTo);
            if (from < to) {
                onRangeChange('custom', from, to);
                setShowCustom(false);
            } else {
                alert('Start time must be before end time');
            }
        }
    };

    return (
        <div className="time-range-selector">
            <div className="time-range-buttons">
                {timeRanges.map((range) => (
                    <button
                        key={range.value}
                        onClick={() => handleRangeSelect(range.value)}
                        className={`time-range-btn ${selectedRange === range.value ? 'active' : ''}`}
                        style={{
                            padding: '0.375rem 0.75rem',
                            fontSize: '0.875rem',
                            fontWeight: 500,
                            borderRadius: '0.375rem',
                            border: '1px solid rgba(255, 255, 255, 0.1)',
                            backgroundColor:
                                selectedRange === range.value
                                    ? '#3b82f6'
                                    : 'rgba(255, 255, 255, 0.05)',
                            color: selectedRange === range.value ? '#fff' : '#9ca3af',
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                        }}
                    >
                        {range.label}
                    </button>
                ))}
            </div>

            {showCustom && (
                <div
                    className="custom-range-inputs"
                    style={{
                        marginTop: '1rem',
                        padding: '1rem',
                        backgroundColor: 'rgba(0, 0, 0, 0.2)',
                        borderRadius: '0.5rem',
                        display: 'flex',
                        gap: '1rem',
                        alignItems: 'flex-end',
                    }}
                >
                    <div style={{ flex: 1 }}>
                        <label
                            htmlFor="custom-from"
                            style={{
                                display: 'block',
                                marginBottom: '0.5rem',
                                color: '#e5e7eb',
                                fontSize: '0.875rem',
                            }}
                        >
                            From
                        </label>
                        <input
                            id="custom-from"
                            type="datetime-local"
                            value={customFrom}
                            onChange={(e) => setCustomFrom(e.target.value)}
                            style={{
                                width: '100%',
                                padding: '0.5rem',
                                backgroundColor: 'rgba(0, 0, 0, 0.3)',
                                border: '1px solid rgba(255, 255, 255, 0.1)',
                                borderRadius: '0.375rem',
                                color: '#e5e7eb',
                            }}
                        />
                    </div>
                    <div style={{ flex: 1 }}>
                        <label
                            htmlFor="custom-to"
                            style={{
                                display: 'block',
                                marginBottom: '0.5rem',
                                color: '#e5e7eb',
                                fontSize: '0.875rem',
                            }}
                        >
                            To
                        </label>
                        <input
                            id="custom-to"
                            type="datetime-local"
                            value={customTo}
                            onChange={(e) => setCustomTo(e.target.value)}
                            style={{
                                width: '100%',
                                padding: '0.5rem',
                                backgroundColor: 'rgba(0, 0, 0, 0.3)',
                                border: '1px solid rgba(255, 255, 255, 0.1)',
                                borderRadius: '0.375rem',
                                color: '#e5e7eb',
                            }}
                        />
                    </div>
                    <div>
                        <button
                            onClick={handleCustomApply}
                            style={{
                                padding: '0.5rem 1rem',
                                backgroundColor: '#10b981',
                                color: '#fff',
                                border: 'none',
                                borderRadius: '0.375rem',
                                cursor: 'pointer',
                                fontWeight: 600,
                                fontSize: '0.875rem',
                            }}
                        >
                            Apply
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}

