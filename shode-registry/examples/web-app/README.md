# Shode Web Application Example

This is an example web application demonstrating the use of Shode official packages.

## Features

- Structured logging with `@shode/logger`
- Configuration management with `@shode/config`
- HTTP client with `@shode/http`

## Installation

1. Install dependencies:
```bash
shode pkg install
```

2. Run the application:
```bash
shode pkg run start
```

## Project Structure

```
.
├── shode.json           # Package configuration
├── config.json          # Application configuration
├── src/
│   └── main.sh         # Main application entry point
└── README.md           # This file
```

## Usage

The application demonstrates:
- Loading and using official packages
- Configuration management
- Structured logging
- Making HTTP requests

## Extending

You can add more official packages:
```bash
shode pkg add @shode/cron ^1.0.0
shode pkg add @shode/database ^1.0.0
```

## License

MIT
