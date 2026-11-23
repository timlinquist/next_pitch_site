# Repository instructions

This is Tim Linquist's personal Next Pitch repository. These instructions apply only to this repository.

## Work tracking and commits

- GUS work items and `@W` commit or PR title prefixes are not required here.
- Create an isolated session worktree before changing repository files.
- Preserve unrelated work in other branches and worktrees.
- Sign every commit with the personal GPG key `180DA58B906CE558`.
- Use the committer email `tim.linquist@gmail.com`.
- Use local Git to create signed commits. Do not create unsigned commits through an API.
- Verify that GitHub reports `commit.verification.verified: true` after each push.
- Scan the changes for credentials before every push. Resolve findings before pushing.
- Remove completed temporary worktrees after verification.

These personal signing and GUS instructions replace the Salesforce defaults for this repository.

## Athlete reports

- Follow `.agents/skills/mechanics-evaluation-report/SKILL.md` for athlete reports.
- Use the shared report tools under `tools/mechanics-report/`.
- Store source stills under the athlete's Google Drive folder in `Next Pitch LLC`.
- Store final reports under `Next Pitch LLC/<athlete-name>/evals/<YYYY-MM-DD>/`.
- Keep athlete reports, stills, notes, metadata, and archives out of Git.
- Treat the ignored `templates/reports/` directory as a local cache only.
- Deliver the final HTML, PDF, notes, metadata, and editable archive to Drive.
- Preserve the user's notes and source images.
- Keep the report editor and server open while the user adds notes or reviews the report.

## Shared instructions

Maintain repository preferences in this file. Claude and Cursor entry files reference this file instead of duplicating its requirements.
