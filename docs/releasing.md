# Release process

Releases are tag-driven. `.github/workflows/release.yml` runs on `v*.*.*` tag pushes, publishes a GitHub Release from the matching `CHANGELOG.md` section, and updates the floating `v1` tag for `v1.x.x` releases.

## Steps to cut `vX.Y.Z`

1. Update `CHANGELOG.md`:
   - Move entries from `[Unreleased]` into a new section `## [X.Y.Z] - YYYY-MM-DD`.
   - Keep the version in the changelog without the `v` prefix.
   - Update the link references at the bottom of the file:
     - `[Unreleased]` should compare from `vX.Y.Z` to `HEAD`
     - `[X.Y.Z]` should point to the matching release URL
2. Commit the changelog update on the release branch.
3. Tag that commit and push the tag:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

4. Verify the automation:
   - The `Release` workflow completes successfully.
   - A GitHub Release appears at <https://github.com/thanhhaudev/github-stats/releases>.
   - For `v1.x.x`, the floating `v1` tag points at the new release commit.

Never tag without updating `CHANGELOG.md` first. The release workflow intentionally fails when no matching `[X.Y.Z]` section exists.

## Floating tags

- `v1` always points at the latest `v1.x.x` release.
- The `update-v1` job only runs for tags starting with `refs/tags/v1.`, so future major lines such as `v2.x.x` do not overwrite `v1`.
