# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive unit tests covering all public and private functions of the library.

## [0.1.0] - 2025-05-03

### Added
- Initial release of the `goconfig` library.
- Support for loading configuration from YAML files.
- Automatic environment variable substitution using the `${VAR_NAME}` syntax.
- `.env` file loading via `LoadEnv()`.
- Configurable options: `WithConfigDir()` and `WithUnmarshaller()`.
- Minimal API: `New()`, `Parse()`, `LoadEnv()`.
- Exported errors: `ErrVariableNotFound`, `ErrInvalidEnvFormat`.

[Unreleased]: https://github.com/salomondevsystems/goconfig/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/salomondevsystems/goconfig/releases/tag/v0.1.0
