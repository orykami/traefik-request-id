# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Traefik middleware plugin that generates a UUID v4 `X-Request-ID` header for every incoming request. The ID is set on both the request (forwarded to backends) and the response. The header name is configurable via `headerName`; defaults to `X-Request-ID`. Client-provided values are always overridden.

This is a [Traefik Plugin Catalog](https://plugins.traefik.io/) plugin — it must have zero external dependencies (only Go stdlib). The plugin manifest is `.traefik.yml`.

## Commands

```bash
# Run all tests
go test -v ./...

# Run a single test
go test -v -run TestServeHTTP_GeneratesID ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Architecture

Single-file plugin (`main.go`) in package `traefik_request_id`. No subdirectories or internal packages.

- **Plugin entry points**: `CreateConfig()` and `New()` — required by Traefik's plugin interface.
- **UUID generation**: `newUUID()` produces RFC 4122 v4 UUIDs using `crypto/rand` with a custom `hexEncode` helper (no `fmt` or `encoding/hex` dependency).
- **Go version**: 1.24 (uses range-over-int syntax in tests).

## Constraints

- **Zero dependencies**: Traefik plugins loaded via Yaegi cannot use modules outside stdlib. Never add `require` entries to `go.mod`.
- **Package name must match**: The package name `traefik_request_id` and import path in `.traefik.yml` must stay in sync.