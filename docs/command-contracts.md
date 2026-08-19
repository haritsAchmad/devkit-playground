# Command Contracts

This document defines the intended v0.1 CLI behavior. It is the implementation target until real usage justifies an explicit contract change.

## Global invocation

```text
devkit [global flags] <command> [subcommand] [arguments]
```

Initial global flags:

| Flag | Behavior |
|---|---|
| `--json` | Emit the structured JSON representation |
| `--help` | Show help for the current command context |
| `--version` | Print the DevKit version |

Global flags precede the command. Command-specific flags follow the command or subcommand.

## Output streams

| Mode | Success | Expected error |
|---|---|---|
| Human | Result on stdout | Message on stderr |
| JSON | JSON envelope on stdout | JSON error envelope on stdout |

JSON mode must not emit banners, progress messages, or duplicate diagnostics to stderr. A process-level failure that prevents JSON rendering entirely is the only exception.

Successful text output ends with a newline. Commands should behave sensibly in pipelines and must not require an interactive terminal.

## JSON envelope

Success:

```json
{
  "schema_version": "1",
  "command": "uuid",
  "ok": true,
  "data": {}
}
```

Error:

```json
{
  "schema_version": "1",
  "command": "jwt inspect",
  "ok": false,
  "error": {
    "code": "invalid_token",
    "message": "token must contain three segments"
  }
}
```

Rules:

- `schema_version`, `command`, and `ok` are always present.
- Exactly one of `data` or `error` is present.
- `error.code` is stable and intended for programmatic branching.
- `error.message` is concise and safe to display, but consumers should not parse it.
- JSON is written as one complete document followed by a newline.
- Field order is not part of the contract.
- Consumers must tolerate additive fields.

## Exit codes

| Code | Meaning |
|---:|---|
| `0` | Success |
| `2` | Invalid command usage, argument, or flag |
| `3` | Supplied data is malformed or semantically invalid |
| `4` | An operation failed, such as reading a file or obtaining randomness |
| `70` | Unexpected internal failure |

The JSON `error.code` provides finer detail than the process exit code. Human and JSON modes must return the same process exit code for the same failure.

## Input conventions

- A command that accepts a file reads stdin when the file argument is omitted.
- A command must reject ambiguous input, such as both piped data and an explicit incompatible source.
- `-` may be accepted as an explicit stdin path when implementation begins, provided it behaves consistently across commands.
- Inputs are not modified in place in v0.1.

## `uuid`

```text
devkit uuid [--count N]
```

- Generates RFC 4122 UUID version 4 values using `crypto/rand`.
- `--count` defaults to `1`.
- An implementation-defined upper bound prevents accidental excessive output and must be documented in command help.
- Human output contains one UUID per line.
- JSON data contains `uuids`, an array of strings.
- Invalid count is usage error `2`; secure-random failure is operation error `4`.

```json
{
  "uuids": ["550e8400-e29b-41d4-a716-446655440000"]
}
```

## `secret`

```text
devkit secret [--length N] [--encoding hex|base64url]
```

- Generates bytes using `crypto/rand`.
- `--length` is the number of random bytes and defaults to `32`.
- `--encoding` defaults to unpadded `base64url`.
- The maximum length is `4096` bytes to prevent accidental excessive output.
- Human output contains only the encoded secret and a newline.
- Invalid length or encoding is usage error `2`; secure-random failure is operation error `4`.

```json
{
  "secret": "generated-value",
  "encoding": "base64url",
  "bytes": 32
}
```

The secret is intentionally present because it is the requested result. It must not be duplicated in errors, logs, or metadata.

## `hash`

```text
devkit hash [--algorithm sha256|sha512] [file]
```

- `--algorithm` defaults to `sha256`.
- Reads the named file, or stdin when no file is supplied.
- Streams input rather than loading an entire file into memory.
- Human output is the lowercase hexadecimal digest and a newline.
- Unsupported algorithms are usage error `2`; read failures are operation error `4`.

```json
{
  "algorithm": "sha256",
  "digest": "lowercase-hex-digest",
  "bytes": 1234,
  "source": "file"
}
```

`source` is `file` or `stdin`; it does not expose a potentially sensitive path. File contents are never echoed.

## `jwt inspect`

```text
devkit jwt inspect [token]
```

- Reads a positional token when provided; otherwise reads one token from stdin.
- Documentation should recommend stdin because arguments may be visible in shell history or process listings.
- Decodes JWT header and payload using base64url rules.
- Does not verify the signature or establish authenticity.
- Never echoes the raw token or signature.
- Human output contains decoded header and payload plus a prominent `signature not verified` warning.
- Malformed structure, base64url, or JSON is data error `3`.

```json
{
  "header": {},
  "claims": {},
  "algorithm": "HS256",
  "signature_verified": false
}
```

Recognized numeric dates such as `iat`, `nbf`, and `exp` may receive a human-friendly rendering without changing their raw JSON values.

## `json pretty`

```text
devkit json pretty [file]
```

- Reads a file or stdin.
- Accepts exactly one JSON value plus optional surrounding whitespace.
- Preserves JSON number precision rather than implicitly converting numbers through `float64`.
- Does not overwrite the input file.
- Human output is indented JSON.
- JSON-mode data contains `value`, which may hold any JSON type.
- Malformed JSON or trailing content is data error `3`; read failure is operation error `4`.

```json
{
  "value": {}
}
```

## `json minify`

```text
devkit json minify [file]
```

Input and error behavior match `json pretty`. Human output is compact JSON followed by a newline. JSON-mode output places the parsed value in `data.value`; the envelope itself is not a minified-text transport.

## `env diff`

```text
devkit env diff <reference-file> <target-file>
```

- Compares key names, never values.
- Ignores blank lines and comments.
- Reports keys missing from the target and keys present only in the target.
- Sorts keys lexicographically for deterministic output.
- Treats duplicate keys and malformed assignments as data errors.
- Never prints environment values.
- Human output groups `Missing in target` and `Extra in target`.

```json
{
  "missing": ["DATABASE_URL"],
  "extra": ["LOCAL_DEBUG"],
  "counts": {
    "missing": 1,
    "extra": 1
  }
}
```

A successful comparison returns `0` whether differences exist. A future `--check` flag may return non-zero for differences in CI, but it is not part of v0.1. Invalid dotenv data is error `3`; file read failures are error `4`.

## Contract testing

Each command must test:

- Domain behavior independently from CLI rendering.
- Human success and error output.
- JSON success and error envelopes.
- Exit-code mapping.
- Deterministic ordering where applicable.
- Sensitive-data redaction.
- File and stdin input paths where supported.
