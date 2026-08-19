# Roadmap

DevKit Playground is an offline-first developer toolbox written in Go. This roadmap is usage-driven: a utility is added only when it solves a concrete, recurring developer problem.

The roadmap describes direction rather than a fixed delivery commitment.

## v0.1 — Useful CLI foundation

### Product and architecture

- [x] Establish the product direction and non-goals.
- [x] Define architectural boundaries.
- [x] Define command, JSON, error, and exit-code contracts.
- [x] Add project documentation and an MIT License.
- [x] Establish the CLI command structure.
- [x] Keep tool logic independent from CLI and presentation code.

### Commands

- [x] Implement `uuid`.
- [x] Implement `secret`.
- [x] Implement `hash`.
- [x] Implement `jwt inspect`.
- [x] Implement `json pretty` and `json minify`.
- [x] Implement `env diff`.

### Quality and delivery

- [x] Add unit tests for each tool's domain logic.
- [x] Add CLI contract tests for human and JSON output.
- [x] Add tests for failure behavior and sensitive-data handling.
- [x] Add CI for formatting, vetting, and tests.
- [x] Add draft release automation for versioned cross-platform binaries and checksums.
- [x] Verify supported behavior on Windows, Linux, and macOS.
- [x] Document installation and practical examples.
- [x] Publish `v0.1.0`.

### v0.1 acceptance criteria

The milestone is complete when:

- Every listed command works without network access.
- Human output is useful in an interactive terminal.
- `--json` produces the documented versioned envelope.
- Errors use the documented JSON codes and process exit codes.
- Tool logic can be called independently of terminal rendering.
- Secret, JWT, and `.env` commands do not leak unnecessary sensitive data.
- Formatting, vetting, and tests pass in CI.
- README examples match actual behavior.

## Next — Real-world validation

- Use DevKit during real development work.
- Collect recurring tasks worth automating.
- Fix usability problems discovered through actual use.
- Improve cross-platform behavior.
- Stabilize the most useful command and JSON contracts.

Potential candidates, subject to demonstrated need:

- Base64 encode/decode.
- Timestamp conversion.
- Port and process inspection.
- File type inspection.

## Later — Programmatic integration

- Define a programmatic tool adapter after the core CLI is useful.
- Integrate selected tools with `go-ai-playground` if needed.
- Evaluate an MCP adapter only if it provides concrete value.
- Define stable compatibility guarantees before `v1.0.0`.

## Compatibility before v1.0

- JSON envelopes carry an explicit schema version from the beginning.
- Additive JSON fields may be introduced in `0.x` releases.
- Removing a field, renaming it, or changing its meaning must be documented as a breaking change.
- Full public compatibility is not promised before `v1.0.0`.

## Non-goals

- AI or LLM inference inside DevKit.
- Network-dependent core functionality.
- Replacing language-specific compilers or linters.
- Adding commands without a concrete developer use case.
- Premature databases, services, containers, plugin systems, or distributed infrastructure.
