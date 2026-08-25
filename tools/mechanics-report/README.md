# Mechanics evaluation reports (local)

Coach-local NPA mechanics HTML reports for THE NEXT PITCH.

This package is **not** wired into the React/Go web app yet. Use the CLI on your machine. Self-service in the site comes later.

## Create a report

```bash
python3 tools/mechanics-report/new_report.py create \
  --athlete "Jane Doe" \
  --stills "/absolute/path/to/stills" \
  --serve
```

## Reopen

```bash
python3 tools/mechanics-report/new_report.py list
python3 tools/mechanics-report/new_report.py serve --slug jane-doe
```

## Layout

- `templates/reports/_master/` — HTML + `serve.py` template
- `templates/reports/<slug>/` — one athlete (stills, `notes.json`, HTML)
- Cursor skill/rule: `.cursor/skills/mechanics-evaluation-report/`, `.cursor/rules/mechanics-evaluation-report.mdc`

**NPA only.** Do not use Tom House language.
