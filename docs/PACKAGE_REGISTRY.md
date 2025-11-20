# Shode Package Registry

## Overview

The Shode Package Registry is a complete package management system for sharing and distributing shell script packages. It provides features similar to npm, PyPI, or RubyGems but specifically designed for shell scripts.

## Features

### 1. Package Publishing
- Publish packages to the registry
- Version management
- Metadata storage
- Checksum verification
- Authentication and authorization
- Ed25519 package signatures with trust-store based verification
- Cloud deployment mode（PostgreSQL + S3），参考 `docs/CLOUD_REGISTRY.md`

### 2. Package Discovery
- Full-text search
- Keyword filtering
- Author filtering
- Download statistics
- Verified package badges

### 3. Package Installation
- Automatic dependency resolution
- Local caching
- Checksum verification
- Fallback to local installation

### 4. Package Management
- Add/remove dependencies
- Dev dependencies support
- Script management
- Package listing

## Architecture

### Components

#### Registry Client (`pkg/registry/client.go`)
Handles all client-side operations:
- Search packages
- Download packages
- Install packages
- Publish packages
- Cache management

#### Registry Server (`pkg/registry/server.go`)
Local/remote registry server:
- HTTP API endpoints
- Package storage
- Metadata management
- Authentication
- Search indexing

#### Cache Manager (`pkg/registry/cache.go`)
Manages local package cache:
- Metadata caching (24-hour TTL)
- Tarball caching
- Disk usage tracking
- Automatic cleanup

## Usage

### Command Line Interface

#### Initialize Package

```bash
./shode pkg init my-package 1.0.0
```

Creates `shode.json`:
```json
{
  "name": "my-package",
  "version": "1.0.0",
  "dependencies": {},
  "devDependencies": {},
  "scripts": {}
}
```

#### Add Dependencies

```bash
# Regular dependency
./shode pkg add lodash 4.17.21

# Dev dependency
./shode pkg add --dev jest 29.7.0
```

#### Install Dependencies

```bash
./shode pkg install
# 或者允许安装未签名包（不推荐）
./shode pkg install --allow-unsigned
```

This will:
1. Read `shode.json`
2. Try to download from registry
3. Fallback to local installation if registry unavailable
4. Create `sh_modules/` directory
5. Extract packages to `sh_modules/package-name/`

#### Search Packages

```bash
./shode pkg search lodash
```

Output:
```
Found 3 package(s):

1. lodash@4.17.21
   Description: A modern JavaScript utility library
   Author: John Doe
   Keywords: [utility, helper, functions]
   Downloads: 1500 ✓ Verified

2. lodash-es@4.17.21
   Description: Lodash exported as ES modules
   Author: John Doe
   Downloads: 800

3. mini-lodash@1.0.0
   Description: Minimal lodash implementation
   Author: Jane Smith
   Downloads: 250
```

#### Publish Package

```bash
# 指定签名者 ID，自动在 ~/.shode/keys/<signer>.ed25519 中寻找密钥
./shode pkg publish --signer my-team

# 指定密钥路径
./shode pkg publish --signer my-team --key ~/.shode/keys/my-team.ed25519
```

Requirements:
- Valid `shode.json` file
- Authentication token (automatically managed)
- Package files ready
- Signing key (Ed25519). Use `shode pkg signer generate alice` to create one.

#### List Dependencies

```bash
./shode pkg list
```

Output:
```
Dependencies:
  lodash: 4.17.21
  request: 2.88.2

Dev Dependencies:
  jest: 29.7.0
  eslint: 8.0.0
```

#### Manage Scripts

```bash
# Add script
./shode pkg script test "echo 'Running tests...'"

# Run script
./shode pkg run test
```

### Signature & Trust Store Management

Use the new `shode pkg signer` namespace来管理密钥与信任：

```bash
# 生成 Ed25519 密钥对，存放于 ~/.shode/keys/
./shode pkg signer generate my-team

# 查看本地已有密钥
./shode pkg signer keys

# 将某个公钥加入信任列表（~/.shode/trust/trusted_signers.json）
./shode pkg signer trust team-b ./public_keys/team-b.pub --desc "Partner team"

# 查看/移除信任
./shode pkg signer trusted
./shode pkg signer untrust team-b
```

