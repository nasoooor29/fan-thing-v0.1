#!/usr/bin/env bash

set -euo pipefail

IPS=(
    # 192.168.1.10
    192.168.100.101
)

BIN_NAME="main-server"
REMOTE_PATH="/root/${BIN_NAME}"

if ((${#IPS[@]} == 0)); then
    echo "Add one or more server IPs to IPS before deploying."
    exit 1
fi

go build -o "$BIN_NAME" ./cmd/app

for ip in "${IPS[@]}"; do
    host="root@${ip}"
    echo "Deploying to ${host}"

    ssh "$host" "if [ -x '$REMOTE_PATH' ]; then '$REMOTE_PATH' uninstall; '$REMOTE_PATH' stop; fi"
    scp "$BIN_NAME" "${host}:${REMOTE_PATH}"
    ssh "$host" "chmod +x $REMOTE_PATH && $REMOTE_PATH install && $REMOTE_PATH start"
done
