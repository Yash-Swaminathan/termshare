#!/usr/bin/env bash
# Smoke-test -print-json: one stdout JSON line, then exit.
set -euo pipefail
cd "$(dirname "$0")/.."

BIN=./dist/termshare-linux-amd64
if [[ ! -x "$BIN" ]]; then
  GOOS=linux GOARCH=amd64 go build -o "$BIN" .
  chmod +x "$BIN"
fi

ADDR=":18765"
OUT=$(mktemp)
ERR=$(mktemp)
cleanup() {
  if [[ -n "${PID:-}" ]] && kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
  fi
  rm -f "$OUT" "$ERR"
}
trap cleanup EXIT

"$BIN" -addr "$ADDR" -host-key smoke -print-json >"$OUT" 2>"$ERR" &
PID=$!

for _ in $(seq 1 50); do
  if grep -q '^{' "$OUT" 2>/dev/null; then
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "process died early:" >&2
    cat "$ERR" >&2
    exit 1
  fi
  sleep 0.1
done

LINE=$(head -n 1 "$OUT")
echo "stdout_line=$LINE"

python3 - "$LINE" <<'PY'
import json, sys
line = sys.argv[1]
obj = json.loads(line)
assert set(obj) == {"viewer", "host", "id"}, obj
assert "key=" not in obj["viewer"], obj
assert obj["host"].endswith("?key=smoke"), obj
assert obj["viewer"].endswith("/s/" + obj["id"]), obj
assert ":18765" in obj["viewer"], obj
print("json_ok", obj)
PY

echo smoke_ok
