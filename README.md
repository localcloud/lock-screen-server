## Simple server that keeps commands.

### Push example:
- Command in json payload are ignored (used always cmd_type:1 on server side)
```bash
curl -X POST http://localhost:8080/command/push \
    -H 'Content-Type: application/json' \
    -d  '{"auth":{"login":"test","password":"test","device_uuid":"bbb"},"to_device_uuid":"aaa","cmd":{"cmd_type":1}}'
```

### Pull example

```bash
curl -X POST http://localhost:8080/command/pull \
    -H 'Content-Type: application/json' \
    -d  '{"auth":{"login":"test","password":"test","device_uuid":"aaa"}}'
```

### Client list example

```bash
curl -X POST http://localhost:8080/clients/list \
    -H 'Content-Type: application/json' \
    -d  '{"auth":{"login":"test","password":"test","device_uuid":"bbb"}}'
```

### TODO
- Improve Auth system (avoid plain password transmitting)
- Add a tiny TCP or WS server for increase push delivery speed
- Add proper HTTP response codes