安装时会自动读取信任列表，只有来自受信签名者的包才会被接受（除非显式传入 `--allow-unsigned`）。

### Programmatic Usage

#### Create Registry Client

```go
import "gitee.com/com_818cloud/shode/pkg/registry"

// Use default configuration
client, err := registry.NewClient(nil)

// Or custom configuration
config := &registry.RegistryConfig{
    URL:      "https://registry.shode.io",
    Token:    "your-auth-token",
    CacheDir: "/path/to/cache",
    Timeout:  30, // seconds
    TrustStorePath: "/path/to/trusted_signers.json",
    AllowUnsigned:  false,
}
client, err := registry.NewClient(config)
```

#### Search Packages

```go
query := &registry.SearchQuery{
    Query:  "lodash",
    Limit:  10,
    Offset: 0,
}

results, err := client.Search(query)
for _, result := range results {
    fmt.Printf("%s@%s - %s\n", 
        result.Name, 
        result.Version, 
        result.Description)
}
```

#### Get Package Metadata

```go
metadata, err := client.GetPackage("lodash")
if err != nil {
    // Handle error
}

fmt.Printf("Latest version: %s\n", metadata.LatestVersion)
fmt.Printf("Versions: %v\n", len(metadata.Versions))
```

#### Download and Install

```go
// Download to cache
tarballPath, err := client.Download("lodash", "4.17.21")

// Install to directory
err = client.Install("lodash", "4.17.21", "./sh_modules")
```

#### Publish Package

```go
pkg := &registry.Package{
    Name:        "my-package",
    Version:     "1.0.0",
    Description: "My awesome package",
    Author:      "Your Name",
    Main:        "index.sh",
    Dependencies: map[string]string{
        "lodash": "4.17.21",
    },
}

// Create tarball (simplified example)
tarballData := createTarball("./my-package")
checksum := calculateChecksum(tarballData)

req := &registry.PublishRequest{
    Package:  pkg,
    Tarball:  tarballData,
    Checksum: checksum,
}

err := client.Publish(req)
```

### Starting Local Registry Server

```go
import "gitee.com/com_818cloud/shode/pkg/registry"

server, err := registry.NewServer("./registry-data", 8080)
if err != nil {
    log.Fatal(err)
}

// Get auth token for publishing
token := server.GetAuthToken()
fmt.Printf("Auth Token: %s\n", token)

// Start server
log.Fatal(server.Start())
```

Server endpoints:
- `POST /api/search` - Search packages
- `GET /api/packages/{name}` - Get package metadata
- `POST /api/packages` - Publish package (requires auth)
- `GET /health` - Health check

## Package Structure

### shode.json Format

```json
{
  "name": "package-name",
  "version": "1.0.0",
  "description": "Package description",
  "author": "Author Name",
  "license": "MIT",
  "homepage": "https://github.com/user/repo",
  "repository": "https://github.com/user/repo.git",
  "keywords": ["utility", "helper"],
  "main": "index.sh",
  "dependencies": {
    "dependency1": "1.0.0",
    "dependency2": "^2.0.0"
  },
  "devDependencies": {
    "test-framework": "1.0.0"
  },
  "scripts": {
    "test": "run-tests.sh",
    "build": "build-package.sh"
  }
}
```

### Package Directory Structure

```
my-package/
├── shode.json          # Package configuration
├── index.sh            # Main entry point
├── lib/                # Library code
│   ├── utils.sh
│   └── helpers.sh
├── tests/              # Test files
│   └── test.sh
└── README.md          # Documentation
```

## Security

### Package Verification

All packages undergo security checks:
- Checksum verification
- Signature validation (future)
- Verified badge for trusted publishers

### Authentication

Publishing requires authentication:
```bash
# Token is automatically generated and managed
# For custom token:
export SHODE_REGISTRY_TOKEN="your-token"
```

### Package Review

Verified packages are reviewed for:
- Security vulnerabilities
- Malicious code
- Best practices compliance
- Documentation quality

## Caching

