# Release Process

This document outlines the release process for `kurl`. The repository utilizes an automated CI/CD pipeline powered by GitHub Actions and GoReleaser.

## Automation

Our releases are fully automated via GitHub Actions (`.github/workflows/release.yml`). 
Whenever a new semantic version tag (e.g., `v1.0.1`) is pushed to the repository, the release pipeline is triggered.

The pipeline performs the following tasks:
1. **Compilation**: Compiles highly optimized, stripped binaries for macOS (Intel & Apple Silicon), Linux, and Windows.
2. **Archiving**: Packages the binaries with the `README.md` and `LICENSE`.
3. **GitHub Release**: Creates a draft or published release on the GitHub repository, attaching all compiled artifacts and auto-generating release notes based on the git commit history.
4. **Homebrew Tap Updates**: Automatically commits the new `Formula/kurl.rb` to the `kavix/homebrew-tap` repository so macOS and Linux users can instantly run `brew upgrade kurl`.

## Triggering a Production Release

As a maintainer, follow these steps to release a new version of `kurl`:

1. **Update Changelog**: Ensure the `CHANGELOG.md` reflects the changes in this upcoming release under the correct version header.
2. **Commit Changes**:
   ```bash
   git add CHANGELOG.md
   git commit -m "chore: prepare release v1.0.1"
   git push origin main
   ```
3. **Create a Version Tag**:
   Create an annotated semantic version tag:
   ```bash
   git tag -a v1.0.1 -m "Release v1.0.1"
   ```
4. **Push the Tag**:
   ```bash
   git push origin v1.0.1
   ```

Once pushed, navigate to the **Actions** tab on GitHub to monitor the deployment. When the workflow completes, the release will be live on the [Releases page](https://github.com/kavix/kurl/releases).

## Local Snapshot Compilation

If you need to test the compilation or cross-compilation matrix locally before pushing a tag, you can build a snapshot release:

```bash
make release-local
```
This requires `goreleaser` to be installed on your local machine (`brew install goreleaser`). It will place the compiled archives in the `dist/` directory.
