#!/usr/bin/env python3
"""Create, rebuild, list, and serve THE NEXT PITCH mechanics evaluation reports."""

from __future__ import annotations

import argparse
import base64
import json
import mimetypes
import re
import shutil
import socket
import subprocess
import sys
import unicodedata
from datetime import datetime, timezone
from pathlib import Path

REPO = Path(__file__).resolve().parents[2]
MASTER = REPO / "templates" / "reports" / "_master"
REPORTS = REPO / "templates" / "reports"

IMAGE_EXTS = {".jpg", ".jpeg", ".png", ".webp"}

PHASES = (
    {
        "id": "balance",
        "file": "balance.jpg",
        "label": "Balance and Posture",
        "keywords": ("balance", "posture"),
        "numbers": ("01", "1"),
    },
    {
        "id": "lift",
        "file": "lift.jpg",
        "label": "Lift and Shift",
        "keywords": ("lift", "shift"),
        "numbers": ("02", "2"),
    },
    {
        "id": "thrust",
        "file": "thrust.jpg",
        "label": "Thrust",
        "keywords": ("thrust",),
        "numbers": ("03", "3"),
    },
    {
        "id": "equal-opposite",
        "file": "equal-opposite.jpg",
        "label": "Equal and Opposite",
        "keywords": ("equal", "opposite"),
        "numbers": ("04", "4"),
    },
    {
        "id": "delayed",
        "file": "delayed.jpg",
        "label": "Delayed Hips / Shoulder",
        "keywords": ("delayed", "hips", "separate"),
        "numbers": ("05", "5"),
    },
    {
        "id": "swivel",
        "file": "swivel.jpg",
        "label": "Swivel and Stabilize",
        "keywords": ("swivel", "stabilize"),
        "numbers": ("06", "6"),
    },
    {
        "id": "stack-track",
        "file": "stack-track.jpg",
        "label": "Stack and Track",
        "keywords": ("stack", "track"),
        "numbers": ("07", "7"),
    },
    {
        "id": "release-point",
        "file": "release-point.jpg",
        "label": "Release Point and Follow Through",
        "keywords": ("release", "follow"),
        "numbers": ("08", "8"),
    },
)

SECTION_KEYS = (
    "cover.session",
    "cover.throws",
    "cover.ageTeam",
    "phase.01.notes",
    "phase.02.notes",
    "phase.03.notes",
    "phase.04.notes",
    "phase.05.notes",
    "phase.06.notes",
    "phase.07.notes",
    "phase.08.notes",
    "priority.1",
    "priority.2",
    "priority.3",
    "priority.4",
    "priority.5",
    "signoff.coach",
    "signoff.date",
    "closing.commentary",
)


def slugify(name: str) -> str:
    normalized = unicodedata.normalize("NFKD", name).encode("ascii", "ignore").decode("ascii")
    slug = re.sub(r"[^a-z0-9]+", "-", normalized.lower()).strip("-")
    return slug or "athlete"


def normalize_stem(path: Path) -> str:
    return re.sub(r"[^a-z0-9]+", "-", path.stem.lower()).strip("-")


def list_images(folder: Path) -> list[Path]:
    return sorted(
        p for p in folder.iterdir() if p.is_file() and p.suffix.lower() in IMAGE_EXTS and not p.name.startswith(".")
    )


def score_candidate(phase: dict, image: Path) -> int:
    stem = normalize_stem(image)
    tokens = set(stem.split("-"))
    score = 0
    canonical = Path(phase["file"]).stem
    if stem == canonical or image.name.lower() == phase["file"]:
        return 100
    for number in phase["numbers"]:
        if stem == number or stem.startswith(number + "-"):
            score += 40
    hits = [kw for kw in phase["keywords"] if kw in tokens or kw in stem]
    if hits == list(phase["keywords"]):
        score += 50
    elif hits:
        score += 12 * len(hits)
    return score


def auto_map_stills(folder: Path) -> tuple[dict[str, Path], list[str]]:
    images = list_images(folder)
    unused = set(images)
    mapping: dict[str, Path] = {}
    missing: list[str] = []
    for phase in PHASES:
        ranked = sorted(
            ((score_candidate(phase, image), image) for image in unused),
            key=lambda item: (-item[0], item[1].name.lower()),
        )
        if ranked and ranked[0][0] >= 40:
            mapping[phase["id"]] = ranked[0][1]
            unused.discard(ranked[0][1])
        else:
            missing.append(f"{phase['id']} ({phase['label']}) -> {phase['file']}")
    return mapping, missing


