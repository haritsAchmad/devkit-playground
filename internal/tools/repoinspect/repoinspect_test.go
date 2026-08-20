package repoinspect

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestInspectReportsDeterministicRepositoryFacts(t *testing.T) {
	root := fstest.MapFS{
		".git/HEAD":                   {Data: []byte("ref: refs/heads/main")},
		"Dockerfile":                  {Data: []byte("ignored")},
		"compose.yaml":                {Data: []byte("ignored")},
		"database/migrations/001.sql": {Data: []byte("ignored")},
		"go.mod":                      {Data: []byte("ignored")},
		"go.sum":                      {Data: []byte("ignored")},
		"package.json":                {Data: []byte("ignored")},
		"pnpm-lock.yaml":              {Data: []byte("ignored")},
		"vitest.config.ts":            {Data: []byte("ignored")},
	}

	result, err := Inspect(root)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if !result.GitRepository {
		t.Error("GitRepository = false, want true")
	}
	wantProjects := []Project{
		{Ecosystem: "go", Manifest: "go.mod", Lockfiles: []string{"go.sum"}},
		{Ecosystem: "javascript", Manifest: "package.json", Lockfiles: []string{"pnpm-lock.yaml"}},
	}
	if !reflect.DeepEqual(result.Projects, wantProjects) {
		t.Errorf("Projects = %+v, want %+v", result.Projects, wantProjects)
	}
	if !reflect.DeepEqual(result.PackageManagers, []string{"go", "pnpm"}) {
		t.Errorf("PackageManagers = %v, want [go pnpm]", result.PackageManagers)
	}
	if !reflect.DeepEqual(result.DockerFiles, []string{"Dockerfile", "compose.yaml"}) {
		t.Errorf("DockerFiles = %v, want sorted Docker files", result.DockerFiles)
	}
	if !reflect.DeepEqual(result.TestConfigs, []string{"vitest.config.ts"}) {
		t.Errorf("TestConfigs = %v, want vitest config", result.TestConfigs)
	}
	if !reflect.DeepEqual(result.MigrationDirs, []string{"database/migrations"}) {
		t.Errorf("MigrationDirs = %v, want database/migrations", result.MigrationDirs)
	}
}

func TestInspectReturnsEmptyArraysForUnrecognizedDirectory(t *testing.T) {
	result, err := Inspect(fstest.MapFS{"README.md": {Data: []byte("ignored")}})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.GitRepository || result.Projects == nil || result.PackageManagers == nil || result.DockerFiles == nil || result.TestConfigs == nil || result.MigrationDirs == nil {
		t.Errorf("result = %+v, want false and non-nil empty slices", result)
	}
}
