#!/usr/bin/env bash
set -euo pipefail

PIDS=()
CONTAINER=""

cleanup() {
    echo ""
    echo "Cleaning up..."
    for pid in "${PIDS[@]}"; do
        kill "$pid" 2>/dev/null && echo "  killed PID $pid" || true
    done
    if [[ -n "$CONTAINER" ]]; then
        podman stop "$CONTAINER" 2>/dev/null && echo "  stopped container $CONTAINER" || true
    fi
    echo "Done."
}
trap cleanup EXIT

echo "Starting test servers..."

# 1) Python HTTP server on :9001
python3 -m http.server 9001 --bind 127.0.0.1 &>/dev/null &
PIDS+=($!)
echo "  :9001  python3 http.server (PID $!)"

# 2) nc listener on :9002
# ncat keeps listening after a connection closes (-k)
ncat -l 127.0.0.1 9002 -k &>/dev/null &
PIDS+=($!)
echo "  :9002  ncat listener (PID $!)"

# 3) Another python server on :9003
python3 -c "
import http.server, socketserver
with socketserver.TCPServer(('127.0.0.1', 9003), http.server.SimpleHTTPRequestHandler) as s:
    s.serve_forever()
" &>/dev/null &
PIDS+=($!)
echo "  :9003  python3 inline server (PID $!)"

# 4) Podman container on :9004 (if podman is available)
if command -v podman &>/dev/null; then
    CONTAINER=$(podman run -d --rm -p 127.0.0.1:9004:80 docker.io/library/nginx:alpine 2>/dev/null) || true
    if [[ -n "$CONTAINER" ]]; then
        echo "  :9004  podman nginx (container ${CONTAINER:0:12})"
    else
        echo "  :9004  skipped (podman run failed)"
    fi
else
    echo "  :9004  skipped (podman not found)"
fi

echo ""
echo "Test with:"
echo "  go run ./cmd/zap/"
echo "  go run ./cmd/zap/ :9001"
echo "  go run ./cmd/zap/ :9001-9004"
echo "  go run ./cmd/zap/ :9001 --dry-run"
echo ""
echo "Press Ctrl+C to stop all servers."
wait
