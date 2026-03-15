#!/bin/sh
set -e

# ---------------------------------------------------------------------------
# Multi-repo MCP server entrypoint
#
# The server scans REPOS_DIR (default: /repos) at startup to discover all
# existing repositories and their worktrees.  No initial clone is performed
# here – use the management API (POST /api/repos) to add repositories.
#
# Environment variables:
#   REPOS_DIR   Root directory for repositories (default: /repos)
#   MCP_ADDR    HTTP listen address              (default: :8080)
# ---------------------------------------------------------------------------

REPOS_DIR="${REPOS_DIR:-/repos}"
MCP_ADDR="${MCP_ADDR:-:8080}"

mkdir -p "$REPOS_DIR"

exec /usr/local/bin/code-mcp \
    --repos-dir "$REPOS_DIR" \
    --addr      "$MCP_ADDR"

