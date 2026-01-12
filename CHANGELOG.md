# Changelog

All notable changes to Shode will be documented in this file.

## [0.3.0] - 2025-01-XX

### 🚀 Production Ready Release

#### Testing & Quality
- **Comprehensive Test Suite**: Added tests for all core modules
  - Registry client tests (tarball extraction, path security)
  - Security checker tests (dangerous commands, file protection)
  - Environment manager tests (variables, working directory)
  - Parser tests (string parsing, file parsing)
  - Boundary tests (large files, concurrency, resource limits)
  - Integration tests (full workflow testing)
- **Test Coverage**: Achieved >80% test coverage across all packages
- **Performance Benchmarks**: Comprehensive benchmark suite for engine and package manager

#### Error Handling System
- **Unified Error Types**: Complete error type system (`pkg/errors`)
  - Security violations, command not found, execution failures
  - Parse errors, timeouts, file not found
  - Error context and stack traces
  - Error wrapping and chaining
- **Error Recovery**: Robust error recovery mechanisms
  - Timeout handling with context cancellation
  - Resource cleanup (defer cleanup, process management)
  - Graceful degradation (cache fallback, partial failure handling)
  - Pipeline error propagation

#### Performance Monitoring
- **Metrics Collection**: Comprehensive metrics system (`pkg/metrics`)
  - Command execution metrics (success/failure/timeout rates)
  - Performance metrics (duration percentiles P50/P95/P99)
  - Cache metrics (hit/miss rates)
  - Process pool metrics
  - Error statistics
  - Pipeline and loop execution metrics
- **Performance Optimization**
  - Optimized cache eviction algorithm
  - Reduced lock contention in cache operations
  - Memory-efficient cache management

#### Developer Tools
- **Debug Tool** (`cmd/shode-debug`): Script debugging with metrics
- **Profile Tool** (`cmd/shode-profile`): Performance profiling and analysis

#### Documentation
- **User Guide** (`docs/USER_GUIDE.md`): Complete user documentation
  - Quick start guide
  - Common use cases
  - Best practices
  - Troubleshooting guide
  - Example scripts

### 🔧 Enhanced Features

#### Execution Engine
- Enhanced timeout handling with proper resource cleanup
- Improved error context and reporting
- Better cache management with optimized eviction

#### Security
- Enhanced security error reporting
- Better error context for security violations

### 📚 Documentation
- Added comprehensive user guide
- Enhanced API documentation
- Added troubleshooting guide
- Performance optimization guide

## [0.2.0] - 2025-10-04

### 🚀 Major Features

#### Execution Engine (Complete)
- **Pipeline Support**: Full implementation of command pipelines with true data flow
  - Commands can be chained with `|` operator
  - Output of one command flows as input to the next
  - Proper error handling and failure propagation
  
- **I/O Redirection**: Complete support for all standard redirection operators
  - Output redirection: `>` (overwrite), `>>` (append)
  - Input redirection: `<`
  - Error redirection: `2>&1`, `&>`
  - File descriptor support
  
- **Control Flow**: Full support for shell control structures
  - `if-then-else` statements with condition evaluation
  - `for` loops with variable iteration
  - `while` loops with safety limits (max 10,000 iterations)
  - Proper variable scoping
  
- **Performance Optimizations**
  - Command result caching with TTL-based expiration
  - Process pooling for repeated commands
  - Configurable cache size and timeout
  - Automatic cleanup of idle resources

#### Package Registry System (Complete)
- **Registry Client** (`pkg/registry/client.go`)
  - Search packages by name, keywords, author
  - Download packages from remote registry
  - Install packages with automatic extraction
  - Publish packages with authentication
  - Checksum verification (SHA256)
  
- **Registry Server** (`pkg/registry/server.go`)
  - HTTP API server for package operations
  - Package metadata management
  - Full-text search with relevance scoring
  - Authentication and authorization
  - Download statistics tracking
  - Verified package badges
  
- **Caching System** (`pkg/registry/cache.go`)
  - Metadata caching with 24-hour TTL
  - Tarball caching with disk management
  - Automatic cleanup of expired entries
  - Cache statistics and monitoring

### 📦 New Commands

#### Execution Commands
- `shode run <script>` - Now with full execution engine support
- `shode exec <command>` - Enhanced with pipeline and redirection support

#### Package Registry Commands
- `shode pkg search <query>` - Search for packages in the registry
- `shode pkg publish` - Publish package to the registry

### 🔧 Enhanced Features

#### Package Manager
- Remote package installation from registry
- Fallback to local installation if registry unavailable
- Improved error handling and reporting
- Registry client integration

#### Security
- Command-level security checks in execution engine
- Context-based timeout support
- Secure tarball verification

#### Documentation
- Complete execution engine documentation (`docs/EXECUTION_ENGINE.md`)
- Complete package registry documentation (`docs/PACKAGE_REGISTRY.md`)
- Updated README with new features
- Example scripts demonstrating new features

### 🐛 Bug Fixes
- Fixed file redirection resource cleanup
- Improved error handling in pipeline execution
- Fixed cache key generation for parameterized commands

### ⚡ Performance Improvements
- Command caching reduces repeated execution overhead
- Process pooling improves performance for shell-out commands
- Efficient pipeline implementation with proper streaming

### 📚 Documentation
- Added `docs/EXECUTION_ENGINE.md` - Complete execution engine guide
- Added `docs/PACKAGE_REGISTRY.md` - Complete package registry guide
- Added `examples/advanced_features.sh` - Demonstration script
- Updated `README.md` with v0.2.0 features
- Added `CHANGELOG.md` - This file

### 🔄 API Changes

#### New Types (`pkg/types/ast.go`)
- `IfNode` - If-then-else statement representation
- `ForNode` - For loop representation
- `WhileNode` - While loop representation
- `FunctionNode` - Function definition representation
- `AssignmentNode` - Variable assignment representation

#### New Packages
- `pkg/registry` - Complete package registry implementation
  - `types.go` - Registry data types
  - `client.go` - Registry client
  - `server.go` - Registry server
  - `cache.go` - Cache manager

#### Enhanced Packages
- `pkg/engine/engine.go` - Complete execution engine implementation
- `pkg/pkgmgr/manager.go` - Registry integration
- `cmd/shode/commands/` - Enhanced commands

## [0.1.0] - 2024-12-XX

### Initial Release

#### Core Features
- CLI framework with Cobra
- Shell script parser with tree-sitter
- AST structure for shell commands
- Security checker with blacklisting
- Standard library for common operations
- Environment manager
- REPL interactive environment
- Package manager with shode.json
- Module system with export/import

#### Commands
- `shode run` - Run script files
- `shode exec` - Execute inline commands
- `shode repl` - Interactive REPL
- `shode pkg init` - Initialize package
- `shode pkg add` - Add dependencies
- `shode pkg install` - Install dependencies
- `shode pkg list` - List dependencies
- `shode pkg run` - Run package scripts
- `shode version` - Show version

---

## Version Numbering

Shode follows [Semantic Versioning](https://semver.org/):
- MAJOR version for incompatible API changes
- MINOR version for new functionality in a backward compatible manner
- PATCH version for backward compatible bug fixes

## Links

- [GitHub Repository](https://gitee.com/com_818cloud/shode)
- [Documentation](./docs/)
- [Issues](https://gitee.com/com_818cloud/shode/issues)
