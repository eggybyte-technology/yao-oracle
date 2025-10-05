// Error boundary component for graceful error handling

import { Component, ReactNode } from 'react';
import './ErrorBoundary.css';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
}

interface State {
    hasError: boolean;
    error: Error | null;
    errorInfo: string | null;
}

export class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = {
            hasError: false,
            error: null,
            errorInfo: null,
        };
    }

    static getDerivedStateFromError(error: Error): State {
        return {
            hasError: true,
            error,
            errorInfo: null,
        };
    }

    componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
        console.error('Error caught by ErrorBoundary:', error, errorInfo);
        this.setState({
            errorInfo: errorInfo.componentStack || null,
        });
    }

    handleReset = () => {
        this.setState({
            hasError: false,
            error: null,
            errorInfo: null,
        });
    };

    render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }

            return (
                <div className="error-boundary">
                    <div className="error-boundary-content">
                        <div className="error-icon">⚠️</div>
                        <h2 className="error-title">Something went wrong</h2>
                        <p className="error-message">
                            {this.state.error?.message || 'An unexpected error occurred'}
                        </p>
                        {this.state.errorInfo && (
                            <details className="error-details">
                                <summary>Error details</summary>
                                <pre className="error-stack">
                                    {this.state.error?.stack}
                                    {this.state.errorInfo}
                                </pre>
                            </details>
                        )}
                        <button
                            className="error-reset-button"
                            onClick={this.handleReset}
                        >
                            Try Again
                        </button>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}

// Functional error display component for async errors
interface ErrorDisplayProps {
    error: string | Error;
    onRetry?: () => void;
}

export function ErrorDisplay({ error, onRetry }: ErrorDisplayProps) {
    const errorMessage = typeof error === 'string' ? error : error.message;

    return (
        <div className="error-display">
            <div className="error-display-content">
                <div className="error-icon">❌</div>
                <h3 className="error-display-title">Error Loading Data</h3>
                <p className="error-display-message">{errorMessage}</p>
                {onRetry && (
                    <button
                        className="error-retry-button"
                        onClick={onRetry}
                    >
                        Retry
                    </button>
                )}
            </div>
        </div>
    );
}

