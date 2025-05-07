# Secure Proxy

for Ollama and other LLMs

## Overview

This is a simple proxy server that forwards requests to an Ollama API. It adds a security layer by requiring an API key for access.

## Features

- Secure access to the Ollama API
- Simple and easy to use
- Lightweight and fast

## Requirements

- Go 1.20 or later

## Build

```bash
go mod tidy
```

```bash
go build -o bin/sproxy
```

## Usage

### Set the API Key

create a file named `.env` in the root directory of the project and add the following lines to it:

```text
# use static keys for test only
# SECURE_PROXY_WITH_STATIC=True
# SECURE_PROXY_STATIC_MAP=key1=user1,key2=user2
SECURE_PROXY_WITH_REDIS=True
REDIS_URL=redis://localhost:6379/0
SECURE_PROXY_TARGET=http://localhost
SECURE_PROXY_PORT=8080
```

### Start the Proxy Server

```bash
bin/sproxy
```

### Test the Proxy Server

```bash
curl -H "Authorization: Bearer key1" http://localhost:8080/
```

## References

- [Redis URL format](https://pkg.go.dev/github.com/redis/go-redis/v9#ParseURL)
