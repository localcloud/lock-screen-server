#!/usr/bin/env bash
curl -X POST http://localhost:8080/command/pull \
    -H 'Content-Type: application/json' \
    -d  '{"auth":{"login":"test","password":"test","device_uuid":"aaa"}}'