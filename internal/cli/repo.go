package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/repoinspect"
)

const repoInspectUsage = `Usage:
  devkit [--json] repo inspect [path]

Arguments:
  path                  repository directory to inspect (default .)

Inspection checks known metadata paths without reading source files, manifests,
lockfiles, configuration contents, environment files, or secrets.
`

type repoProjectData struct {
	Ecosystem string   `json:"ecosystem"`
	Manifest  string   `json:"manifest"`
	Lockfiles []string `json:"lockfiles"`
}

type repoInspectData struct {
	Path            string            `json:"path"`
	GitRepository   bool              `json:"git_repository"`
	Projects        []repoProjectData `json:"projects"`
	PackageManagers []string          `json:"package_managers"`
	DockerFiles     []string          `json:"docker_files"`
	TestConfigs     []string          `json:"test_configs"`
	MigrationDirs   []string          `json:"migration_dirs"`
}

func runRepo(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "repo", "invalid_usage", "repo requires the inspect subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeRepoInspectHelp(stdout, jsonMode)
	}
	if args[0] != "inspect" {
		return writeFailure(stdout, stderr, jsonMode, "repo", "unknown_subcommand", "unknown repo subcommand", ExitUsage)
	}
	return runRepoInspect(args[1:], stdout, stderr, jsonMode)
}

func runRepoInspect(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("repo inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeRepoInspectHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "repo inspect", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() > 1 {
		return writeFailure(stdout, stderr, jsonMode, "repo inspect", "invalid_usage", "repo inspect accepts at most one path", ExitUsage)
	}
	rootPath := "."
	if flags.NArg() == 1 {
		rootPath = flags.Arg(0)
	}
	info, err := os.Lstat(rootPath)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "repo inspect", "directory_read_failed", "could not inspect repository directory", ExitOperation)
	}
	if !info.IsDir() {
		return writeFailure(stdout, stderr, jsonMode, "repo inspect", "not_directory", "repository path is not a directory", ExitData)
	}

	result, err := repoinspect.Inspect(os.DirFS(rootPath))
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "repo inspect", "directory_read_failed", "could not read repository metadata", ExitOperation)
	}
	data := repoInspectData{
		Path: rootPath, GitRepository: result.GitRepository,
		Projects:        make([]repoProjectData, 0, len(result.Projects)),
		PackageManagers: result.PackageManagers, DockerFiles: result.DockerFiles,
		TestConfigs: result.TestConfigs, MigrationDirs: result.MigrationDirs,
	}
	for _, project := range result.Projects {
		data.Projects = append(data.Projects, repoProjectData{
			Ecosystem: project.Ecosystem, Manifest: project.Manifest, Lockfiles: project.Lockfiles,
		})
	}

	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "repo inspect", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	return writeRepoInspectHuman(stdout, data)
}

func writeRepoInspectHuman(stdout io.Writer, data repoInspectData) int {
	if _, err := fmt.Fprintf(stdout, "Path: %s\nGit repository: %t\n", data.Path, data.GitRepository); err != nil {
		return ExitInternal
	}
	if err := writeRepoList(stdout, "Projects", projectSummaries(data.Projects)); err != nil {
		return ExitInternal
	}
	for _, section := range []struct {
		label  string
		values []string
	}{
		{"Package managers", data.PackageManagers},
		{"Docker files", data.DockerFiles},
		{"Test configs", data.TestConfigs},
		{"Migration directories", data.MigrationDirs},
	} {
		if err := writeRepoList(stdout, section.label, section.values); err != nil {
			return ExitInternal
		}
	}
	return ExitSuccess
}

func projectSummaries(projects []repoProjectData) []string {
	values := make([]string, 0, len(projects))
	for _, project := range projects {
		value := project.Ecosystem + " (" + project.Manifest + ")"
		if len(project.Lockfiles) > 0 {
			value += "; lockfiles: " + strings.Join(project.Lockfiles, ", ")
		}
		values = append(values, value)
	}
	return values
}

func writeRepoList(stdout io.Writer, label string, values []string) error {
	if _, err := fmt.Fprintf(stdout, "%s:\n", label); err != nil {
		return err
	}
	if len(values) == 0 {
		_, err := io.WriteString(stdout, "  (none)\n")
		return err
	}
	for _, value := range values {
		if _, err := fmt.Fprintf(stdout, "  %s\n", value); err != nil {
			return err
		}
	}
	return nil
}

func writeRepoInspectHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "repo inspect help", map[string]string{"usage": repoInspectUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, repoInspectUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
