# @shode/logger

Structured logging library for Shode applications.

## Features

- Multiple log levels: debug, info, warn, error
- Configurable output formats: text, JSON
- Easy to use API
- Lightweight and fast

## Installation

```bash
shode pkg add @shode/logger ^1.0.0
```

## Usage

```bash
# Set log level
export LOG_LEVEL=debug

# Use in your scripts
. sh_modules/@shode/logger/index.sh

LogInfo "Application started"
LogWarn "High memory usage"
LogError "Failed to connect to database"
```

## API

### Log Functions

- `LogInfo(message)` - Log an info message
- `LogWarn(message)` - Log a warning message
- `LogError(message)` - Log an error message
- `LogDebug(message)` - Log a debug message

### Configuration

- `SetLogLevel(level)` - Set minimum log level (debug, info, warn, error)
- `AddLogTransport()` - Add a log transport (future)

### Environment Variables

- `LOG_LEVEL` - Minimum log level (default: info)
- `LOG_FORMAT` - Output format: text or json (default: text)

## License

MIT
