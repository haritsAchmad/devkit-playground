# Architecture

This document defines DevKit's initial architectural boundaries. It intentionally favors a small, direct design that can evolve from real use.

## Context

DevKit provides deterministic developer utilities through a CLI. The same tool logic may later be called through a programmatic adapter or an MCP server, so terminal concerns must not become part of the domain implementation.

```text
Human ───────────┐
                 ├─> CLI orchestration ─> Tool logic
Program/agent ───┘          │
                            ├─> Human renderer
                            └─> JSON renderer
```

## Intended repository structure

```text
cmd/
  devkit/
    main.go
internal/
  cli/
    commands/
  output/
  tools/
    uuid/
    secret/
    hash/
    jwt/
    jsonutil/
    envdiff/
testdata/
docs/
```

This is a target structure, not a requirement to create empty directories in advance.

## Dependency rules

Dependencies flow inward:

```text
cmd/devkit -> internal/cli -> internal/tools
                         \-> internal/output
```

- `cmd/devkit` wires dependencies and maps the final result to a process exit code.
- `internal/cli` parses arguments, invokes tool logic, and selects a renderer.
- `internal/tools` implements operations and returns typed results and errors.
- `internal/output` renders results for humans or as structured JSON.
- Tool packages must not print, terminate the process, parse global CLI flags, or depend on a renderer.
- Renderers must not contain tool behavior.

Packages begin under `internal/`. A package should become public only when a concrete external Go consumer requires a supported API.

## Tool design

Each tool should have:

- A narrow input type.
- A typed result independent of presentation.
- Validation close to the domain logic.
- Errors that can be mapped to stable CLI error codes.
- Unit tests that do not invoke a subprocess.

CLI handlers should remain thin. They translate command-line inputs into domain inputs, call the tool, and pass its result to a renderer.

## Output boundary

Human output prioritizes readability and shell use. JSON output prioritizes stable field names and unambiguous types. They are separate representations of the same typed result; one should not be produced by parsing the other.

The JSON envelope and exit-code mapping are defined in [Command contracts](command-contracts.md).

## Offline-first behavior

Core commands must not require network access, remote APIs, telemetry, a database, Docker, Node.js, or an LLM. A future network-capable feature must be explicitly separated from the core and must not change offline command behavior.

## Security principles

- Use `crypto/rand` for UUIDs and secret generation.
- Never log generated secrets, raw JWTs, signatures, environment values, or file contents.
- Prefer standard input over command arguments for sensitive values because arguments may appear in shell history or process listings.
- JWT inspection decodes data but does not verify authenticity. Output must state this clearly.
- Environment comparison operates on key names only.
- Errors should identify what failed without echoing sensitive input.
- JSON output contains sensitive output only when that value is the explicit result requested by the caller.

## Dependency policy

Prefer the Go standard library where practical. A third-party dependency is acceptable only when it provides concrete value that outweighs its maintenance, security, binary-size, and API-stability costs.

Avoid adding frameworks, dependency injection containers, plugin systems, or generic abstraction layers before repeated use demonstrates a need.

## Compatibility

Before `v1.0.0`, command and JSON contracts may evolve, but changes must be intentional and documented. JSON includes a `schema_version` from the first release so machine consumers can reject unsupported representations.

Additive JSON fields may be introduced in `0.x`. Removing, renaming, or changing the meaning or type of a documented field is a breaking change and must be called out in the changelog.

## Deferred decisions

The following are intentionally deferred:

- A public Go package API.
- A general programmatic adapter interface.
- An MCP adapter.
- A plugin system.
- Persistent configuration or state.

These decisions should be driven by working commands and real consumers.
