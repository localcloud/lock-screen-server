#!/usr/bin/env bash
curl -X POST http://localhost:8080/command/push \
    -H 'Content-Type: application/json' \
    -d  '{"auth":{"login":"test","password":"test","device_uuid":"bbb"},"to_device_uuid":"aaa","cmd":{"cmd_type":1}}'