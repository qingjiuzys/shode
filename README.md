# Shode - Secure Shell Script Runtime Platform

Shode is a modern shell script runtime platform that solves the inherent chaos, unmaintainability, and security issues of traditional shell scripting. It provides a unified, safe, and high-performance environment for writing and managing automation scripts.

## 🎯 Vision

Transform shell scripting from a manual workshop model to a modern engineering discipline, creating a unified, secure, high-performance platform that serves as the foundation for AI-era operations.

## ✨ Core Features

- **Complete Shell Syntax**: Control flow, pipelines, redirections, variables, functions
- **Execution Engine**: Full support for pipelines, redirections, control flow, variable assignment
- **Package Management**: Dependency management with `shode.json` and package registry
- **Module System**: Import/export modules, support for local and remote packages
- **Security Sandbox**: Command blacklist, sensitive file protection, pattern detection
- **Standard Library**: Built-in functions for filesystem, network, string, environment management
- **Interactive REPL**: Command history and built-in commands

## 🚀 Getting Started

### Installation

```bash
git clone https://gitee.com/com_818cloud/shode.git
cd shode
go build -o shode ./cmd/shode
```

### Basic Usage

```bash
# Run a script file
./shode run examples/test.sh

# Execute a command
./shode exec "echo hello world"

# Interactive REPL
./shode repl

# Package management
./shode pkg init my-project 1.0.0
./shode pkg add lodash 4.17.21
./shode pkg install
```

## 📁 Project Structure

```
shode/
├── cmd/shode/          # Main CLI application
├── pkg/                # Core packages (parser, engine, stdlib, sandbox, etc.)
├── examples/           # Example scripts
└── docs/               # Documentation
```

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Platform**: Cross-platform (macOS, Linux, Windows)

## 📝 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions and feedback are welcome! The project is production-ready.

## 🌟 Why Shode?

1. **Security**: Prevents dangerous operations and protects sensitive systems
2. **Maintainability**: Modern code organization and dependency management
3. **Portability**: Cross-platform compatibility with consistent behavior
4. **Productivity**: Rich standard library and development tools
5. **Modernization**: Brings shell scripting into the modern development era
