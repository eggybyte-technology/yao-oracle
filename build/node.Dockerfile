# build/node.Dockerfile
# Lightweight runtime-only Dockerfile
# Binary must be built locally before running docker build

# Use buildx automatic platform args
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Runtime stage
FROM alpine:3.22.1

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# Set working directory
WORKDIR /app

# Copy pre-built binary for the target architecture
# The binary should be at bin/${TARGETOS}/${TARGETARCH}/node
ARG TARGETOS
ARG TARGETARCH
COPY bin/${TARGETOS}/${TARGETARCH}/node /app/node

# Change ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/node", "--health-check"] || exit 1

# Run the application
ENTRYPOINT ["/app/node"]

