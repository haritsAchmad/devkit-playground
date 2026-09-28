// Command devkittool adapts a safe subset of devkit commands to the external
// tool contract of local-agent-playground: each tool.json under agent-tools/
// runs this executable with one fixed subcommand, the model's input arrives
// as a JSON object on stdin, the result goes to stdout, and errors go to
// stderr with exit code 1.
//
// It translates that JSON into the same argv and stdin the devkit CLI
// receives and calls cli.Run, so behavior and limits stay identical to
// devkit itself. Only offline commands without side effects whose input is
// inline text or arguments are exposed. Commands that take file or directory
// paths are left out on purpose: devkit would read them outside the
// workspace boundary local-agent enforces for its own file tools. secret and
// jwt inspect (sensitive output) and port inspect (binds a local port) are
// left out too.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/haritsAchmad/devkit-playground/internal/cli"
)

// maxInputBytes bounds the JSON payload read from stdin.
const maxInputBytes = 4 << 20

var version = "dev"

type base64Input struct {
	Text    string `json:"text"`
	Variant string `json:"variant"`
	Padding string `json:"padding"`
}

type base64DecodeInput struct {
	Data    string `json:"data"`
	Variant string `json:"variant"`
	Padding string `json:"padding"`
}

type hashInput struct {
	Text      string `json:"text"`
	Algorithm string `json:"algorithm"`
}

type hashVerifyInput struct {
	Text      string `json:"text"`
	Expected  string `json:"expected"`
	Algorithm string `json:"algorithm"`
}

type jsonInput struct {
	JSON string `json:"json"`
}

type timestampInput struct {
	Value string `json:"value"`
	From  string `json:"from"`
}

type uuidInput struct {
	Count int `json:"count"`
}

// invocation is the devkit argv (after global flags) and stdin for one call.
type invocation struct {
	args  []string
	stdin string
}

// flag appends --name value when value is set.
func flag(args []string, name, value string) []string {
	if value == "" {
		return args
	}
	return append(args, "--"+name, value)
}

// subcommands maps each adapter subcommand to a function that decodes its
// JSON input and builds the devkit invocation.
var subcommands = map[string]func(raw []byte) (invocation, error){
	"base64-encode": func(raw []byte) (invocation, error) {
		var in base64Input
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		args := flag(flag([]string{"base64", "encode"}, "variant", in.Variant), "padding", in.Padding)
		return invocation{args: args, stdin: in.Text}, nil
	},
	"base64-decode": func(raw []byte) (invocation, error) {
		var in base64DecodeInput
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		if strings.TrimSpace(in.Data) == "" {
			return invocation{}, errors.New("field data is required")
		}
		args := flag(flag([]string{"base64", "decode"}, "variant", in.Variant), "padding", in.Padding)
		return invocation{args: args, stdin: in.Data}, nil
	},
	"hash": func(raw []byte) (invocation, error) {
		var in hashInput
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		return invocation{args: flag([]string{"hash"}, "algorithm", in.Algorithm), stdin: in.Text}, nil
	},
	"hash-verify": func(raw []byte) (invocation, error) {
		var in hashVerifyInput
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		if strings.TrimSpace(in.Expected) == "" {
			return invocation{}, errors.New("field expected is required")
		}
		args := flag([]string{"hash", "verify", "--expected", in.Expected}, "algorithm", in.Algorithm)
		return invocation{args: args, stdin: in.Text}, nil
	},
	"json-pretty": jsonCommand("pretty"),
	"json-minify": jsonCommand("minify"),
	"timestamp-convert": func(raw []byte) (invocation, error) {
		var in timestampInput
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		if strings.TrimSpace(in.Value) == "" {
			return invocation{}, errors.New("field value is required")
		}
		// "--" keeps a negative Unix value from being read as a flag.
		return invocation{args: append(flag([]string{"timestamp", "convert"}, "from", in.From), "--", in.Value)}, nil
	},
	"uuid": func(raw []byte) (invocation, error) {
		var in uuidInput
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		args := []string{"uuid"}
		if in.Count != 0 {
			args = append(args, "--count", strconv.Itoa(in.Count))
		}
		return invocation{args: args}, nil
	},
}

func jsonCommand(mode string) func(raw []byte) (invocation, error) {
	return func(raw []byte) (invocation, error) {
		var in jsonInput
		if err := decode(raw, &in); err != nil {
			return invocation{}, err
		}
		if strings.TrimSpace(in.JSON) == "" {
			return invocation{}, errors.New("field json is required")
		}
		return invocation{args: []string{"json", mode}, stdin: in.JSON}, nil
	}
}

// decode parses the stdin payload. An empty payload means no fields.
func decode(raw []byte, v any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("input is not valid JSON: %w", err)
	}
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fail := func(msg string) int {
		fmt.Fprintln(stderr, msg)
		return 1
	}
	if len(args) != 1 {
		return fail("usage: devkittool <subcommand> < input.json")
	}
	build, ok := subcommands[args[0]]
	if !ok {
		return fail("unknown subcommand: " + args[0])
	}
	raw, err := io.ReadAll(io.LimitReader(stdin, maxInputBytes+1))
	if err != nil {
		return fail("read stdin: " + err.Error())
	}
	if len(raw) > maxInputBytes {
		return fail(fmt.Sprintf("input exceeds %d bytes", maxInputBytes))
	}
	inv, err := build(raw)
	if err != nil {
		return fail(err.Error())
	}
	if args[0] == "base64-decode" {
		return runBase64Decode(inv, stdout, stderr)
	}
	var out, errOut bytes.Buffer
	if code := cli.Run(inv.args, strings.NewReader(inv.stdin), &out, &errOut, version); code != 0 {
		return fail(strings.TrimSpace(errOut.String() + out.String()))
	}
	stdout.Write(out.Bytes())
	return 0
}

// runBase64Decode uses devkit's JSON mode so that decoded bytes that are not
// UTF-8 text are reported as binary instead of written raw into the model's
// context.
func runBase64Decode(inv invocation, stdout, stderr io.Writer) int {
	var out, errOut bytes.Buffer
	code := cli.Run(append([]string{"--json"}, inv.args...), strings.NewReader(inv.stdin), &out, &errOut, version)
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
			OutputBytes    int    `json:"output_bytes"`
		} `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		fmt.Fprintln(stderr, strings.TrimSpace("base64 decode: "+errOut.String()+" "+err.Error()))
		return 1
	}
	if code != 0 || !envelope.OK {
		fmt.Fprintln(stderr, envelope.Error.Message)
		return 1
	}
	value := envelope.Data.Value
	if envelope.Data.Representation == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err == nil && utf8.Valid(decoded) {
			value = string(decoded)
		} else {
			fmt.Fprintf(stdout, "decoded %d bytes of binary data (not UTF-8 text); standard Base64 of the bytes:\n%s\n", envelope.Data.OutputBytes, value)
			return 0
		}
	}
	fmt.Fprint(stdout, value)
	if !strings.HasSuffix(value, "\n") {
		fmt.Fprintln(stdout)
	}
	return 0
}
