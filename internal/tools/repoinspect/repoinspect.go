// Package repoinspect detects common repository metadata without reading file contents.
package repoinspect

import (
	"errors"
	"io/fs"
	"sort"
)

type Project struct {
	Ecosystem string
	Manifest  string
	Lockfiles []string
}

type Result struct {
	GitRepository   bool
	Projects        []Project
	PackageManagers []string
	DockerFiles     []string
	TestConfigs     []string
	MigrationDirs   []string
}

type projectDefinition struct {
	ecosystem string
	manifest  string
	locks     map[string]string
}

var projectDefinitions = []projectDefinition{
	{ecosystem: "go", manifest: "go.mod", locks: map[string]string{"go.sum": "go"}},
	{ecosystem: "javascript", manifest: "package.json", locks: map[string]string{
		"bun.lock": "bun", "bun.lockb": "bun", "package-lock.json": "npm", "pnpm-lock.yaml": "pnpm", "yarn.lock": "yarn",
	}},
	{ecosystem: "php", manifest: "composer.json", locks: map[string]string{"composer.lock": "composer"}},
	{ecosystem: "python", manifest: "pyproject.toml", locks: map[string]string{"poetry.lock": "poetry", "uv.lock": "uv"}},
	{ecosystem: "rust", manifest: "Cargo.toml", locks: map[string]string{"Cargo.lock": "cargo"}},
	{ecosystem: "java-maven", manifest: "pom.xml"},
	{ecosystem: "java-gradle", manifest: "build.gradle"},
	{ecosystem: "java-gradle", manifest: "build.gradle.kts"},
}

var dockerCandidates = []string{"Dockerfile", "compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}
var migrationCandidates = []string{"database/migrations", "db/migrations", "migrations"}
var testConfigPatterns = []string{
	"jest.config.*", "phpunit.xml", "phpunit.xml.dist", "playwright.config.*", "pytest.ini", "vitest.config.*",
}

// Inspect reports known repository markers from the supplied filesystem root.
func Inspect(root fs.FS) (Result, error) {
	result := Result{
		Projects:      []Project{},
		TestConfigs:   []string{},
		MigrationDirs: []string{},
	}
	gitType, err := entryType(root, ".git")
	if err != nil {
		return Result{}, err
	}
	result.GitRepository = gitType != entryMissing

	managers := make(map[string]struct{})
	for _, definition := range projectDefinitions {
		kind, err := entryType(root, definition.manifest)
		if err != nil {
			return Result{}, err
		}
		if kind != entryFile {
			continue
		}
		project := Project{Ecosystem: definition.ecosystem, Manifest: definition.manifest, Lockfiles: []string{}}
		for lockfile, manager := range definition.locks {
			kind, err := entryType(root, lockfile)
			if err != nil {
				return Result{}, err
			}
			if kind == entryFile {
				project.Lockfiles = append(project.Lockfiles, lockfile)
				managers[manager] = struct{}{}
			}
		}
		sort.Strings(project.Lockfiles)
		result.Projects = append(result.Projects, project)
	}

	result.PackageManagers = sortedKeys(managers)
	result.DockerFiles, err = existingFiles(root, dockerCandidates)
	if err != nil {
		return Result{}, err
	}
	for _, pattern := range testConfigPatterns {
		matches, err := fs.Glob(root, pattern)
		if err != nil {
			return Result{}, err
		}
		for _, match := range matches {
			kind, err := entryType(root, match)
			if err != nil {
				return Result{}, err
			}
			if kind == entryFile {
				result.TestConfigs = append(result.TestConfigs, match)
			}
		}
	}
	result.TestConfigs = uniqueSorted(result.TestConfigs)
	for _, candidate := range migrationCandidates {
		kind, err := entryType(root, candidate)
		if err != nil {
			return Result{}, err
		}
		if kind == entryDirectory {
			result.MigrationDirs = append(result.MigrationDirs, candidate)
		}
	}
	return result, nil
}

type entryKind uint8

const (
	entryMissing entryKind = iota
	entryFile
	entryDirectory
	entryOther
)

func entryType(root fs.FS, name string) (entryKind, error) {
	info, err := fs.Stat(root, name)
	if err != nil {
		if isMissing(err) {
			return entryMissing, nil
		}
		return entryMissing, err
	}
	if info.Mode().IsRegular() {
		return entryFile, nil
	}
	if info.IsDir() {
		return entryDirectory, nil
	}
	return entryOther, nil
}

func isMissing(err error) bool { return errors.Is(err, fs.ErrNotExist) }

func existingFiles(root fs.FS, candidates []string) ([]string, error) {
	result := make([]string, 0)
	for _, candidate := range candidates {
		kind, err := entryType(root, candidate)
		if err != nil {
			return nil, err
		}
		if kind == entryFile {
			result = append(result, candidate)
		}
	}
	sort.Strings(result)
	return result, nil
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func uniqueSorted(values []string) []string {
	sort.Strings(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
