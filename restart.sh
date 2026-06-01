#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="/opt/project/expenseOwl"
LOG="/tmp/restart_expenseowl.log"
: > "$LOG"

# Choose logfile for the running app: prefer /var/log/expenseowl.log when writable, else fallback to repo log
if [ -w "/var/log" ] || [ -w "/var/log/expenseowl.log" ]; then
  APP_LOG="/var/log/expenseowl.log"
else
  APP_LOG="$ROOT_DIR/expenseowl.log"
fi

echo "[restart.sh] Starting restart at $(date)" | tee -a "$LOG"

cd "$ROOT_DIR" || { echo "[restart.sh] ERROR: cannot cd $ROOT_DIR" | tee -a "$LOG"; exit 1; }

# Build frontend if present
if [ -d "frontend" ]; then
  echo "[restart.sh] Building frontend..." | tee -a "$LOG"
  cd frontend
  if [ -f package-lock.json ] || [ -f yarn.lock ]; then
    npm ci --silent --no-audit --no-fund 2>&1 | tee -a "$LOG" || true
  else
    npm install --silent --no-audit --no-fund 2>&1 | tee -a "$LOG" || true
  fi
  npm run build --silent 2>&1 | tee -a "$LOG" || { echo "[restart.sh] Frontend build failed" | tee -a "$LOG"; }
  cd "$ROOT_DIR"
else
  echo "[restart.sh] No frontend directory, skipping frontend build" | tee -a "$LOG"
fi

# Build Go backend
echo "[restart.sh] Building Go backend..." | tee -a "$LOG"
CGO_ENABLED=0 go build -o expenseowl ./cmd/expenseowl 2>&1 | tee -a "$LOG" || { echo "[restart.sh] Go build failed" | tee -a "$LOG"; }

# Try systemd
if command -v systemctl >/dev/null 2>&1; then
  if systemctl list-unit-files | grep -q "expenseowl.service"; then
    echo "[restart.sh] Restarting systemd service: expenseowl.service" | tee -a "$LOG"
    sudo systemctl daemon-reload 2>&1 | tee -a "$LOG" || true
    sudo systemctl restart expenseowl.service 2>&1 | tee -a "$LOG" || { echo "[restart.sh] systemctl restart failed" | tee -a "$LOG"; }
    sudo systemctl status expenseowl.service --no-pager 2>&1 | tee -a "$LOG" || true
    echo "[restart.sh] Done (systemd)" | tee -a "$LOG"
    echo "OK: systemd" > /tmp/expenseowl-restart-method
    exit 0
  else
    echo "[restart.sh] systemd available but expenseowl.service not found" | tee -a "$LOG"
  fi
else
  echo "[restart.sh] systemctl not available" | tee -a "$LOG"
fi

# Try docker-compose
if [ -f docker-compose.yml ] && command -v docker-compose >/dev/null 2>&1; then
  echo "[restart.sh] Using docker-compose to rebuild and restart" | tee -a "$LOG"
  docker-compose pull 2>&1 | tee -a "$LOG" || true
  docker-compose up -d --build 2>&1 | tee -a "$LOG" || { echo "[restart.sh] docker-compose up failed" | tee -a "$LOG"; }
  echo "[restart.sh] Done (docker-compose)" | tee -a "$LOG"
  echo "OK: docker-compose" > /tmp/expenseowl-restart-method
  exit 0
else
  echo "[restart.sh] docker-compose not available or docker-compose.yml missing" | tee -a "$LOG"
fi

# Fallback: restart binary process
echo "[restart.sh] Fallback: restarting binary process" | tee -a "$LOG"
# Find PIDs matching exact path first, then generic name
PIDS="$(pgrep -f '/opt/project/expenseOwl/expenseowl' || true)"
if [ -z "$PIDS" ]; then
  PIDS="$(pgrep -f '\bexpenseowl\b' || true)"
fi
if [ -n "$PIDS" ]; then
  echo "[restart.sh] Found existing PIDs: $PIDS" | tee -a "$LOG"
  for pid in $PIDS; do
    if [ -n "$pid" ]; then
      echo "[restart.sh] Killing PID $pid" | tee -a "$LOG"
      kill "$pid" 2>&1 | tee -a "$LOG" || true
    fi
  done
  sleep 1
else
  echo "[restart.sh] No running expenseowl processes found" | tee -a "$LOG"
fi

# Start binary
if [ -f "$ROOT_DIR/expenseowl" ]; then
  echo "[restart.sh] Starting ./expenseowl (nohup). Logs -> $APP_LOG" | tee -a "$LOG"
  nohup "$ROOT_DIR/expenseowl" > "$APP_LOG" 2>&1 &
  sleep 1
  NEWPIDS="$(pgrep -f '/opt/project/expenseOwl/expenseowl' || true)"
  if [ -z "$NEWPIDS" ]; then
    NEWPIDS="$(pgrep -f '\bexpenseowl\b' || true)"
  fi
  echo "[restart.sh] Started PIDs: $NEWPIDS" | tee -a "$LOG"
  echo "OK: nohup" > /tmp/expenseowl-restart-method
else
  echo "[restart.sh] ERROR: binary not found at $ROOT_DIR/expenseowl" | tee -a "$LOG"
fi

echo "[restart.sh] Restart finished at $(date)" | tee -a "$LOG"
exit 0
