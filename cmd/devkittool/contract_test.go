package main

// Contract tests: every agent-tools/*/tool.json is checked against
// devkittool, then every operation runs through the real binary exactly as
// local-agent-playground calls it -- args from the manifest, JSON payload on
// stdin, result on stdout, errors on stderr with exit code 1.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

type manifest struct {
	SchemaVersion    int    `json:"schema_version"`
	Name             string `json:"name"`
	Executable       string `json:"executable"`
	Args             []string
	WorkingDirectory string `json:"working_directory"`
	Network          bool   `json:"network"`
	ActionKind       string `json:"action_kind"`
	DryRunSupported  bool   `json:"dry_run_supported"`
	Timeout          string `json:"timeout"`
	MaxOutputBytes   int    `json:"max_output_bytes"`
	InputSchema      struct {
		Type                 string                     `json:"type"`
		Properties           map[string]json.RawMessage `json:"properties"`
		Required             []string                   `json:"required"`
		AdditionalProperties *bool                      `json:"additionalProperties"`
	} `json:"input_schema"`
	dir string
}

// inputTypes maps each subcommand to the struct decoded from stdin. It must
// match the subcommands map.
var inputTypes = map[string]reflect.Type{
	"base64-encode":     reflect.TypeFor[base64Input](),
	"base64-decode":     reflect.TypeFor[base64DecodeInput](),
	"hash":              reflect.TypeFor[hashInput](),
	"hash-verify":       reflect.TypeFor[hashVerifyInput](),
	"json-pretty":       reflect.TypeFor[jsonInput](),
	"json-minify":       reflect.TypeFor[jsonInput](),
	"timestamp-convert": reflect.TypeFor[timestampInput](),
	"uuid":              reflect.TypeFor[uuidInput](),
}

func loadManifests(t *testing.T) []manifest {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "..", "agent-tools", "*", "tool.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no manifests found: %v", err)
	}
	var out []manifest
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		m.dir = filepath.Base(filepath.Dir(path))
		out = append(out, m)
	}
	return out
}

func jsonTags(typ reflect.Type) []string {
	var tags []string
	for field := range typ.Fields() {
		if tag := strings.Split(field.Tag.Get("json"), ",")[0]; tag != "" && tag != "-" {
			tags = append(tags, tag)
		}
	}
	slices.Sort(tags)
	return tags
}

// Every manifest is consistent with devkittool: name matches its folder, the
// subcommand exists, input_schema properties equal the decoded JSON fields
// (json.Unmarshal silently drops unknown fields, so a misspelled property
// would never be noticed in use), and every tool is offline, read-only, and
// has no dry-run.
func TestManifestsMatchDevkittool(t *testing.T) {
	used := map[string]bool{}
	for _, m := range loadManifests(t) {
		t.Run(m.dir, func(t *testing.T) {
			if len(m.Args) != 1 {
				t.Fatalf("args = %v, want exactly one subcommand", m.Args)
			}
			op := m.Args[0]
			typ, known := inputTypes[op]
			if _, handled := subcommands[op]; !known || !handled {
				t.Fatalf("subcommand %q unknown to devkittool", op)
			}
			used[op] = true
			if m.dir != "devkit-"+op || m.Name != "devkit_"+strings.ReplaceAll(op, "-", "_") || m.SchemaVersion != 1 {
				t.Errorf("folder %q name %q schema_version %d, want devkit-%s / devkit_%s / 1", m.dir, m.Name, m.SchemaVersion, op, strings.ReplaceAll(op, "-", "_"))
			}
			if m.Executable != "./devkittool.exe" || m.WorkingDirectory != "." || m.Network || m.ActionKind != "read" || m.DryRunSupported {
				t.Errorf("executable=%q working_directory=%q network=%v action_kind=%q dry_run=%v, want ./devkittool.exe . false read false", m.Executable, m.WorkingDirectory, m.Network, m.ActionKind, m.DryRunSupported)
			}
			if d, err := time.ParseDuration(m.Timeout); err != nil || d <= 0 {
				t.Errorf("timeout %q must be a positive duration", m.Timeout)
			}
			if m.MaxOutputBytes <= 0 {
				t.Errorf("max_output_bytes = %d", m.MaxOutputBytes)
			}
			schema := m.InputSchema
			if schema.Type != "object" || schema.AdditionalProperties == nil || *schema.AdditionalProperties {
				t.Errorf("input_schema must be an object with additionalProperties false")
			}
			var props []string
			for name := range schema.Properties {
				props = append(props, name)
			}
			slices.Sort(props)
			if want := jsonTags(typ); !slices.Equal(props, want) {
				t.Errorf("input_schema properties %v, devkittool input fields %v", props, want)
			}
			for _, name := range schema.Required {
				if _, ok := schema.Properties[name]; !ok {
					t.Errorf("required %q is not a property", name)
				}
			}
		})
	}
	for op := range subcommands {
		if !used[op] {
			t.Errorf("subcommand %q is not used by any manifest", op)
		}
	}
}

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

func devkittoolBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "devkittool-contract")
		if err != nil {
			buildErr = err
			return
		}
		name := "devkittool"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		binPath = filepath.Join(dir, name)
		if out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput(); err != nil {
			buildErr = errors.New(string(out))
		}
	})
	if buildErr != nil {
		t.Fatalf("build devkittool: %v", buildErr)
	}
	return binPath
}

func runDevkittool(t *testing.T, bin, op, payload string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, op)
	cmd.Stdin = strings.NewReader(payload)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return out.String(), errOut.String(), code
}

// Every operation through its manifest args.
func TestEndToEndThroughManifests(t *testing.T) {
	bin := devkittoolBinary(t)
	args := map[string]string{}
	for _, m := range loadManifests(t) {
		args[m.Name] = m.Args[0]
	}
	uuidLine := regexp.MustCompile(`(?m)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for i, s := range []struct {
		tool, payload string
		wantCode      int
		want          string // exact stdout on success, substring of stderr on error
	}{
		{"devkit_base64_encode", `{"text":"halo dunia"}`, 0, "aGFsbyBkdW5pYQ==\n"},
		{"devkit_base64_encode", `{"text":"??>","variant":"url","padding":"raw"}`, 0, "Pz8-\n"},
		{"devkit_base64_decode", `{"data":"aGFsbyBkdW5pYQ=="}`, 0, "halo dunia\n"},
		// Binary result is described, never written raw.
		{"devkit_base64_decode", `{"data":"//4AYWJj"}`, 0, "decoded 6 bytes of binary data (not UTF-8 text); standard Base64 of the bytes:\n//4AYWJj\n"},
		{"devkit_base64_decode", `{"data":"bukan base64!"}`, 1, "not valid for the selected Base64 variant"},
		{"devkit_base64_decode", `{}`, 1, "data is required"},
		{"devkit_hash", `{"text":"abc"}`, 0, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad\n"},
		// Exact bytes: a trailing newline changes the digest.
		{"devkit_hash", `{"text":"abc\n"}`, 0, "edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb\n"},
		{"devkit_hash_verify", `{"text":"abc","expected":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}`, 0, "Verified sha256: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad\n"},
		{"devkit_hash_verify", `{"text":"abd","expected":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}`, 1, "devkit:"},
		{"devkit_hash_verify", `{"text":"abc"}`, 1, "expected is required"},
		{"devkit_json_pretty", `{"json":"{\"a\":[1,2]}"}`, 0, "{\n  \"a\": [\n    1,\n    2\n  ]\n}\n"},
		{"devkit_json_minify", `{"json":"{ \"a\" : 1 }"}`, 0, "{\"a\":1}\n"},
		{"devkit_json_pretty", `{"json":"{bad"}`, 1, "exactly one valid JSON value"},
		{"devkit_timestamp_convert", `{"value":"1700000000"}`, 0, "UTC: 2023-11-14T22:13:20Z\nUnix seconds: 1700000000\nUnix milliseconds: 1700000000000\nSubsecond nanoseconds: 0\n"},
		// A negative Unix value is a value, not a flag.
		{"devkit_timestamp_convert", `{"value":"-1"}`, 0, "UTC: 1969-12-31T23:59:59Z\nUnix seconds: -1\nUnix milliseconds: -1000\nSubsecond nanoseconds: 0\n"},
		{"devkit_timestamp_convert", `{"value":"2026-09-28T10:00:00+07:00","from":"rfc3339"}`, 0, "UTC: 2026-09-28T03:00:00Z\nUnix seconds: 1790564400\nUnix milliseconds: 1790564400000\nSubsecond nanoseconds: 0\n"},
		{"devkit_uuid", `{"count":5000}`, 1, "between 1 and 1000"},
		{"devkit_uuid", `bukan json`, 1, "input is not valid JSON"},
	} {
		op, ok := args[s.tool]
		if !ok {
			t.Fatalf("step %d: manifest %s missing", i, s.tool)
		}
		stdout, stderr, code := runDevkittool(t, bin, op, s.payload)
		if code != s.wantCode {
			t.Fatalf("step %d %s: exit %d, want %d\nstdout: %s\nstderr: %s", i, s.tool, code, s.wantCode, stdout, stderr)
		}
		// Output contract: success -> stdout only; error -> stderr only.
		if code == 0 && (stderr != "" || stdout != s.want) {
			t.Errorf("step %d %s: stdout %q stderr %q, want stdout %q", i, s.tool, stdout, stderr, s.want)
		}
		if code != 0 && (stdout != "" || !strings.Contains(stderr, s.want)) {
			t.Errorf("step %d %s: stdout %q stderr %q, want stderr containing %q", i, s.tool, stdout, stderr, s.want)
		}
	}

	stdout, _, code := runDevkittool(t, bin, args["devkit_uuid"], `{"count":3}`)
	if code != 0 || len(uuidLine.FindAllString(stdout, -1)) != 3 {
		t.Errorf("uuid count 3: exit %d stdout %q", code, stdout)
	}
	if _, stderr, code := runDevkittool(t, bin, "secret", `{}`); code != 1 || !strings.Contains(stderr, "unknown subcommand") {
		t.Errorf("subcommand outside the subset: exit %d stderr %q", code, stderr)
	}
}
