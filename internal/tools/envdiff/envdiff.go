// Package envdiff compares dotenv key sets without exposing their values.
package envdiff

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

var (
	ErrMalformedLine = errors.New("malformed dotenv line")
	ErrDuplicateKey  = errors.New("duplicate dotenv key")
)

type Result struct {
	Missing []string
	Extra   []string
}

// Compare returns keys missing from target and keys found only in target.
func Compare(reference, target io.Reader) (Result, error) {
	referenceKeys, err := parse(reference)
	if err != nil {
		return Result{}, fmt.Errorf("parse reference environment: %w", err)
	}
	targetKeys, err := parse(target)
	if err != nil {
		return Result{}, fmt.Errorf("parse target environment: %w", err)
	}

	result := Result{
		Missing: difference(referenceKeys, targetKeys),
		Extra:   difference(targetKeys, referenceKeys),
	}
	return result, nil
}

func parse(input io.Reader) (map[string]struct{}, error) {
	keys := make(map[string]struct{})
	scanner := bufio.NewScanner(input)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") || strings.HasPrefix(line, "export\t") {
			line = strings.TrimSpace(line[len("export"):])
		}

		separator := strings.IndexByte(line, '=')
		if separator < 1 {
			return nil, fmt.Errorf("%w at line %d", ErrMalformedLine, lineNumber)
		}
		key := strings.TrimSpace(line[:separator])
		if !validKey(key) {
			return nil, fmt.Errorf("%w at line %d", ErrMalformedLine, lineNumber)
		}
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("%w %q at line %d", ErrDuplicateKey, key, lineNumber)
		}
		keys[key] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read dotenv input: %w", err)
	}

	return keys, nil
}

func validKey(key string) bool {
	if key == "" || !letterOrUnderscore(key[0]) {
		return false
	}
	for index := 1; index < len(key); index++ {
		character := key[index]
		if !letterOrUnderscore(character) && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func letterOrUnderscore(character byte) bool {
	return character == '_' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
}

func difference(left, right map[string]struct{}) []string {
	result := make([]string, 0)
	for key := range left {
		if _, exists := right[key]; !exists {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
