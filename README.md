# Reproxy

```
  _____  ______ _____  _____   ______   ____     __
 |  __ \|  ____|  __ \|  __ \ / __ \ \ / /\ \   / /
 | |__) | |__  | |__) | |__) | |  | \ V /  \ \_/ / 
 |  _  /|  __| |  ___/|  _  /| |  | |> <    \   /  
 | | \ \| |____| |    | | \ \| |__| / . \    | |   
 |_|  \_\______|_|    |_|  \_\\____/_/ \_\   |_|
```

A highly configurable reverse proxy server with support for static responses, file serving, and load balancing.

## Features

- **Declarative YAML Configuration**: Easy to set up with a single YAML file
- **Multiple Host/Port Binding**: Route traffic based on different hosts and ports
- **Static Response Handlers**: Return fixed responses for specific paths
- **Static File Serving**: Serve files from local filesystem
- **Reverse Proxy with Load Balancing**: Distribute traffic across multiple backends
- **Load Balancing Strategies**:
  - Round Robin (default)
  - Least Connections
  - Random
  - IP Hash
  - URI Hash 
  - Sticky Sessions
- **Health Checking**: Automatic health monitoring of backend servers
- **Header Manipulation**: Add or remove HTTP headers
- **Path-based Routing**: Route requests based on URL paths

## Installation

### Using Go

```bash
go install github.com/letronghoangminh/reproxy/cmd/reproxy@latest
```

### Using Docker

```bash
docker pull letronghoangminh/reproxy:latest
docker run -p 2209:2209 -v /path/to/config.yaml:/app/config/config.yaml letronghoangminh/reproxy
```

## Configuration

Create a `config.yaml` file:

```yaml
listeners:
  - host: ["example.com:80"]
    handlers:
      - path: "/api"
        reverse_proxy:
          upstreams: ["http://backend1:8080", "http://backend2:8080", "http://backend3:8080"]
          load_balancing:
            strategy: round_robin
            retries: 3
            try_interval: 5
          add_headers:
            X-Real-IP: "{remote_ip}"
          remove_headers:
            - "X-Test-Header"

      - path: "/static"
        static_response:
          status: 200
          body: "Hello, World!"

      - path: "/files"
        static_files:
          root: "/path/to/files"
          index_files: ["index.html"]

global:
  port: 2209
  log_level: info
```

## Usage

```bash
reproxy --config /path/to/config.yaml
```

### Command-line Options

- `--config`: Path to the configuration file (default: `config/config.yaml`)

## Configuration Reference

### Global Configuration

| Field | Type | Description |
|-------|------|-------------|
| port | int | Default port for the proxy server |
| log_level | string | Logging level (debug, info, warn, error, fatal) |

### Listener Configuration

| Field | Type | Description |
|-------|------|-------------|
| host | []string | List of host:port combinations to listen on |
| handlers | []HandlerConfig | List of request handlers |

### Handler Configuration

| Field | Type | Description |
|-------|------|-------------|
| path | string | URL path to match |
| static_response | StaticResponseConfig | Static response configuration |
| static_files | StaticFilesConfig | Static file serving configuration |
| reverse_proxy | ReverseProxyConfig | Reverse proxy configuration |

### Static Response Configuration

| Field | Type | Description |
|-------|------|-------------|
| status | int | HTTP status code (default: 200) |
| body | string | Response body |

### Static Files Configuration

| Field | Type | Description |
|-------|------|-------------|
| root | string | Root directory for file serving |
| index_files | []string | List of index files to try |

### Reverse Proxy Configuration

| Field | Type | Description |
|-------|------|-------------|
| upstreams | []string | List of backend URLs |
| load_balancing | LoadBalancingConfig | Load balancing configuration |
| add_headers | map[string]string | Headers to add to the request |
| remove_headers | []string | Headers to remove from the request |

### Load Balancing Configuration

| Field | Type | Description |
|-------|------|-------------|
| strategy | string | Load balancing strategy (round_robin, least_conn, random, ip_hash, uri_hash, sticky) |
| retries | int | Maximum number of retries (default: 3) |
| try_interval | int | Interval between retries in seconds (default: 5) |

## Header Variables

When adding headers, the following variables can be used:

| Variable | Description |
|----------|-------------|
| {remote_ip} | Client's IP address |
| {scheme} | Request scheme (http/https) |
| {host} | Request host |
| {path} | Request path |
| {query} | Request query string |
| {method} | Request method |
| {user_agent} | Client's User-Agent |

## Upcoming Features

- Dynamic upstreams for reverse proxy
- More handler filters (method, header, query, etc.)
- Automatic HTTPS via Let's Encrypt or local CA certificates
- HTTP/1.1 and HTTP/2 support
- Response compression and OLTP

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgements

- The reverse proxy implementation is inspired by [golang-load-balancer](https://github.com/leonardo5621/golang-load-balancer)
- The features are inspired by [Caddy](https://github.com/caddyserver/caddy)