def parse_map_arg(raw: str | None) -> dict[str, Path]:
    if not raw:
        return {}
    out: dict[str, Path] = {}
    for part in raw.split(","):
        if not part.strip():
            continue
        if "=" not in part:
            raise SystemExit(f"Invalid --map item '{part}'. Use phase=path.")
        key, value = part.split("=", 1)
        phase_id = key.strip()
        if phase_id not in {p["id"] for p in PHASES}:
            raise SystemExit(f"Unknown phase '{phase_id}'. Valid: {', '.join(p['id'] for p in PHASES)}")
        path = Path(value.strip()).expanduser().resolve()
        if not path.is_file():
            raise SystemExit(f"Still not found: {path}")
        out[phase_id] = path
    return out


def dest_for_phase(phase: dict, source: Path) -> str:
    suffix = source.suffix.lower()
    if suffix == ".jpeg":
        suffix = ".jpg"
    return f"{Path(phase['file']).stem}{suffix}"


def find_free_port(start: int = 8765) -> int:
    for port in range(start, start + 40):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            try:
                sock.bind(("127.0.0.1", port))
                return port
            except OSError:
                continue
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def report_dir(slug: str) -> Path:
    return REPORTS / slug


def render_source(athlete: str, slug: str) -> str:
    master = (MASTER / "index.source.html").read_text(encoding="utf-8")
    return master.replace("__ATHLETE_NAME__", athlete).replace("__ATHLETE_SLUG__", slug)


def rewrite_still_refs(html: str, dest_names: dict[str, str]) -> str:
    out = html
    for phase in PHASES:
        dest = dest_names[phase["id"]]
        canonical = phase["file"]
        if dest == canonical:
            continue
        out = out.replace(f"assets/{canonical}", f"assets/{dest}")
    return out


def embed_images(report: Path, source_html: str) -> str:
    def embed(match: re.Match[str]) -> str:
        rel = match.group(1)
        file_path = report / rel
        if not file_path.exists():
            raise SystemExit(f"Missing report image: {file_path}")
        mime = mimetypes.guess_type(file_path.name)[0] or "application/octet-stream"
        encoded = base64.b64encode(file_path.read_bytes()).decode("ascii")
        return f'src="data:{mime};base64,{encoded}"'

    return re.sub(r'src="(assets/[^"]+)"', embed, source_html)


def empty_notes(athlete: str) -> dict:
    return {
        "athlete": athlete,
        "report": "mechanics-evaluation",
        "updatedAt": None,
        "sections": {key: "" for key in SECTION_KEYS},
        "ui": {"hiddenPriorities": []},
    }


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path: Path, payload: dict) -> None:
    path.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def cmd_create(args: argparse.Namespace) -> int:
    athlete = args.athlete.strip()
    slug = args.slug.strip() if args.slug else slugify(athlete)
    stills_dir = Path(args.stills).expanduser().resolve()
    dest = report_dir(slug)

    if not stills_dir.is_dir():
        raise SystemExit(f"Stills folder not found: {stills_dir}")
    if not (MASTER / "index.source.html").exists() or not (MASTER / "serve.py").exists():
        raise SystemExit(f"Master template missing under {MASTER}")
    if dest.exists() and not args.force:
        raise SystemExit(f"Report already exists: {dest}\nUse --force to rebuild stills/HTML (notes.json is kept).")

    mapping, missing = auto_map_stills(stills_dir)
    mapping.update(parse_map_arg(args.map))
    still_missing = [p for p in PHASES if p["id"] not in mapping]
    if still_missing:
        print("Could not map these stills:", file=sys.stderr)
        for phase in still_missing:
            print(f"  - {phase['id']} ({phase['label']}) expected like {phase['file']}", file=sys.stderr)
        print("\nFiles in stills folder:", file=sys.stderr)
        for image in list_images(stills_dir):
            print(f"  - {image.name}", file=sys.stderr)
        print(
            "\nRename files to the canonical names, or pass --map "
            "balance=/path/a.jpg,lift=/path/b.jpg,...",
            file=sys.stderr,
        )
        return 2

    dest.mkdir(parents=True, exist_ok=True)
    assets = dest / "assets"
    assets.mkdir(exist_ok=True)

    dest_names: dict[str, str] = {}
    for phase in PHASES:
        source = mapping[phase["id"]]
        name = dest_for_phase(phase, source)
        shutil.copy2(source, assets / name)
        dest_names[phase["id"]] = name

    source_html = rewrite_still_refs(render_source(athlete, slug), dest_names)
    (dest / "index.source.html").write_text(source_html, encoding="utf-8")
    (dest / "index.html").write_text(embed_images(dest, source_html), encoding="utf-8")
    shutil.copy2(MASTER / "serve.py", dest / "serve.py")

    notes_path = dest / "notes.json"
    if notes_path.exists() and args.force:
        notes = load_json(notes_path)
        notes["athlete"] = athlete
        for key in SECTION_KEYS:
            notes.setdefault("sections", {}).setdefault(key, "")
        notes.setdefault("ui", {}).setdefault("hiddenPriorities", [])
        write_json(notes_path, notes)
    elif not notes_path.exists():
        write_json(notes_path, empty_notes(athlete))

    if args.session or args.throws or args.age_team:
        notes = load_json(notes_path)
        sections = notes.setdefault("sections", {})
        if args.session:
            sections["cover.session"] = args.session
        if args.throws:
            sections["cover.throws"] = args.throws
        if args.age_team:
            sections["cover.ageTeam"] = args.age_team
        write_json(notes_path, notes)

    meta_path = dest / "athlete.json"
    port = args.port or (load_json(meta_path).get("port") if meta_path.exists() else None) or find_free_port()
    write_json(
        meta_path,
        {
            "athlete": athlete,
            "slug": slug,
            "report": "mechanics-evaluation",
            "createdAt": datetime.now(timezone.utc).isoformat(),
            "stillsSource": str(stills_dir),
            "coverStill": dest_names["release-point"],
            "stills": dest_names,
            "port": int(port),
        },
    )

    print(f"Created {dest}")
    print(f"Stills source: {stills_dir}")
    for phase in PHASES:
        print(f"  {phase['id']:16} {mapping[phase['id']].name} -> assets/{dest_names[phase['id']]}")
    print(f"Notes: {notes_path}")
    print(f"Serve: python3 {dest / 'serve.py'}")
    print(f"Or:    python3 {Path(__file__)} serve --slug {slug}")
    if args.serve:
        return cmd_serve(argparse.Namespace(slug=slug))
    return 0


