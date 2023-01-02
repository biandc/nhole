# nhole

```bash
███▄▄▄▄      ▄█    █▄     ▄██████▄   ▄█          ▄████████
███▀▀▀██▄   ███    ███   ███    ███ ███         ███    ███
███   ███   ███    ███   ███    ███ ███         ███    █▀  
███   ███  ▄███▄▄▄▄███▄▄ ███    ███ ███        ▄███▄▄▄     
███   ███ ▀▀███▀▀▀▀███▀  ███    ███ ███       ▀▀███▀▀▀     
███   ███   ███    ███   ███    ███ ███         ███    █▄  
███   ███   ███    ███   ███    ███ ███▌    ▄   ███    ███
▀█   █▀    ███    █▀     ▀██████▀  █████▄▄██   ██████████
▀                      
```

nhole is an intranet penetration tool (agent).

## config

### server

./configfiles/nhole-server.yaml

```yaml
server:
  ip: "0.0.0.0"
  control_port: 65531
```

### client

./configfiles/nhole-client.yaml

```yaml
server:
  ip: "127.0.0.1"   // nhole-server ip
  control_port: 65531 // nhole-server control port

service:    // services
  - ip: "127.0.0.1"     // nhole-client local ip
    port: 22            // nhole-client local port
    forward_port: 65532 // nhole-server forward port

  - ip: "127.0.0.1"
    port: 80
    forward_port: 65533
    
    ...
```

## start-up

### server

```bash
go run ./cmd/server/main.go
./nhole-server -c nhole-server.yaml
```

### client

```bash
go run ./cmd/client/main.go
./nhole-client -c nhole-client.yaml
```

## compile

```bash
make clean
make all  # generate files path ./bin
```