#!/bin/sh

systemctl daemon-reload
systemctl enable --now fleetlock.service
