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

The order of global flags does not change their meaning, so `--help --json` and `--json --help` both emit JSON help. `--help` and `--version` cannot be combined. A failure before command routing uses `global` as the JSON envelope's `command` value.

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

### `hash verify`

```text
devkit hash verify --expected HEX [--algorithm sha256|sha512] [file]
```

- Reads the named file, or stdin when no file is supplied or the path is `-`.
- Uses SHA-256 by default and supports SHA-512 through the existing `--algorithm` flag.
- Requires an expected hexadecimal digest of exactly the length required by the selected algorithm.
- Accepts uppercase or lowercase hexadecimal input and reports the computed digest in lowercase.
- Streams input and compares decoded digest bytes using a constant-time comparison.
- A match returns success with `verified: true`.
- A mismatch returns data error `3` with stable JSON code `checksum_mismatch`; callers cannot accidentally treat mismatch as successful verification.
- Output and errors do not expose the input file path or file contents.

```json
{
  "algorithm": "sha256",
  "digest": "lowercase-hex-digest",
  "bytes": 1234,
  "source": "file",
  "verified": true
}
```

Missing or malformed expected digests and unsupported algorithms are usage error `2`; checksum mismatch is data error `3`; file and input read failures are operation error `4`.

## `jwt inspect`

```text
devkit jwt inspect [token]
```

- Reads a positional token when provided; otherwise reads one token from stdin.
- Rejects raw token input larger than 1 MiB before trimming surrounding stdin whitespace.
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
- Rejects input larger than 16 MiB before parsing to bound memory and JSON output.
- Accepts exactly one JSON value plus optional surrounding whitespace.
- Preserves JSON number precision rather than implicitly converting numbers through `float64`.
- Does not overwrite the input file.
- Human output is indented JSON.
- JSON-mode data contains `value`, which may hold any JSON type.
- Malformed, trailing, or oversized JSON is data error `3`; read failure is operation error `4`.

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
- Accepts an optional `export` prefix and portable keys matching `[A-Za-z_][A-Za-z0-9_]*`.
- Requires each assignment to be contained on one line in v0.1.
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

## `file inspect`

```text
devkit file inspect <path>
```

- Accepts exactly one regular-file path; directories, symlinks, and special files are rejected.
- Reads the file once to calculate its size and SHA-256 digest.
- Detects MIME using content sniffing over at most the first 512 bytes. This is useful identification, not a complete file-format validator.
- Normalizes the final filename extension to lowercase.
- Reports `extension_check` as `match`, `mismatch`, or `unknown` using a small deterministic mapping of well-known binary formats.
- Uses `unknown` when either the extension is absent or the detected MIME does not have a stable built-in mapping.
- Returns the supplied path because identifying the inspected file is part of the explicit result. Errors do not echo the path.

```json
{
  "path": "document.pdf.exe",
  "name": "document.pdf.exe",
  "extension": ".exe",
  "size_bytes": 183920,
  "detected_mime": "application/pdf",
  "extension_check": "mismatch",
  "sha256": "lowercase-hex-digest"
}
```

Invalid usage is error `2`; a non-regular input is data error `3`; metadata, open, and read failures are operation error `4`.

## `repo inspect`

```text
devkit repo inspect [path]
```

- Inspects the named directory, defaulting to the current directory.
- Rejects files, symlinks, and other non-directory roots.
- Detects a Git marker, root-level project manifests and lockfiles, common Docker files, root-level test configuration, and known migration directories.
- Does not read source files, manifest contents, lockfile contents, configuration contents, environment files, or secrets.
- Reports only known relative metadata paths, plus the explicitly supplied root path.
- Returns arrays in deterministic order and uses empty arrays rather than `null` when no markers are found.
- Initial detection is intentionally root-oriented; recursive workspace and monorepo discovery are outside this contract.

```json
{
  "path": ".",
  "git_repository": true,
  "projects": [
    {
      "ecosystem": "go",
      "manifest": "go.mod",
      "lockfiles": ["go.sum"]
    }
  ],
  "package_managers": ["go"],
  "docker_files": ["Dockerfile"],
  "test_configs": [],
  "migration_dirs": ["migrations"]
}
```

Invalid usage is error `2`; a non-directory root is data error `3`; metadata access failures are operation error `4`.

## `text inspect`

```text
devkit text inspect <path>
```

- Accepts exactly one regular-file path; directories, symlinks, and special files are rejected.
- Streams input instead of loading the complete file into memory.
- Reports total bytes, a recognized Unicode BOM, detected encoding, and UTF-8 validity.
- Detects UTF-8, UTF-8 BOM, UTF-16LE/BE BOM, and UTF-32LE/BE BOM. Invalid UTF-8 without a recognized BOM is `unknown`.
- For valid UTF-8, reports logical line count, final-newline presence, LF/CRLF/CR counts, and a newline style of `lf`, `crlf`, `cr`, `mixed`, or `none`.
- A CRLF sequence counts as one newline and one line terminator.
- An empty file has zero lines; a file ending in a newline does not gain an additional empty line.
- `line_analysis` is `null` for unsupported or invalid encodings rather than presenting byte-level guesses as decoded text facts.
- Returns the supplied path as part of the explicit result. Errors do not echo it.

