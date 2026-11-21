# Mechanics evaluation reports

Use the shared [skill](../../.agents/skills/mechanics-evaluation-report/SKILL.md) from Cursor, Claude, or Codex.

Athlete data belongs in Google Drive at `Next Pitch LLC/<athlete-name>/evals/<YYYY-MM-DD>/`. The ignored `templates/reports/` directory is a local cache. Keep reusable templates in `tools/mechanics-report/template/`.

```bash
python3 tools/mechanics-report/new_report.py --reports-root /external/evals/2026-11-20 create --athlete "Jane Doe" --stills /external/stills --session 2026-11-20
python3 tools/mechanics-report/new_report.py --reports-root /external/evals/2026-11-20 serve --slug jane-doe
python3 tools/mechanics-report/serve.py /external/evals/2026-11-20/jane-doe
```

The single shared server saves notes into the selected report directory. Upload the changed notes and report archive to Drive after each edit session. Preserve existing notes before a rebuild.
