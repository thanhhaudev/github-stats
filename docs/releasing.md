# Release process

Releases are tag-driven. `.github/workflows/release.yml` triggers on any `v*.*.*` tag push, creates a GitHub Release using the matching `CHANGELOG.md` section as the body, and force-updates the `v1` floating tag so users on `@v1` track the latest v1.x.x.

## Steps to cut `vX.Y.Z`

1. Update `CHANGELOG.md`: move items from `[Unreleased]` into a new `[X.Y.Z] - YYYY-MM-DD` section. Add the corresponding link reference at the bottom of the file.
2. Commit the CHANGELOG update on `master` (or merge a PR).
3. Tag the commit:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
4. The workflow runs automatically. Verify:
   - The release appears at <https://github.com/thanhhaudev/github-stats/releases>.
   - `git ls-remote --tags origin v1` points at the new commit (only for `v1.*` tags).

Never tag without first updating `CHANGELOG.md` — the workflow refuses to publish if no matching `[X.Y.Z]` section exists, by design.

## Floating tags

- `v1` always points at the latest `v1.x.x` release. End users pin `@v1` to receive non-breaking updates automatically.
- The `update-v1` job in `release.yml` only runs when the pushed tag starts with `refs/tags/v1.`, so future major releases (`v2.0.0`, etc.) won't accidentally clobber `v1`. When `v2` ships, add a sibling `update-v2` job with the matching prefix.