```json
{
  "path": "README.md",
  "bytes": 4905,
  "encoding": "utf-8",
  "bom": "none",
  "utf8_valid": true,
  "line_analysis": {
    "style": "lf",
    "line_count": 122,
    "lf": 122,
    "crlf": 0,
    "cr": 0,
    "final_newline": true
  }
}
```

Invalid usage is error `2`; a non-regular input is data error `3`; metadata, open, and read failures are operation error `4`.

## `port inspect`

```text
devkit port inspect [--host IP] <port>
```

- Accepts a TCP port from `1` through `65535` and a local IPv4 or IPv6 address.
- Defaults to `127.0.0.1`; hostnames are rejected to avoid DNS or resolver-dependent behavior.
- Briefly binds the requested local address and immediately closes it. The command does not connect to the service occupying a port.
- Reports `available` when the bind succeeds, `in_use` when the operating system reports that the address is already bound, and `unavailable` when the operating system denies the bind.
- Reports a point-in-time observation; another process may claim or release the port immediately afterward.
- Does not claim that `in_use` means a reachable or healthy listening service.
- The portable implementation does not resolve PID or process name. Those fields are `null` and `owner_inspection` is `not_supported`.
- `reason` is `bind_access_denied` for the recognized unavailable case and `null` otherwise. Possible causes include permissions, reserved ranges, or an exclusive wildcard binding; the portable probe does not guess which one applies.

```json
{
  "host": "127.0.0.1",
  "port": 5432,
  "protocol": "tcp",
  "state": "in_use",
  "reason": null,
  "pid": null,
  "process": null,
  "owner_inspection": "not_supported"
}
```

Invalid host, port, or usage is error `2`; an unclassified operating-system probe failure is operation error `4`.

## `timestamp convert`

```text
devkit timestamp convert [--from unix|unix-ms|rfc3339] <value>
```

- Converts exactly one timestamp value; `--from` defaults to `unix` seconds.
- Requires an explicit supported format rather than guessing whether a numeric value uses seconds or milliseconds.
- Accepts signed base-10 integers for `unix` and `unix-ms` and RFC3339 with optional fractional seconds for `rfc3339`.
- Normalizes calendar output to UTC using RFC3339Nano formatting, independent of the machine timezone.
- Reports Unix seconds, Unix milliseconds, and the nanosecond fraction within the second.
- Preserves RFC3339 nanosecond precision through `unix_seconds` plus `subsecond_nanoseconds`; `unix_milliseconds` is the corresponding millisecond representation.
- Rejects values outside years `0000` through `9999`, the range representable by the documented RFC3339 output.

```json
{
  "input_format": "rfc3339",
  "utc": "2026-08-20T03:30:15.123456789Z",
  "unix_seconds": 1787196615,
  "unix_milliseconds": 1787196615123,
  "subsecond_nanoseconds": 123456789
}
```

Unsupported formats and invalid command usage are error `2`; malformed or out-of-range timestamp values are data error `3`.

## `base64 encode` and `base64 decode`

```text
devkit base64 encode [--variant standard|url] [--padding padded|raw] [file]
devkit base64 decode [--variant standard|url] [--padding padded|raw] [file]
```

- Reads the named file, or stdin when no file is supplied or the path is `-`.
- Supports the standard and URL-safe alphabets with padded or raw encoding; defaults are `standard` and `padded`.
- Applies the selected alphabet and padding strictly instead of silently accepting a different variant.
- Limits input to 16 MiB to bound memory and JSON output. Oversized input is rejected before transformation.
- Encode preserves every input byte. Decode ignores surrounding whitespace and the CR/LF behavior accepted by Go's standard Base64 decoder.
- Human encode output ends with a newline. Human decode output writes the exact decoded bytes and does not append a newline.
- JSON encode output uses `representation: "utf-8"` because Base64 text is ASCII.
- JSON decode output uses `representation: "utf-8"` when decoded bytes are valid UTF-8. For arbitrary binary, `value` carries canonical padded standard Base64 and `representation` is `base64`; `output_bytes` still describes the decoded bytes.
- `source` is `file` or `stdin` and does not expose a potentially sensitive path. Invalid-data errors never echo input.

```json
{
  "operation": "decode",
  "variant": "standard",
  "padding": "padded",
  "input_bytes": 8,
  "output_bytes": 5,
  "value": "hello",
  "representation": "utf-8",
  "source": "stdin"
}
```

Invalid flags and modes are error `2`; invalid or oversized Base64 is data error `3`; file and input read failures are operation error `4`.

## Contract testing

Each command must test:

- Domain behavior independently from CLI rendering.
- Human success and error output.
- JSON success and error envelopes.
- Exit-code mapping.
- Deterministic ordering where applicable.
- Sensitive-data redaction.
- File and stdin input paths where supported.
