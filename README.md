# 🚀 Reproxy

<div align="center">

### ⚡ A highly configurable reverse proxy server with support for static responses, file serving, and load balancing ⚡

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)
![Status](https://img.shields.io/badge/status-active-success?style=for-the-badge)

</div>

## ✨ Features

- 📝 **Declarative Configuration**: Simple YAML configuration with multiple host/port binding
- 🌐 **Request Handling**:
    - 📋 Static responses, file serving with security protections
    - 🎯 Advanced matching (path, method, headers, query params, client IP)
    - 🛣️ Path-based routing and URL rewriting
- 🔄 **Reverse Proxy**:
    - ⚖️ Multiple load balancing strategies (Round Robin, Least Connections, Random, IP/URI Hash, Sticky Sessions)
    - 🔌 Static and dynamic (DNS-based) upstreams
    - 💓 Automatic health checking of backend servers
- 🔒 **Response Processing**:
    - 📝 Header manipulation (add/remove)
    - 🔒 Automatic security headers
    - 📦 Response compression (Gzip)
    - 🔍 Request tracing

## 🔧 Installation

### 🧙‍♂️ Using Go

```bash
go install github.com/letronghoangminh/reproxy/cmd/reproxy@latest
```

### 🐳 Using Docker

```bash
docker pull psycholog1st/reproxy:latest
docker run -p 2209:2209 -v /path/to/config.yaml:/app/config/config.yaml letronghoangminh/reproxy
```

## ⚙️ Configuration

Create a `config.yaml` file:

```yaml
listeners:
  - host: ["example.com:80"]
    handlers:
      - matchers:
          path: "/api"
          method: ["GET", "POST", "*"]
          headers:
            X-API-Key: "valid-key"
          client_cidrs: ["192.168.1.0/24", "0.0.0.0/0"]
          query:
            version: "v2"
        reverse_proxy:
          rewrite: "/rewrite/{path}"
          upstreams: 
            static: ["http://localhost:8081", "http://localhost:8082", "http://localhost:8083"]
            dynamic:
              - type: A
                value: "http://example.com:8080"
          load_balancing:
            strategy: round_robin
            retries: 3
            try_interval: 5
          add_headers:
            X-Real-IP: "{remote_ip}"
          remove_headers:
            - "X-Test-Header"

      - matchers:
          path: "/static"
        static_response:
          status: 200
          body: "Hello, World!"

      - matchers:
          path: "/files"
        static_files:
          root: "/path/to/files"

global:
  port: 2209
  log_level: info
```

## 🚀 Usage

```bash
reproxy --config /path/to/config.yaml
```

### 🔣 Command-line Options

- `--config`: Path to the configuration file (default: `config/config.yaml`)
- `--log-format`: Log format, either json or console (default: `json`)
- `--version`: Print version information and exit

## 📑 Configuration Reference

### 🌍 Global Configuration

| Field | Type | Description |
|-------|------|-------------|
| port | int | Default port for the proxy server |
| log_level | string | Logging level (debug, info, warn, error, fatal) |

### 🔌 Listener Configuration

| Field | Type | Description |
|-------|------|-------------|
| host | []string | List of host:port combinations to listen on |
| handlers | []HandlerConfig | List of request handlers |

### 🎮 Handler Configuration

| Field | Type | Description |
|-------|------|-------------|
| matchers | MatchersConfig | Request matching configuration |
| static_response | StaticResponseConfig | Static response configuration |
| static_files | StaticFilesConfig | Static file serving configuration |
| reverse_proxy | ReverseProxyConfig | Reverse proxy configuration |

### 🎯 Matchers Configuration

| Field | Type | Description |
|-------|------|-------------|
| path | string | URL path to match |
| method | []string | HTTP methods to match (GET, POST, etc. or * for any) |
| headers | map[string]string | Headers to match |
| query | map[string]string | Query parameters to match |
| client_cidrs | []string | Client IP CIDR ranges to match |

### 📋 Static Response Configuration

| Field | Type | Description |
|-------|------|-------------|
| status | int | HTTP status code (default: 200) |
| body | string | Response body |

### 📂 Static Files Configuration

| Field | Type | Description |
|-------|------|-------------|
| root | string | Root directory for file serving |

### 🔄 Reverse Proxy Configuration

| Field | Type | Description |
|-------|------|-------------|
| upstreams | UpstreamConfig | Upstream configuration |
| rewrite | string | URL rewriting pattern (e.g., "/rewrite/{path}") |
| load_balancing | LoadBalancingConfig | Load balancing configuration |
| add_headers | map[string]string | Headers to add to the request |
| remove_headers | []string | Headers to remove from the request |

### ⚖️ Load Balancing Configuration

| Field | Type | Description |
|-------|------|-------------|
| strategy | string | Load balancing strategy (round_robin, least_conn, random, ip_hash, uri_hash, sticky) |
| retries | int | Maximum number of retries (default: 3) |
| try_interval | int | Interval between retries in seconds (default: 5) |

### 🔌 Upstream Configuration

| Field | Type | Description |
|-------|------|-------------|
| static | []string | List of static upstream server URLs |
| dynamic | []DynamicUpstreamConfig | List of dynamic upstream configurations |

### 🌐 Dynamic Upstream Configuration

| Field | Type | Description |
|-------|------|-------------|
| type | string | DNS record type (A, AAAA, CNAME) |
| value | string | Domain/hostname to resolve |

## 🔄 Header Variables

When adding headers, the following variables can be used:

| Variable | Description |
|----------|-------------|
| {remote_ip} | 🌐 Client's IP address |
| {scheme} | 🔐 Request scheme (http/https) |
| {host} | 🏠 Request host |
| {path} | 🛣️ Request path |
| {query} | ❓ Request query string |
| {method} | 📤 Request method |
| {user_agent} | 🔍 Client's User-Agent |

## 🔮 Upcoming Features
- 🔒 Automatic HTTPS via Let's Encrypt or local CA certificates (with HTTP/2 support)
- 📦 OLTP

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 🙏 Acknowledgements

- The reverse proxy implementation is inspired by [golang-load-balancer](https://github.com/leonardo5621/golang-load-balancer).
- The features are inspired by [Caddy](https://github.com/caddyserver/caddy).
- README (and code) generated by [Claude](https://claude.ai/).
