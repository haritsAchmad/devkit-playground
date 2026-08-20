package jwtinspect

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestInspectDecodesHeaderAndClaims(t *testing.T) {
	token := makeToken(`{"alg":"HS256","typ":"JWT"}`, `{"sub":"123","large":9007199254740993}`, "signature")

	result, err := Inspect(token)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Algorithm != "HS256" || result.Header["typ"] != "JWT" || result.Claims["sub"] != "123" {
		t.Errorf("result = %+v, want decoded JWT", result)
	}
	large, ok := result.Claims["large"].(json.Number)
	if !ok || large.String() != "9007199254740993" {
		t.Errorf("large claim = %v, want precision-preserving json.Number", result.Claims["large"])
	}
}

func TestInspectAllowsEmptySignature(t *testing.T) {
	token := makeToken(`{"alg":"none"}`, `{"sub":"123"}`, "")
	result, err := Inspect(token)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Algorithm != "none" {
		t.Errorf("algorithm = %q, want none", result.Algorithm)
	}
}

func TestInspectRejectsMalformedTokensWithoutEchoingThem(t *testing.T) {
	secretMarker := "do-not-echo-this-token"
	cases := []struct {
		name  string
		token string
		want  error
	}{
		{name: "structure", token: secretMarker, want: ErrInvalidStructure},
		{name: "encoding", token: "!!!!.e30.signature", want: ErrInvalidEncoding},
		{name: "json", token: makeToken("[]", `{}`, "signature"), want: ErrInvalidJSON},
		{name: "algorithm", token: makeToken(`{}`, `{}`, "signature"), want: ErrMissingAlgorithm},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Inspect(test.token)
			if !errors.Is(err, test.want) {
				t.Fatalf("Inspect() error = %v, want %v", err, test.want)
			}
			if strings.Contains(err.Error(), secretMarker) {
				t.Error("error exposes raw token")
			}
		})
	}
}

func TestInspectRejectsOversizedToken(t *testing.T) {
	_, err := Inspect(strings.Repeat("x", MaxTokenSize+1))
	if !errors.Is(err, ErrTokenTooLarge) {
		t.Errorf("Inspect() error = %v, want ErrTokenTooLarge", err)
	}
}

func FuzzInspect(f *testing.F) {
	f.Add(makeToken(`{"alg":"HS256"}`, `{"sub":"123"}`, "signature"))
	f.Add("not-a-token")
	f.Fuzz(func(t *testing.T, token string) {
		_, _ = Inspect(token)
	})
}

func makeToken(header, claims, signature string) string {
	encode := base64.RawURLEncoding.EncodeToString
	return encode([]byte(header)) + "." + encode([]byte(claims)) + "." + encode([]byte(signature))
}
