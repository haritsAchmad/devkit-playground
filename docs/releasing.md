# Releasing DevKit

Releases are built by GitHub Actions from an existing Git tag. The workflow creates a draft GitHub Release so its notes and artifacts can be reviewed before publication.

## Release outputs

Each release contains:

- Linux binaries for `amd64` and `arm64` in `.tar.gz` archives.
- Windows binaries for `amd64` and `arm64` in `.zip` archives.
- macOS binaries for Intel (`amd64`) and Apple Silicon (`arm64`) in `.tar.gz` archives.
- `checksums.txt` containing SHA-256 checksums for every archive.

Each archive also includes the README and MIT License.

## Before tagging

1. Confirm the target commit is on `main` and the CI workflow is green.
2. Run the local quality checks:

   ```sh
   go fmt ./...
   go vet ./...
   go test ./...
   git diff --check
   ```

3. Move changelog entries from `Unreleased` into a versioned section with the release date.
4. Confirm README development-version examples and release commands name the target minor version.
5. Commit and merge the release documentation update before creating the tag.

## Creating a release candidate

Use a pre-release tag while validating the process:

```sh
git switch main
git pull --ff-only
git tag -a v0.2.0-rc.1 -m "DevKit v0.2.0-rc.1"
git push origin v0.2.0-rc.1
```

The tag starts the Release workflow. It tests the tagged commit, builds all target archives, creates checksums, and creates a draft GitHub Release.

## Reviewing the draft

Before publishing the draft release:

- Confirm the workflow is green.
- Confirm all six platform archives and `checksums.txt` are attached.
- Review the automatically generated release notes.
- Download and smoke-test at least the Windows `amd64` archive.
- Confirm `devkit --version` prints the tag.
- Mark release-candidate versions as pre-releases in the GitHub UI.

Only publish the draft after these checks pass.

## Stable release

After validating a release candidate, repeat the process with the stable tag:

```sh
git tag -a v0.2.0 -m "DevKit v0.2.0"
git push origin v0.2.0
```

The workflow again creates a draft. Publishing remains a deliberate manual action.

## Verifying a download

PowerShell:

```powershell
Get-FileHash -Algorithm SHA256 .\devkit_0.2.0_windows_amd64.zip
```

Compare the displayed hash with the matching line in `checksums.txt`.

Linux or macOS, after downloading all release files:

```sh
sha256sum --check checksums.txt
```
