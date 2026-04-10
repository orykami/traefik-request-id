<p align="center">
  <img src=".assets/icon.png" alt="traefik-request-id" width="120" />
</p>

<h1 align="center">traefik-request-id</h1>

<p align="center">
  A <a href="https://traefik.io">Traefik</a> middleware plugin that generates a unique <code>X-Request-ID</code> header (UUID v4) for every incoming request.
</p>

<p align="center">
  <a href="https://github.com/orykami/traefik-request-id/actions"><img src="https://github.com/orykami/traefik-request-id/workflows/Main/badge.svg?branch=main" alt="Build Status" /></a>
  <a href="https://github.com/orykami/traefik-request-id/releases"><img src="https://img.shields.io/github/v/release/orykami/traefik-request-id" alt="Release" /></a>
  <a href="https://plugins.traefik.io"><img src="https://img.shields.io/badge/traefik-plugin-blue?logo=traefikproxy" alt="Traefik Plugin" /></a>
  <a href="https://github.com/orykami/traefik-request-id/blob/main/LICENSE"><img src="https://img.shields.io/github/license/orykami/traefik-request-id" alt="License" /></a>
</p>

---

## Features

- Generates a **UUID v4** (RFC 4122) for each request
- Sets the ID on the **request** (forwarded to backends) and optionally on the **response**
- Preserves a valid UUID v4 already present in the incoming request
- Overrides invalid or non-UUID client-provided values
- Configurable header name (defaults to `X-Request-ID`)
- Zero external dependencies — uses only Go stdlib

## Configuration

### Static

Add the plugin to your Traefik static configuration:

```yaml
experimental:
  plugins:
    request-id:
      moduleName: github.com/orykami/traefik-request-id
      version: v1.0.0
```

### Dynamic

Enable the middleware on a router:

```yaml
http:
  routers:
    my-router:
      rule: Host(`example.localhost`)
      service: my-service
      entryPoints:
        - web
      middlewares:
        - request-id

  middlewares:
    request-id:
      plugin:
        request-id:
          headerName: X-Request-ID
          setResponseHeader: true
```

### Options

| Option              | Type   | Default        | Description                                      |
|---------------------|--------|----------------|--------------------------------------------------|
| `headerName`        | string | `X-Request-ID` | Name of the header to set                        |
| `setResponseHeader` | bool   | `true`         | Whether to include the ID in the response header |

## Local Development

For local testing without publishing to the Plugin Catalog, use Traefik's local plugin mode:

```
./plugins-local/
    └── src
        └── github.com
            └── orykami
                └── traefik-request-id
                    ├── main.go
                    ├── main_test.go
                    ├── go.mod
                    └── .traefik.yml
```

```yaml
# Static configuration
experimental:
  localPlugins:
    request-id:
      moduleName: github.com/orykami/traefik-request-id
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```