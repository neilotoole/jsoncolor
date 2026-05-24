# CLAUDE.md

Guidance for Claude Code (and other AI agents) working in this repository.

## Releases

- **Never create a GitHub Release.** Do not run `gh release create`, and do not
  use any release-creation API or MCP tool — not proactively, not as a
  follow-up suggestion.
- Tagging is the release mechanism: an annotated `vX.Y.Z` tag pushed to
  `origin`. The `## CHANGELOG` section in `README.md` is the human-readable
  changelog; add an entry there for each release.
