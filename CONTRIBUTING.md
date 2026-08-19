# Contributing

Thanks for helping improve DevKit Playground. The project is pre-1.0, but changes should still be easy to understand, verify, and reverse.

## Before starting

For a substantial change, open an issue or discussion first so the problem and intended scope can be agreed on. Small fixes can go directly to a pull request.

## Local setup

```sh
git clone https://github.com/haritsAchmad/devkit-playground.git
cd devkit-playground
go test ./...
```

## Making a change

1. Create a focused branch from `main`.
2. Keep the change narrowly scoped.
3. Add or update tests when behavior changes.
4. Update the README, roadmap, or changelog when relevant.
5. Run the local checks before opening a pull request.

```sh
go fmt ./...
go vet ./...
go test ./...
```

## Architecture rules

Changes must preserve the boundaries described in [docs/architecture.md](docs/architecture.md):

- Keep tool logic independent from CLI argument parsing and rendering.
- Return typed results and errors from tool packages; do not print or terminate the process there.
- Keep human and JSON renderers separate.
- Do not introduce network access into core utilities.
- Prefer the Go standard library and justify every new dependency.
- Do not add abstractions, adapters, or extension systems without a demonstrated consumer.

New commands must have a concrete recurring use case. A command proposal should include its inputs, outputs, failure behavior, security considerations, and why an existing tool does not already solve the problem.

## Command contracts

[docs/command-contracts.md](docs/command-contracts.md) is the source of truth for CLI behavior. A command change should include tests for:

- Domain behavior without launching a subprocess.
- Human-readable success and error output.
- JSON success and error envelopes.
- Exit-code mapping.
- File and stdin input where supported.
- Deterministic ordering where applicable.
- Sensitive-data redaction.

Additive JSON fields are allowed during `0.x` development. Removing, renaming, changing the type of, or changing the meaning of a documented field is a breaking change and must be clearly documented.

## Sensitive data

- Never commit real secrets, tokens, credentials, private keys, or populated `.env` files.
- Test fixtures must use obviously fake values.
- Do not include generated secrets, raw JWTs, JWT signatures, environment values, or file contents in errors or logs.
- Prefer stdin in documentation when an argument may contain sensitive data.

## Commit messages

Use short, imperative commit messages. Conventional Commit prefixes are encouraged:

- `feat:` for new behavior
- `fix:` for bug fixes
- `docs:` for documentation-only changes
- `test:` for test changes
- `refactor:` for internal changes without behavior changes
- `chore:` for maintenance work

## Pull requests

A pull request should describe:

- The problem being solved.
- The chosen approach and meaningful tradeoffs.
- How the change was tested.
- Any follow-up work intentionally left out of scope.

Avoid mixing unrelated refactors with functional changes.

## Changelog

Add user-visible changes under `Unreleased` in [CHANGELOG.md](CHANGELOG.md). Contract and security changes must be called out explicitly. Purely internal changes do not need an entry unless they affect contributors or the release process.