def cmd_rebuild(args: argparse.Namespace) -> int:
    dest = report_dir(args.slug)
    source = dest / "index.source.html"
    if not source.exists():
        raise SystemExit(f"No report source at {source}")
    html = source.read_text(encoding="utf-8")
    (dest / "index.html").write_text(embed_images(dest, html), encoding="utf-8")
    print(f"Rebuilt {dest / 'index.html'}")
    return 0


def cmd_list(_args: argparse.Namespace) -> int:
    rows = []
    for folder in sorted(p for p in REPORTS.iterdir() if p.is_dir() and not p.name.startswith("_")):
        meta = load_json(folder / "athlete.json") if (folder / "athlete.json").exists() else {}
        notes = load_json(folder / "notes.json") if (folder / "notes.json").exists() else {}
        name = meta.get("athlete") or notes.get("athlete") or folder.name
        updated = notes.get("updatedAt") or meta.get("createdAt") or ""
        port = meta.get("port", "")
        rows.append((folder.name, name, str(port), str(updated)))
    if not rows:
        print("No athlete reports yet.")
        return 0
    print(f"{'slug':20} {'athlete':24} {'port':6} updated")
    for slug, name, port, updated in rows:
        print(f"{slug:20} {name:24} {str(port):6} {updated}")
    return 0


def cmd_serve(args: argparse.Namespace) -> int:
    dest = report_dir(args.slug)
    server = dest / "serve.py"
    if not server.exists():
        raise SystemExit(f"No serve.py in {dest}. Create the report first.")
    index = dest / "index.html"
    if not index.exists() and (dest / "index.source.html").exists():
        cmd_rebuild(argparse.Namespace(slug=args.slug))
    print(f"Starting server for {args.slug}...")
    return subprocess.call([sys.executable, str(server)], cwd=str(dest))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="THE NEXT PITCH mechanics evaluation reports")
    sub = parser.add_subparsers(dest="command", required=True)

    create = sub.add_parser("create", help="Create a new athlete report from stills")
    create.add_argument("--athlete", required=True, help='Athlete name, e.g. "Jane Doe"')
    create.add_argument("--stills", required=True, help="Folder of delivery stills")
    create.add_argument("--slug", help="Folder name under templates/reports/ (default: slugified athlete)")
    create.add_argument("--map", help="Override still mapping: balance=/a.jpg,lift=/b.jpg,...")
    create.add_argument("--session", help="Pre-fill cover.session")
    create.add_argument("--throws", help="Pre-fill cover.throws")
    create.add_argument("--age-team", dest="age_team", help="Pre-fill cover.ageTeam")
    create.add_argument("--port", type=int, help="Preferred local server port")
    create.add_argument("--force", action="store_true", help="Overwrite stills/HTML; keep notes.json")
    create.add_argument("--serve", action="store_true", help="Start the local notes server after create")
    create.set_defaults(func=cmd_create)

    rebuild = sub.add_parser("rebuild", help="Rebuild index.html with embedded stills")
    rebuild.add_argument("--slug", required=True)
    rebuild.set_defaults(func=cmd_rebuild)

    listing = sub.add_parser("list", help="List local athlete reports")
    listing.set_defaults(func=cmd_list)

    serve = sub.add_parser("serve", help="Serve an existing athlete report")
    serve.add_argument("--slug", required=True)
    serve.set_defaults(func=cmd_serve)
    return parser


def main() -> int:
    args = build_parser().parse_args()
    return int(args.func(args) or 0)


if __name__ == "__main__":
    raise SystemExit(main())
