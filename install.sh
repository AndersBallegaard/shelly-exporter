#!/bin/bash

# Check if the script is running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Check if the script is running on Ubuntu
if ! grep -q "Ubuntu" /etc/os-release; then
    echo "This script is only supported on Ubuntu."
    exit 1
fi

# Install dependencies
apt-get update
apt-get install -y golang-go git

# Clone the repository
cd /tmp
git clone https://github.com/AndersBallegaard/shelly-exporter
cd shelly-exporter


# Build the exporter
go build

# Move the exporter to /usr/local/bin
mv shelly-exporter /usr/local/bin

# Create a systemd service
cp shelly-exporter.service /etc/systemd/system
systemctl daemon-reload
systemctl enable shelly-exporter
systemctl start shelly-exporter

# Cleanup
cd ..
rm -rf shelly-exporter