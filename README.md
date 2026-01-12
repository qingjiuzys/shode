# Shode - Secure Shell Script Runtime Platform

Shode is a modern shell script runtime platform that solves the inherent chaos, unmaintainability, and security issues of traditional shell scripting. It provides a unified, safe, and high-performance environment for writing and managing automation scripts with a rich ecosystem.

## 🎯 Vision

Transform shell scripting from a manual workshop model to a modern engineering discipline, creating a unified, secure, high-performance platform with a rich ecosystem that serves as the foundation for AI-era operations.

## ✨ Features

### 🎯 Complete Shell Feature Support

Shode provides comprehensive Shell syntax and feature support:

- **Control Flow**: If-Then-Else, For/While loops, Break/Continue statements
- **Pipelines & Redirections**: Full support for `|`, `>`, `>>`, `<`, `2>&1`, `&>` operators
- **Variable System**: Variable assignment, expansion (`$VAR`, `${VAR}`), string concatenation
- **Function System**: Function definition, calling, parameter passing, scope isolation
- **Module System**: Import/export, package.json support, path resolution
- **Annotation System**: `@Annotation` syntax similar to Spring framework
- **Execution Modes**: Interpreted (stdlib), Process (external), Hybrid
- **Security Features**: Command blacklist, sensitive file protection, pattern detection
- **Performance**: Command caching, process pooling, metrics collection

📖 [View Complete Shell Features →](docs/guides/SHELL_FEATURES.md)

### ✅ Phase 1: Core Engine (Completed)
- **CLI Interface**: Comprehensive command-line interface with Cobra
- **Advanced Parser**: Robust shell command parser with quote handling and comment support
- **AST Structure**: Complete Abstract Syntax Tree representation for shell commands
- **Execution Framework**: Ready for execution engine integration
- **Security Foundation**: Architecture prepared for sandbox implementation

### ✅ Phase 2: User Experience & Security (Completed)
- **Standard Library**: Built-in functions for filesystem, network, string operations, environment management
- **Enhanced Security**: Advanced security checker with dangerous command blacklisting, sensitive file protection, and pattern matching
- **Environment Manager**: Complete environment variable management, path manipulation, and session isolation
- **REPL Interface**: Interactive Read-Eval-Print Loop with command history and built-in commands

### ✅ Phase 3: Ecosystem & Extensions (Completed)
- **Package Manager**: Complete dependency management with shode.json configuration
- **Dependency Management**: Support for regular and development dependencies
- **Script Management**: Project script definition and execution
- **Package Installation**: Automatic sh_models creation and package simulation

### ✅ Phase 4: Tools & Integration (Completed)
- **Module System**: Complete module loading and resolution system
- **Export/Import**: Function export detection and module import capabilities
- **Path Resolution**: Support for local files and node_modules packages
- **Module Information**: Comprehensive module metadata and export management

## 🚀 Getting Started

### Installation

```bash
# Build from source
git clone https://gitee.com/com_818cloud/shode.git
cd shode
go build -o shode ./cmd/shode
```

### Basic Usage

```bash
# Run a shell script file (with full execution engine)
./shode run examples/test.sh

# Execute an inline command
./shode exec "echo hello world"

# Execute with pipeline
./shode exec "cat file.txt | grep pattern | wc -l"

# Start interactive REPL session
./shode repl

# Show version information
./shode version

# Get help
./shode --help
```

### Package Management

```bash
# Initialize a new package
./shode pkg init my-project 1.0.0

# Search for packages in registry
./shode pkg search lodash

# Add dependencies (installs from registry)
./shode pkg add lodash 4.17.21
./shode pkg add --dev jest 29.7.0

# Install all dependencies from registry
./shode pkg install

# List dependencies
./shode pkg list

# Publish package to registry
./shode pkg publish

# Manage scripts
./shode pkg script test "echo 'Running tests...'"
./shode pkg run test
```

### Module System

```bash
# Create a module with exports
cat > my-module/index.sh << 'EOF'
#!/bin/sh
export_hello() {
    echo "Hello from module!"
}
export_greet() {
    echo "Greetings, $1!"
}
EOF

# Test module loading (using module-test utility)
go build -o module-test ./cmd/module-test
./module-test
```

## 📁 Project Structure

```
shode/
├── cmd/
│   ├── shode/           # Main CLI application
│   │   └── commands/    # Command implementations (run, exec, repl, pkg, version)
│   ├── parser-test/     # Parser testing utility
│   ├── stdlib-test/     # Standard library testing
│   ├── security-test/   # Security checker testing
│   ├── environment-test/# Environment manager testing
│   ├── repl-test/       # REPL component testing
│   └── module-test/     # Module system testing
├── pkg/
│   ├── parser/          # Shell script parsing
│   ├── types/           # AST type definitions
│   ├── stdlib/          # Standard library implementation
│   ├── sandbox/         # Security checker and sandbox
│   ├── environment/     # Environment variable management
│   ├── repl/            # REPL interactive interface
│   ├── pkgmgr/          # Package manager implementation
│   ├── module/          # Module system implementation
│   └── engine/          # Execution engine (future integration)
├── examples/            # Example shell scripts
├── docs/                # Documentation
└── internal/            # Internal packages
```

