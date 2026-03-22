---
name: shrinkray-release
description: Cut a new semver release — updates CHANGELOG.md, commits, tags, and pushes to trigger the GoReleaser CI pipeline. Use this skill whenever the user says "release", "cut a release", "bump version", "tag a release", "ship it", or wants to publish a new version. Also use when the user types "/release".
---

# Release Workflow

This skill walks through cutting a new semver release for shrinkray. The release pipeline is fully automated after pushing a tag — GitHub Actions runs tests then GoReleaser builds binaries for all platforms.

## Prerequisites

Before starting, verify:
- You are on the `main` branch
- Working tree is clean (no uncommitted changes) — if not, ask the user whether to commit or stash first
- `CHANGELOG.md` exists and has an `[Unreleased]` section with content

If any prerequisite fails, stop and help the user resolve it before continuing.

## Step 1: Determine the Next Version

Run `git tag --sort=-v:refname | head -1` to find the latest tag.

Read the `[Unreleased]` section of `CHANGELOG.md` and categorize the changes:

| Change Type | Version Bump |
|-------------|-------------|
| `### Removed` or `### Changed` with breaking API/behavior changes | **Major** (X+1.0.0) |
| `### Added` (new features or capabilities) | **Minor** (X.Y+1.0) |
| `### Fixed`, `### Security`, or minor `### Changed` only | **Patch** (X.Y.Z+1) |

Present the proposed version to the user with your reasoning. For example:

> The `[Unreleased]` section has 5 additions and 6 fixes. I'd recommend **v0.2.0** (minor bump for new features). Want to go with this, or use a different version?

Wait for confirmation. The user may override (e.g., they might want a major bump even if there are only features).

## Step 2: Update CHANGELOG.md

Once the version is confirmed (e.g., `0.2.0`):

1. **Rename `[Unreleased]`** to `[0.2.0] - YYYY-MM-DD` using today's date
2. **Add a fresh `[Unreleased]` section** above the new release — just the header, no subsection placeholders
3. **Update comparison links** at the bottom of the file:
   - Change the `[Unreleased]` link to compare from the new tag: `[Unreleased]: https://github.com/jparkerweb/shrinkray/compare/v0.2.0...HEAD`
   - Add the new version link: `[0.2.0]: https://github.com/jparkerweb/shrinkray/compare/v0.1.0...v0.2.0`

Show the user a summary of what changed in CHANGELOG.md before proceeding.

## Step 3: Commit and Tag

Stage and commit:

```bash
git add CHANGELOG.md
git commit -m "Release v0.2.0"
```

Create an annotated tag:

```bash
git tag -a v0.2.0 -m "v0.2.0"
```

Tell the user the commit and tag have been created locally.

## Step 4: Push (with confirmation)

**Always ask before pushing.** Say something like:

> Ready to push the release commit and tag to origin. This will trigger the GitHub Actions release pipeline which builds binaries for all platforms. Push now?

Only after the user confirms:

```bash
git push origin main --tags
```

## Step 5: Post-Push Summary

After pushing, show the user:

- **CI Monitor**: `https://github.com/jparkerweb/shrinkray/actions`
- **Release Page**: `https://github.com/jparkerweb/shrinkray/releases/tag/v0.2.0`

And a brief note:

> Release v0.2.0 pushed. GitHub Actions will run tests and then GoReleaser builds binaries for Windows, macOS, and Linux (amd64 + arm64). Monitor progress at the Actions link above. Binaries will appear on the Releases page once complete.

## Reference

- GoReleaser config: `.goreleaser.yaml`
- CI workflow: `.github/workflows/release.yml` (triggers on `v*` tags)
- Changelog format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
- Install script (`scripts/install-local.ps1`) derives version from `git describe --tags`
