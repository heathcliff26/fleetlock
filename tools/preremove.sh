#!/bin/sh

if [ "$1" != "0" ] && [ "$1" != "remove" ] && [ "$1" != "purge" ]; then
    exit 0
fi

echo "Clean up fleetlock service"
systemctl unmask fleetlock.service
systemctl stop fleetlock.service
systemctl disable fleetlock.service
