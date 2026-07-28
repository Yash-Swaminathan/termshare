#!/usr/bin/env bash
set -euo pipefail
WIN_IP="${1:?win ip}"
PORT="${2:?port}"
cd /mnt/c/termshare
pkill -f /tmp/termshare-lan-test 2>/dev/null || true
rm -f /tmp/ts-lan.json /tmp/ts-lan.err /tmp/ts-lan.pid
go build -o /tmp/termshare-lan-test .
nohup /tmp/termshare-lan-test -print-json -host-key lan-test -lan-ip "$WIN_IP" -addr ":$PORT" \
  >/tmp/ts-lan.json 2>/tmp/ts-lan.err &
echo $! >/tmp/ts-lan.pid
sleep 1
if ! kill -0 "$(cat /tmp/ts-lan.pid)" 2>/dev/null; then
  echo "process died" >&2
  cat /tmp/ts-lan.err >&2 || true
  exit 1
fi
if grep -q 'address already in use' /tmp/ts-lan.err; then
  echo "port in use" >&2
  cat /tmp/ts-lan.err >&2
  exit 1
fi
test -s /tmp/ts-lan.json
cat /tmp/ts-lan.json
