#!/usr/bin/env bash

set -euo pipefail

IPS=(
    # 192.168.1.10
    192.168.100.101
    192.168.100.102
)

KEY_FILE="${HOME}/.ssh/id_ed25519.pub"

if ((${#IPS[@]} == 0)); then
    echo "Add one or more server IPs to IPS before authorizing."
    exit 1
fi

if [[ ! -f "$KEY_FILE" ]]; then
    echo "SSH public key not found: $KEY_FILE"
    exit 1
fi

for ip in "${IPS[@]}"; do
    ssh-copy-id -i "$KEY_FILE" "root@${ip}"
done
