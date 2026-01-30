# Shode Official Package Registry

This is the official package registry for Shode, containing curated packages developed and maintained by the Shode team.

## Official Packages

### Core Packages

#### [@shode/logger](./packages/@shode/logger/)
Structured logging library with support for multiple log levels and formats.

```bash
shode pkg add @shode/logger ^1.0.0
```

#### [@shode/config](./packages/@shode/config/)
Configuration management library supporting JSON, ENV, and Shell formats.

```bash
shode pkg add @shode/config ^1.0.0
```

#### [@shode/cron](./packages/@shode/cron/)
Cron-like task scheduling for Shode applications.

```bash
shode pkg add @shode/cron ^1.0.0
```

#### [@shode/http](./packages/@shode/http/)
HTTP client library for making web requests.

```bash
shode pkg add @shode/http ^1.0.0
```

#### [@shode/database](./packages/@shode/database/)
Database abstraction layer supporting MySQL, PostgreSQL, and SQLite.

```bash
shode pkg add @shode/database ^1.0.0
```

## Package Naming Convention

Official packages use the `@shode` scope:
```
@shode/<package-name>
```

## Contributing

We welcome contributions to the official package registry! Please see our [Contributing Guidelines](CONTRIBUTING.md) for more information.

## Publishing Packages

To publish a package to the official registry:

1. Ensure your package follows our [Package Guidelines](PACKAGE_GUIDELINES.md)
2. Include tests and documentation
3. Submit a pull request to this repository

## Support

For issues or questions about official packages:
- GitHub Issues: https://github.com/shode/packages/issues
- Documentation: https://docs.shode.io
- Community: https://discord.gg/shode

## License

All official packages are released under the MIT License unless otherwise specified.

---

© 2026 Shode Project
