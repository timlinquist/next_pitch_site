---
name: mechanics-evaluation-report
description: Create, edit, and reopen THE NEXT PITCH NPA mechanics evaluation HTML reports from athlete stills in Google Drive. Use for athlete mechanics reports, coach notes, and PDF preparation.
---

# Mechanics evaluation report

Use the repository CLI at `tools/mechanics-report/new_report.py`. Cursor, Claude, and Codex share this skill through repository skill directories. Run commands from the repository root. Keep this workflow separate from the React/Go application.

## Storage

Google Drive is the source of truth:

- `Next Pitch LLC/<athlete-name>/`: source stills, or a verified stills subfolder.
- `Next Pitch LLC/<athlete-name>/evals/<YYYY-MM-DD>/`: evaluation HTML, athlete metadata, notes, and a report archive with assets.

Resolve folder IDs through the connected Drive tools. Do not infer IDs from names. Reuse existing athlete folders, including legacy names with hyphens. Use the evaluation date, not the upload date. Ask only when the date or source remains uncertain.

Use a Drive-synced directory when available. Otherwise download authenticated stills to a temporary directory and stage the report outside Git. The ignored `templates/reports/` directory is a local cache only. Never commit athlete names, notes, stills, report archives, or Drive manifests. Reusable HTML lives at `tools/mechanics-report/template/`; the only server lives at `tools/mechanics-report/serve.py`.

## Create

1. Confirm the athlete name, evaluation date, and Drive source folder. List its contents. Download the actual stills with authenticated Drive tools. Preserve the full frames. See [stills.md](stills.md) for phase mapping.
2. Create a report in a temporary or Drive-synced evaluation directory:

```bash
python3 tools/mechanics-report/new_report.py --reports-root "/external/evals/2026-11-20" create \
  --athlete "Jane Doe" --stills "/temporary/stills" --session "2026-11-20"
```

3. If mapping fails, inspect the missing phases. Use `--map phase=/path/image.jpg` only after verification. Never substitute an unrelated still.
4. Keep coach cues empty unless the user supplies them. Use NPA language. Use the release-point still on the cover. Do not invent age, team, handedness, or mechanical findings.
5. Serve through the shared server:

```bash
python3 tools/mechanics-report/new_report.py --reports-root "/external/evals/2026-11-20" serve --slug jane-doe
```

The server binds to loopback and selects a free port. Read the printed URL. Validate all eight phase images, the athlete name, date, note load/save, and print layout. Preserve the complete delivery frame inside each image view; remove phone interface areas as described below.
6. Upload `index.html`, a final PDF, `notes.json`, `athlete.json`, and an archive of the complete report directory to the verified Drive evaluation folder. The archive includes `index.source.html` and assets; it excludes server code. Verify uploaded file names and sizes through Drive readback. Record the source folder URL in the external athlete metadata. Give the user the verified Drive folder link.

## Phone screenshots and final review

Inspect every still before the first review. Phone screenshots often include black margins, status bars, playback buttons, and thumbnail strips. Crop the report view to the actual video frame. Preserve the athlete’s entire body, throwing hand, feet, mound, and useful field context. Keep original Drive files unchanged.

Use non-destructive HTML/CSS image viewports for report crops. Determine bounds separately for every image from its actual dimensions and visible frame. Do not reuse fixed coordinates across phones, orientations, or camera views. A portrait camera view needs a portrait viewport; constrain its display size without cutting off the athlete. Apply the same crop to the cover still. Verify both screen and print CSS, especially rules that override image height or position.

Wait for the user to finish the notes. Keep their editor tab and server open during entry and review. Never treat inactivity as completion. Read the saved notes after the user confirms completion.

Create a separate standalone review HTML with the exact saved notes embedded in its markup, embedded images, approved crops, and hidden priorities removed. Remove editor scripts from this static copy so a server or browser storage cannot overwrite its notes. Preserve the editable report separately. Show the review before final publication when the user requests review.

A final report always includes a PDF. Export the approved static HTML with its print styles, fonts, embedded notes, and cropped image views. Use a descriptive filename such as `<athlete-slug>-<evaluation-date>-report.pdf`. Enable background graphics and disable browser URL/date headers and footers.

Render and inspect every PDF page before upload. Check the athlete identity, evaluation date, all eight checkpoints, all saved notes, priorities, and signoff. Check for missing images, clipped text, blank pages, unexpected phase splits, and phone interface areas. Print rules can override viewport image positions; responsive screen rules can stack phase columns in print. Correct those print-specific rules and regenerate when needed. Match the page labels to the actual PDF pages. Keep a reasonable readable type size.

After approval, upload the final PDF and the reviewed static HTML as `index.html`, the latest `notes.json`, and the refreshed archive. Include the editable HTML as `editor.html`, plus source HTML and original assets. When reopening, use `editor.html` for note entry (or restore it as the local `index.html` first); regenerate the static final report after edits. Include the final PDF in the archive too. Verify that the PDF, archive, and standalone HTML use the same approved notes, identity, and crops. Verify the uploaded PDF name, nonzero size, and evaluation folder through Drive readback.

## Edit and reopen

Download and extract the saved report archive outside Git. Restore the latest separately saved `notes.json` before opening. Preserve all prior notes. Never use `--force` without a backup of the existing report.

```bash
python3 tools/mechanics-report/serve.py /external/report/jane-doe
python3 tools/mechanics-report/new_report.py --reports-root /external/report rebuild --slug jane-doe
```

The browser saves notes to the local report directory. Upload the latest notes and refreshed report archive after edits. A local save alone does not complete a Drive save. If Drive upload fails, retain the local report and report the failure. Do not claim that Drive contains it.

Use the browser's Download PDF action for an editor PDF, or print the standalone final HTML. The static final HTML must include saved notes without the server. Keep the JSON and archive beside it.

## Completion

Confirm eight stills, the supplied evaluation date, preserved notes, one shared server, and a verified Drive save that includes the final PDF. Do not mark the report complete when only HTML or an archive exists. Check `git status` to ensure that no athlete artifacts enter the commit.
