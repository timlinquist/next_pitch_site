---
name: mechanics-evaluation-report
description: >-
  Create and edit THE NEXT PITCH NPA mechanics evaluation HTML reports for any
  athlete from delivery stills, persist coach notes locally, and serve for
  Print/Save as PDF. Use when the user asks for a mechanics evaluation, athlete
  report, stills-based pitching report, NPA checkpoint walkthrough, or to open
  notes for an existing athlete.
---

# Mechanics evaluation report

Lives in this repo (`next_pitch_site`) as a **local CLI** under `tools/mechanics-report/`. Not wired into the React/Go app yet. Self-service comes later.

Repeatable local report for THE NEXT PITCH. One folder per athlete. Coach types notes in the browser. Download PDF via print.

**NPA only.** Do not use Tom House language.

## When to use

- "Create a mechanics report for [athlete]"
- "Stills are in [folder]"
- "Open Ben's report / add notes"
- "Rebuild / Download PDF"

## Layout

```
templates/reports/_master/          # HTML + serve.py template
templates/reports/<slug>/           # one athlete
  athlete.json                      # name, stills source, port
  notes.json                        # coach notes (source of truth)
  index.source.html                 # editable HTML
  index.html                        # embedded stills (generated)
  serve.py
  assets/                           # 8 stills
tools/mechanics-report/new_report.py
```

Canonical still names and fuzzy matching: [stills.md](stills.md)

## Create a new athlete

1. Get **athlete name** and **stills folder** (absolute path). Optional: session date, throws, age/team.
2. Run:

```bash
python3 tools/mechanics-report/new_report.py create \
  --athlete "Jane Doe" \
  --stills "/absolute/path/to/stills" \
  --serve
```

Optional flags: `--session "8/8/2026" --throws R --age-team "13/Premier 14u" --map balance=/a.jpg,lift=/b.jpg`

3. If the command exits `2`, mapping failed. Show the missing phases and files. Ask the user to rename stills or pass `--map`. Do not invent stills.
4. Hard-refresh the printed URL. Confirm notes load and stills show full frame (no crop).
5. Tell the user: type cues in the report; they auto-save to `notes.json`. Use **Download PDF** when done. Keep `http://127.0.0.1:<port>/index.html` (not `file://`).

## Edit an existing athlete

```bash
python3 tools/mechanics-report/new_report.py list
python3 tools/mechanics-report/new_report.py serve --slug jane-doe
```

If stills changed:

```bash
python3 tools/mechanics-report/new_report.py create \
  --athlete "Jane Doe" --slug jane-doe \
  --stills "/new/stills" --force
python3 tools/mechanics-report/new_report.py serve --slug jane-doe
```

`--force` overwrites stills and HTML. It keeps `notes.json`.

After HTML/CSS edits to `index.source.html`:

```bash
python3 tools/mechanics-report/new_report.py rebuild --slug jane-doe
```

## Agent rules

- Store everything under `templates/reports/<slug>/`. Local only.
- Do not draft coach cues. Leave note fields empty unless the user provides them.
- Do not mention Tom House. Framework chip is **NPA**.
- Cover photo is the release-point still.
- Print CSS is already in the master. Do not switch back to `object-fit: cover` on phase stills (it crops portrait frames).
- If port 8765 is in use, the server picks the next free port. Read the printed URL.
- Ben Staples (`templates/reports/ben-staples/`) is the first filled example. Do not wipe his `notes.json`.

## Done when

- Athlete folder exists with 8 stills, `notes.json`, `index.html`, `athlete.json`
- Local server is up and the user has the URL
- User can edit notes and Download PDF
