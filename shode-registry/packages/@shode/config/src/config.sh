#!/bin/sh
# Configuration management implementation

# ConfigLoad loads a configuration file
ConfigLoad() {
    local config_file="$1"
    local config_type="${2:-auto}"
    
    if [ ! -f "$config_file" ]; then
        echo "Error: Config file not found: $config_file" >&2
        return 1
    fi
    
    # Auto-detect format
    if [ "$config_type" = "auto" ]; then
        case "$config_file" in
            *.json)
                config_type="json"
                ;;
            *.env)
                config_type="env"
                ;;
            *)
                config_type="sh"
                ;;
        esac
    fi
    
    # Load based on type
    case "$config_type" in
        json)
            if command -v jq &> /dev/null; then
                cat "$config_file"
            else
                echo "Error: jq required for JSON config files" >&2
                return 1
            fi
            ;;
        env)
            set -a
            . "$config_file"
            set +a
            ;;
        sh)
            . "$config_file"
            ;;
    esac
}

# ConfigGet gets a configuration value
ConfigGet() {
    local key="$1"
    local default="${2:-}"
    
    eval "echo \"\${$key:-\$default}\""
}

# ConfigSet sets a configuration value
ConfigSet() {
    local key="$1"
    local value="$2"
    export "$key=$value"
}

# ConfigHas checks if a configuration key exists
ConfigHas() {
    local key="$1"
    [ -n "${!key+x}" ]
}

# ConfigMerge merges multiple configuration sources
ConfigMerge() {
    local priority="$1"
    shift
    
    # Load configs in priority order
    for config in "$@"; do
        if [ -f "$config" ]; then
            ConfigLoad "$config"
        fi
    done
    
    # Priority config overrides
    if [ -f "$priority" ]; then
        ConfigLoad "$priority"
    fi
}
