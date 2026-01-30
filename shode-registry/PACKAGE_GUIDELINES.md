# Shode Package Development Guidelines

This document provides guidelines for developing Shode packages, especially for official `@shode` scoped packages.

## Package Structure

A well-structured Shode package should have the following layout:

```
package-name/
├── package.json          # Package metadata (required)
├── index.sh             # Entry point (required)
├── README.md            # Documentation (required)
├── LICENSE              # License file (recommended)
├── src/                 # Source code
│   └── *.sh            # Implementation files
├── tests/               # Test files
│   └── *_test.sh
└── examples/            # Usage examples (optional)
    └── *.sh
```

## package.json Format

Every package must have a `package.json` file:

```json
{
  "name": "@shode/package-name",
  "version": "1.0.0",
  "description": "Brief description of the package",
  "main": "index.sh",
  "scripts": {
    "test": "shode test"
  },
  "dependencies": {},
  "devDependencies": {},
  "keywords": ["keyword1", "keyword2"],
  "author": "Your Name",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/user/repo"
  }
}
```

### Required Fields

- `name` - Package name (use `@shode/` prefix for official packages)
- `version` - Semantic version (MAJOR.MINOR.PATCH)
- `description` - Brief description
- `main` - Entry point file
- `author` - Author name
- `license` - License type

### Optional but Recommended Fields

- `scripts` - Available scripts (test, build, etc.)
- `dependencies` - Runtime dependencies
- `devDependencies` - Development dependencies
- `keywords` - Search keywords
- `repository` - Source code repository
- `homepage` - Project homepage

## Shell Script Guidelines

### Shebang

All executable scripts should start with:
```bash
#!/bin/sh
```

### Error Handling

Use proper error handling:
```bash
my_function() {
    if [ $# -lt 1 ]; then
        echo "Error: Missing required argument" >&2
        return 1
    fi
    
    # Your code here
}
```

### Namespace and Exporting

Avoid polluting the global namespace. Export only public APIs:
```bash
# Private function (prefixed with _)
_my_internal_function() {
    # Internal logic
}

# Public function (exported in index.sh)
PublicAPI() {
    _my_internal_function
}
```

### Documentation

Document all public functions:
```bash
# FunctionName does something useful
# Arguments:
#   $1 - First argument description
#   $2 - Second argument description
# Returns:
#   0 on success, 1 on error
FunctionName() {
    # Implementation
}
```

## Testing Guidelines

### Test File Naming

Test files should be named: `<module>_test.sh`

### Test Structure

```bash
#!/bin/sh
# test/example_test.sh

# Test setup
setup() {
    echo "Setting up tests..."
}

# Test teardown
teardown() {
    echo "Cleaning up..."
}

# Test cases
test_case_name() {
    # Arrange
    local input="test"
    local expected="output"
    
    # Act
    local result=$(MyFunction "$input")
    
    # Assert
    if [ "$result" != "$expected" ]; then
        echo "FAIL: test_case_name"
        echo "  Expected: $expected"
        echo "  Got: $result"
        return 1
    fi
    
    echo "PASS: test_case_name"
}

# Run tests
setup
test_case_name
teardown
```

## README Guidelines

Every package should have a comprehensive README.md:

### Required Sections

1. **Title and brief description**
2. **Installation instructions**
3. **Usage examples**
4. **API documentation**

### Recommended Sections

5. **Features**
6. **Configuration options**
7. **Dependencies**
8. **Examples**
9. **Contributing**
10. **License**

### README Template

```markdown
# @shode/package-name

Brief description of what the package does.

## Features

- Feature 1
- Feature 2
- Feature 3

## Installation

\`\`\`bash
shode pkg add @shode/package-name ^1.0.0
\`\`\`

## Usage

\`\`\`bash
. sh_modules/@shode/package-name/index.sh

PackageName "argument"
\`\`\`

## API

### Functions

- `FunctionName(args)` - Description
- `OtherFunction(args)` - Description

## Configuration

Environment variables or configuration options.

## Examples

More detailed usage examples.

## License

MIT
```

## Versioning

Shode packages follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version - Incompatible API changes
- **MINOR** version - Backwards-compatible functionality additions
- **PATCH** version - Backwards-compatible bug fixes

Example: `1.2.3`

## Publishing

### Pre-publish Checklist

- [ ] All tests pass
- [ ] README is complete
- [ ] package.json is valid
- [ ] No sensitive information included
- [ ] License file is present

### Publishing to Registry

\`\`\`bash
# Set registry token
export SHODE_TOKEN=your_token

# Publish package
shode pkg publish
\`\`\`

## Code Style

### Indentation

Use 4 spaces for indentation (no tabs).

### Line Length

Keep lines under 100 characters when possible.

### Comments

- Use `#` for single-line comments
- Document complex logic
- Comment why, not what (code should be self-explanatory)

## Security Considerations

### Input Validation

Always validate user input:
```bash
safe_input=$(echo "$user_input" | tr -d ';&|><$()')
```

### Command Injection Prevention

Avoid eval with user input. Use arrays instead:
```bash
# Bad
eval "command $user_input"

# Good
command "$user_input"
```

### File Operations

Validate file paths to prevent traversal attacks:
```bash
if [[ "$file_path" == *".."* ]]; then
    echo "Error: Invalid file path" >&2
    return 1
fi
```

## Official Package Requirements

In addition to the guidelines above, official `@shode` packages must:

1. **Comprehensive Testing**: Minimum 80% test coverage
2. **Documentation**: Complete README with examples
3. **Code Quality**: Follow style guidelines
4. **Security**: Pass security review
5. **Peer Review**: Approved by maintainers
6. **Maintenance**: Active maintenance and support

## Support

For questions or help with package development:
- Documentation: https://docs.shode.io
- GitHub Issues: https://github.com/shode/packages/issues
- Community: https://discord.gg/shode

---

© 2026 Shode Project
