# @shode/http

HTTP client library for Shode applications.

## Features

- GET, POST, PUT, DELETE requests
- Custom headers support
- JSON data support
- Uses curl or wget (auto-detect)

## Installation

```bash
shode pkg add @shode/http ^1.0.0
```

## Usage

```bash
. sh_modules/@shode/http/index.sh

# GET request
response=$(HttpGet "https://api.example.com/data")

# POST request with JSON
HttpPost "https://api.example.com/users" '{"name":"John"}'

# POST with headers
HttpPost "https://api.example.com/data" '{"key":"value"}' "Content-Type: application/json"

# PUT request
HttpPut "https://api.example.com/posts/123" '{"title":"New Title"}"

# DELETE request
HttpDelete "https://api.example.com/posts/123"
```

## API

### Functions

- `HttpGet(url, headers)` - Perform GET request
- `HttpPost(url, data, headers)` - Perform POST request
- `HttpPut(url, data, headers)` - Perform PUT request
- `HttpDelete(url, headers)` - Perform DELETE request
- `HttpRequest(method, url, data, headers)` - Custom request

## Requirements

- curl or wget

## License

MIT
