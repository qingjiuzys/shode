#!/bin/sh
# HTTP client implementation

# HttpGet performs an HTTP GET request
HttpGet() {
    local url="$1"
    local headers="${2:-}"
    
    _http_request "GET" "$url" "" "$headers"
}

# HttpPost performs an HTTP POST request
HttpPost() {
    local url="$1"
    local data="$2"
    local headers="${3:-}"
    
    _http_request "POST" "$url" "$data" "$headers"
}

# HttpPut performs an HTTP PUT request
HttpPut() {
    local url="$1"
    local data="$2"
    local headers="${3:-}"
    
    _http_request "PUT" "$url" "$data" "$headers"
}

# HttpDelete performs an HTTP DELETE request
HttpDelete() {
    local url="$1"
    local headers="${2:-}"
    
    _http_request "DELETE" "$url" "" "$headers"
}

# HttpRequest performs a custom HTTP request
HttpRequest() {
    local method="$1"
    local url="$2"
    local data="${3:-}"
    local headers="${4:-}"
    
    _http_request "$method" "$url" "$data" "$headers"
}

# _http_request performs the actual HTTP request
_http_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local headers="$4"
    
    # Use curl if available
    if command -v curl &> /dev/null; then
        local curl_cmd="curl -s -X $method"
        
        # Add headers
        if [ -n "$headers" ]; then
            curl_cmd="$curl_cmd -H '$headers'"
        fi
        
        # Add data
        if [ -n "$data" ]; then
            curl_cmd="$curl_cmd -d '$data'"
        fi
        
        curl_cmd="$curl_cmd '$url'"
        eval "$curl_cmd"
        return $?
    fi
    
    # Fallback to wget
    if command -v wget &> /dev/null; then
        local wget_cmd="wget -q -O - --method=$method"
        
        if [ -n "$headers" ]; then
            wget_cmd="$wget_cmd --header='$headers'"
        fi
        
        if [ -n "$data" ]; then
            wget_cmd="$wget_cmd --body-data='$data'"
        fi
        
        wget_cmd="$wget_cmd '$url'"
        eval "$wget_cmd"
        return $?
    fi
    
    echo "Error: Neither curl nor wget is available" >&2
    return 1
}