### Metadata Cache
- TTL: 24 hours
- Location: `~/.shode/cache/metadata/`
- Format: JSON files

### Tarball Cache
- No expiration (manual cleanup)
- Location: `~/.shode/cache/`
- Format: `.tar.gz` files

### Cache Management

```go
// Get cache statistics
stats := client.cache.GetCacheStats()
fmt.Printf("Disk usage: %.2f MB\n", stats["disk_usage_mb"])

// Clear cache
err := client.cache.Clear()

// Clean expired entries
client.cache.CleanExpired()
```

## API Reference

### Registry Client Methods

```go
type Client struct {
    // Search for packages
    Search(query *SearchQuery) ([]*SearchResult, error)
    
    // Get package metadata
    GetPackage(name string) (*PackageMetadata, error)
    
    // Get specific version
    GetPackageVersion(name, version string) (*PackageVersion, error)
    
    // Download package tarball
    Download(name, version string) (string, error)
    
    // Publish package
    Publish(req *PublishRequest) error
    
    // Install package
    Install(name, version, targetDir string) error
    
    // Set authentication token
    SetToken(token string)
}
```

### Data Types

#### Package
```go
type Package struct {
    Name            string
    Version         string
    Description     string
    Author          string
    License         string
    Main            string
    Dependencies    map[string]string
    DevDependencies map[string]string
    // ... more fields
}
```

#### SearchQuery
```go
type SearchQuery struct {
    Query    string
    Keywords []string
    Author   string
    Limit    int
    Offset   int
}
```

#### SearchResult
```go
type SearchResult struct {
    Name        string
    Version     string
    Description string
    Author      string
    Downloads   int
    Verified    bool
    Score       float64
}
```

## Best Practices

### Package Development

1. **Use Semantic Versioning**
   ```
   MAJOR.MINOR.PATCH
   1.0.0 -> 1.0.1 -> 1.1.0 -> 2.0.0
   ```

2. **Document Your Package**
   - Clear README.md
   - Usage examples
   - API documentation

3. **Test Before Publishing**
   ```bash
   ./shode pkg script test "run-all-tests.sh"
   ./shode pkg run test
   ```

4. **Use Meaningful Keywords**
   ```json
   "keywords": ["utility", "string", "manipulation", "helper"]
   ```

5. **Keep Dependencies Minimal**
   - Only include necessary dependencies
   - Use devDependencies for development tools

### Package Consumption

1. **Pin Critical Dependencies**
   ```json
   "dependencies": {
     "core-lib": "1.2.3"  // Exact version
   }
   ```

2. **Use Ranges for Flexible Updates**
   ```json
   "dependencies": {
     "utils": "^1.0.0"  // Compatible updates
   }
   ```

3. **Regular Updates**
   ```bash
   # Check for updates
   ./shode pkg list
   
   # Update dependencies
   ./shode pkg install
   ```

4. **Review Package Before Use**
   ```bash
   # Search and review
   ./shode pkg search package-name
   
   # Check verified status
   # Look at download count
   ```

## Troubleshooting

### Registry Connection Issues

```bash
# Test registry connection
curl http://localhost:8080/health

# Check cache
ls ~/.shode/cache/

# Clear cache and retry
rm -rf ~/.shode/cache/
./shode pkg install
```

### Package Not Found

1. Check package name spelling
2. Verify registry URL in config
3. Try searching: `./shode pkg search package-name`
4. Check network connectivity

### Checksum Mismatch

1. Clear cache: `rm -rf ~/.shode/cache/`
2. Retry installation
3. Report to package maintainer if persists

### Authentication Failures

1. Check token: `echo $SHODE_REGISTRY_TOKEN`
2. Regenerate token from server
3. Set token: `export SHODE_REGISTRY_TOKEN="new-token"`

## Future Enhancements

- [ ] Package dependencies auto-resolution
- [ ] Semantic version range support
- [ ] Package signing and verification
- [ ] Private registry support
- [ ] Package deprecation warnings
- [ ] Automatic security scanning
- [ ] Package migration tools
- [ ] CDN integration for tarball delivery
- [ ] Package analytics dashboard
- [ ] Multi-registry support