## 🛠️ Technology Stack

- **Language**: Go (Golang) 1.21+
- **CLI Framework**: Cobra
- **Parser**: Custom simple parser with tree-sitter integration available
- **Platform**: Cross-platform (macOS, Linux, Windows)
- **Package Management**: Custom shode.json based system
- **Module System**: Custom module resolution and loading

## 🔧 Development Status

**Current Version**: 0.2.0 (Production Ready with Enhanced Features)

### ✅ Completed Features

#### Core Infrastructure
- Project structure and Go module setup
- CLI framework with multiple commands
- Advanced shell command parser
- Complete AST structure implementation

#### Execution Engine (NEW in v0.2.0)
- **Pipeline Support**: True data flow between commands
- **Redirection**: Input/output redirection (>, >>, <, 2>&1, &>)
- **Control Flow**: if-then-else, for loops, while loops
- **Variable Assignment**: Environment variable management
- **Command Caching**: Performance optimization with TTL-based cache
- **Process Pooling**: Reusable process pool for repeated commands
- **Three Execution Modes**: Interpreted, Process, and Hybrid

#### Package Registry (NEW in v0.2.0)
- **Registry Client**: Complete client for package operations
- **Registry Server**: Local/remote registry server
- **Package Search**: Full-text search with keyword filtering
- **Package Publishing**: Publish packages with authentication
- **Package Installation**: Download and install from remote registry
- **Caching**: Intelligent caching with 24-hour TTL
- **Checksum Verification**: SHA256 verification for security

#### User Experience
- File system operations (ReadFile, WriteFile, ListFiles, FileExists)
- String manipulation (Contains, Replace, ToUpper, ToLower, Trim)
- Environment management (GetEnv, SetEnv, WorkingDir, ChangeDir)
- Utility functions (Print, Println, Error, Errorln)
- Path manipulation (GetPath, SetPath, AppendToPath, PrependToPath)

#### Security
- Dangerous command blacklist (rm, dd, mkfs, shutdown, iptables, etc.)
- Sensitive file protection (/etc/passwd, /root/, /boot/, etc.)
- Pattern matching detection (recursive delete, password leaks, shell injection)
- Dynamic rule management and security reporting
- Command-level security checks in execution engine

#### Package Management
- shode.json configuration management
- Dependency and devDependency support
- Script definition and execution
- Remote package installation from registry
- Local package fallback
- sh_models directory structure
- Package search command
- Package publish command

#### Module System
- Module loading and resolution
- Export function detection (export_ prefix)
- Import functionality
- Module information and metadata
- Path resolution for local and sh_models packages

#### Interactive Environment
- REPL with command history
- Built-in command support (cd, pwd, ls, cat, echo, env, history)
- Security integration for all commands
- Standard library function integration

## 📝 License

MIT License - see LICENSE file for details

## 🤝 Contributing

This project is production-ready with advanced features! Contributions and feedback are welcome for:
- Enhanced security features and monitoring
- Additional standard library functions
- IDE plugin development (VSCode, IntelliJ, etc.)
- Package signing and verification
- Cloud-native deployment tools
- Performance optimizations
- Documentation and tutorials

## 🎯 Roadmap

### Completed Phases
- ✅ Phase 1: Core Engine
- ✅ Phase 2: User Experience & Security
- ✅ Phase 3: Ecosystem & Extensions
- ✅ Phase 4: Tools & Integration

### Latest Enhancements (v0.2.0)
- ✅ **Complete Execution Engine**: Full pipeline, redirection, and control flow support
- ✅ **Package Registry**: Complete package repository with search and publish
- ✅ **Remote Package Management**: Download and install packages from registry
- ✅ **Enhanced Performance**: Command caching and process pooling

### Future Enhancements
- **Enhanced Security**: Real-time security monitoring and policy enforcement
- **Cloud Integration**: Cloud-native deployment and management
- **AI Assistance**: AI-powered script generation and optimization
- **Package Signing**: Cryptographic verification for packages
- **IDE Integration**: VSCode and other IDE plugins

## 🌟 Why Shode?

Shode addresses the fundamental problems with traditional shell scripting:

1. **Security**: Prevents dangerous operations and protects sensitive systems
2. **Maintainability**: Provides modern code organization and dependency management
3. **Portability**: Cross-platform compatibility with consistent behavior
4. **Productivity**: Rich standard library and development tools
5. **Modernization**: Brings shell scripting into the modern development era

Shode is now ready for production use and represents a significant step forward in shell script development and operations automation.
</content>
