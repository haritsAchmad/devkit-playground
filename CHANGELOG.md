# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-08-20

### Added

- `file inspect` command for deterministic content MIME detection, extension consistency checks, size, and SHA-256 metadata.
- `repo inspect` command for deterministic, content-free detection of common repository metadata.
- `text inspect` command for streaming encoding, BOM, UTF-8 validity, line-ending, line-count, and final-newline inspection.
- `port inspect` command for portable, connection-free local TCP availability checks.
- `timestamp convert` command for explicit Unix, Unix-millisecond, and RFC3339 conversion to deterministic UTC output.
- `base64 encode` and `base64 decode` commands with standard/URL-safe alphabets, explicit padding, bounded input, and binary-safe JSON transport.
- `hash verify` command for streaming SHA-256/SHA-512 checksum verification with failure-on-mismatch semantics.
- `capabilities` command with versioned, machine-readable discovery metadata for programs and agents.
- Fuzz targets for JWT, dotenv, JSON, Base64, file-inspection, text-inspection, and timestamp parsing paths.

### Changed

- Distinguished bind-access-denied ports from generic probe failures.
- Made global help/version JSON behavior independent of global-flag order and identified pre-routing JSON errors with the `global` command name.
- Bounded JSON input at 16 MiB and enforced the JWT stdin limit before trimming whitespace.
- Listed all callable commands consistently in root help.
- Grouped human capability output by category and added optional category filtering.

### Tests

- Locked the contract that global flags such as `--json` must precede the command.

## [0.1.0] - 2026-08-19

### Added

- First stable release of the offline-first DevKit CLI.
- Cross-platform binaries and SHA-256 checksums for Windows, Linux, and macOS.

### Changed

- Promoted the command, JSON, error, and exit-code contracts validated in `v0.1.0-rc.1` to the stable `v0.1.0` release.

## [0.1.0-rc.1] - 2026-08-19

### Added

- Initial Go module.
- Initial project documentation and contribution guidelines.
- Baseline Go-oriented ignore rules.
- MIT License.
- Product direction for an offline-first, cross-stack developer toolbox.
- Architecture boundaries and dependency principles.
- Initial v0.1 CLI, JSON, error, and exit-code contracts.
- CLI entry point with help, version, and structured JSON foundations.
- Contract tests for CLI routing and JSON envelopes.
- `uuid` command with secure UUID v4 generation and multi-value output.
- `secret` command with secure base64url and hexadecimal generation.
- `hash` command with streaming SHA-256 and SHA-512 support for files and stdin.
- `jwt inspect` command with safe, unverified header and claims decoding.
- `json pretty` and `json minify` commands with precise number handling.
- `env diff` command for deterministic, value-safe dotenv key comparison.
- Cross-platform GitHub Actions checks for formatting, module tidiness, vetting, race tests, unit tests, and builds.
- Local installation and version-injected build instructions.
- Draft GitHub Release automation for versioned Windows, Linux, and macOS archives with SHA-256 checksums.

### Changed

- Reworked the README and roadmap around concrete v0.1 outcomes and non-goals.
- Expanded contribution guidance for architecture, compatibility, testing, and sensitive data.

[Unreleased]: https://github.com/haritsAchmad/devkit-playground/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/haritsAchmad/devkit-playground/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/haritsAchmad/devkit-playground/compare/v0.1.0-rc.1...v0.1.0
[0.1.0-rc.1]: https://github.com/haritsAchmad/devkit-playground/releases/tag/v0.1.0-rc.1
