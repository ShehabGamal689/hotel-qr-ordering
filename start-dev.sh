#!/bin/bash

# Use the stable mDNS (.local) hostname so QR codes never depend on a floating LAN IP.
# Guests on the same Wi-Fi resolve this name via mDNS/Avahi regardless of the laptop's
# current IP address -- it survives DHCP reassignment AND switching Wi-Fi networks.
# Override by exporting HOST_IP (e.g. a public domain) before running this script.
if [ -z "$HOST_IP" ]; then
  export HOST_IP="$(hostname).local"
  echo "[SUCCESS] Using stable mDNS hostname: $HOST_IP"
  echo "[INFO] IP-independent: survives DHCP changes and switching Wi-Fi networks."
else
  echo "[INFO] Using predefined HOST_IP: $HOST_IP"
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
