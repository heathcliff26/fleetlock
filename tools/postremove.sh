#!/bin/sh

if [ "$1" != "0" ] && [ "$1" != "remove" ] && [ "$1" != "purge" ]; then
    exit 0
fi

systemctl daemon-reload
systemctl reset-failed
