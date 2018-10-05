#!/usr/bin/env bash
curl -X POST http://localhost:8080/clients/list \
    -H 'Content-Type: application/json' \
    -d  '{"auth":{"login":"test","password":"test","device_uuid":"bbb"}}'