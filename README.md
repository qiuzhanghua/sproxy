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
create a file named `.env` in the root directory of the project and add the following line to it:
```text
SECURE_PROXY_WITH_STATIC=True
SECURE_PROXY_STATIC_MAP=key1=user1,key2=user2
```
### Start the Proxy Server
```bash
bin/sproxy
```
### Test the Proxy Server
```bash
curl -H "Authorization: Bearer key1" http://localhost:8080/
```
