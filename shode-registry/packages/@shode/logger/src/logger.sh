#!/bin/sh
# Logger implementation for Shode

# Configuration
LOG_LEVEL="${LOG_LEVEL:-info}"
LOG_FORMAT="${LOG_FORMAT:-text}"
LOG_TRANSPORTS="${LOG_TRANSPORTS:-console}"

# Log levels map
declare -A LOG_LEVELS=(
    ["debug"]=0
    ["info"]=1
    ["warn"]=2
    ["error"]=3
)

# SetLogLevel sets the minimum log level
SetLogLevel() {
    local level="$1"
    if [ -z "${LOG_LEVELS[$level]}" ]; then
        echo "Error: Invalid log level '$level'. Must be one of: debug, info, warn, error" >&2
        return 1
    fi
    LOG_LEVEL="$level"
}

# _log_level_value returns the numeric value of a log level
_log_level_value() {
    echo "${LOG_LEVELS[$1]}"
}

# _should_log checks if a message should be logged based on level
_should_log() {
    local level="$1"
    local level_value="$(_log_level_value "$level")"
    local current_value="$(_log_level_value "$LOG_LEVEL")"
    
    [ "$level_value" -ge "$current_value ]
}

# _log writes a log message
_log() {
    local level="$1"
    shift
    
    if ! _should_log "$level"; then
        return 0
    fi
    
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    local message="$*"
    
    case "$LOG_FORMAT" in
        json)
            echo "{\"level\":\"$level\",\"timestamp\":\"$timestamp\",\"message\":\"$message\"}"
            ;;
        text)
            echo "[$timestamp] [$level] $message"
            ;;
    esac
}

# Public API

# LogDebug logs a debug message
LogDebug() {
    _log "debug" "$@"
}

# LogInfo logs an info message
LogInfo() {
    _log "info" "$@"
}

# LogWarn logs a warning message
LogWarn() {
    _log "warn" "$@"
}

# LogError logs an error message
LogError() {
    _log "error" "$@"
}

# AddLogTransport adds a log transport (reserved for future use)
AddLogTransport() {
    # Future: support for file, syslog, remote transports
    :
}
