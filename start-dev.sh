#!/bin/bash

# Detect primary active LAN IP (excluding docker interfaces, loopbacks, etc.)
DETECTED_IP=$(hostname -I | awk '{print $1}')

if [ -z "$DETECTED_IP" ]; then
  echo "[WARNING] No active LAN IP detected. Falling back to localhost."
  export HOST_IP="localhost"
else
  echo "[SUCCESS] Detected active LAN IP: $DETECTED_IP"
  export HOST_IP="$DETECTED_IP"
fi

# Stop any currently running containers to avoid conflicts
echo "Stopping existing containers..."
docker compose down

# Boot up the stack and rebuild
echo "Starting hotel QR ordering system..."
docker compose up -d --build

echo ""
echo "=========================================================="
echo "      Hotel QR Ordering System Started Successfully"
echo "=========================================================="
echo "Admin Panel URL:      http://localhost:3000/admin"
echo "Mobile/Phone URL:     http://${HOST_IP}:3000"
echo "=========================================================="
