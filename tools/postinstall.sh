#!/bin/sh

if ! command -V systemctl >/dev/null 2>&1; then
    exit 0
fi

name=fleetlock

systemctl daemon-reload || true
systemctl enable "${name}.service" || true
systemctl restart "${name}.service" || true
