# DevKit Playground

DevKit Playground is a lightweight, offline-first developer toolbox written in Go. It collects small, deterministic utilities that developers repeatedly need while working across Go, PHP/Symfony, JavaScript/Nuxt, SQL, and general backend development.

DevKit is implemented in Go for simple distribution and predictable behavior. It is not limited to Go projects.

> [!IMPORTANT]
> DevKit is in early development. Command and JSON contracts are versioned but are not yet stable until the project reaches `v1.0.0`.

## Who it is for

DevKit is designed for two consumers.

### Developers

Developers can use the CLI directly for recurring tasks:

```sh
devkit uuid
devkit secret
devkit hash ./artifact.zip
devkit jwt inspect < token.txt
devkit json pretty response.json
devkit env diff .env.example .env
```

### Programs and AI agents

Commands will support structured JSON output so programs and future local AI agents can use the same deterministic tools without parsing terminal-oriented text:

```sh
devkit --json uuid
```

DevKit is not an AI agent. It may eventually serve as a deterministic tool layer for one.

## Design principles

- Keep tool logic independent from CLI presentation.
- Expose reusable Go logic behind each command.
- Separate human-readable output from machine-readable JSON output.
- Prefer deterministic code and the Go standard library.
- Keep dependencies minimal and justified.
- Keep core utilities fully functional without network access.
- Avoid exposing secrets and sensitive values unnecessarily.
- Use meaningful, consistent exit codes.
- Add utilities only when they solve a concrete recurring problem.
- Avoid premature infrastructure and abstraction.

## Planned v0.1 commands

| Command | Purpose |
|---|---|
| `devkit uuid` | Generate cryptographically secure UUIDs |
| `devkit secret` | Generate cryptographically secure secrets |
| `devkit hash` | Hash a file or standard input |
| `devkit jwt inspect` | Decode JWT metadata and claims without verification |
| `devkit json pretty` | Format JSON for readability |
| `devkit json minify` | Remove insignificant JSON whitespace |
| `devkit env diff` | Compare environment-file key sets without exposing values |

The precise interfaces are defined in [Command contracts](docs/command-contracts.md).

## Non-goals

DevKit is not:

- An AI agent or an LLM-powered application.
- A replacement for compilers, language-specific linters, or IDEs.
- A cloud service.
- A collection of novelty commands.
- Dependent on Docker, a database, Node.js, or an LLM for core functionality.

## Requirements

- Go 1.26.1 or newer for development from source
- Git

No runtime dependency beyond the compiled DevKit binary is intended for core commands.

## Development setup

```sh
git clone https://github.com/haritsAchmad/devkit-playground.git
cd devkit-playground
go test ./...
```

The repository currently provides the CLI foundation, versioned output envelope, UUID generation, and secure secret generation. Other v0.1 utility commands are not implemented yet.

Before submitting a code change, run:

```sh
go fmt ./...
go vet ./...
go test ./...
```

## Documentation

- [Architecture](docs/architecture.md) — boundaries and design constraints
- [Command contracts](docs/command-contracts.md) — CLI, JSON, error, and exit-code behavior
- [Roadmap](ROADMAP.md) — milestones and acceptance criteria
- [Contributing guide](CONTRIBUTING.md) — contribution and quality conventions
- [Changelog](CHANGELOG.md) — notable changes

## License

This project is licensed under the [MIT License](LICENSE).
