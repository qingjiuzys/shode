# @shode/config

Configuration management library for Shode applications.

## Features

- Load configurations from multiple formats (JSON, ENV, Shell)
- Get and set configuration values
- Merge multiple configuration sources
- Environment variable support

## Installation

```bash
shode pkg add @shode/config ^1.0.0
```

## Usage

```bash
. sh_modules/@shode/config/index.sh

# Load configuration
ConfigLoad "config.json"

# Get values
db_host=$(ConfigGet "DB_HOST" "localhost")
db_port=$(ConfigGet "DB_PORT" "5432")

# Set values
ConfigSet "API_KEY" "secret123"

# Check if key exists
if ConfigHas "API_KEY"; then
    echo "API_KEY is set"
fi
```

## API

### Functions

- `ConfigLoad(file, type)` - Load configuration file
- `ConfigGet(key, default)` - Get configuration value
- `ConfigSet(key, value)` - Set configuration value
- `ConfigHas(key)` - Check if key exists
- `ConfigMerge(priority, files...)` - Merge configs

## License

MIT
