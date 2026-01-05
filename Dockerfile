# Dockerfile for OpenPlantbook MCP Server with native Streamable HTTP transport
# This exposes MCP directly over HTTP (no MCPO needed)
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS requests
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the pre-compiled Linux binary
COPY bin/openplantbook-mcp-linux-amd64 /usr/local/bin/openplantbook-mcp
RUN chmod +x /usr/local/bin/openplantbook-mcp

# Expose the MCP HTTP port
EXPOSE 8000

# Health check (MCP endpoint is at /mcp)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD test -f /proc/1/cmdline || exit 1

# Environment variables:
# - OPENPLANTBOOK_API_KEY (required)
# - OPENPLANTBOOK_LOG_LEVEL (optional, default: info)
# - OPENPLANTBOOK_CACHE_ENABLED (optional, default: true)
# - OPENPLANTBOOK_CACHE_TTL_HOURS (optional, default: 24)

# Run MCP server in HTTP mode
ENTRYPOINT ["/usr/local/bin/openplantbook-mcp"]
CMD ["--http", "--port", "8000"]
