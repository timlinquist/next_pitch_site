#!/usr/bin/env python3
"""Local report server: serves this athlete HTML and persists notes.json on disk."""

from __future__ import annotations

import json
import socket
import threading
import webbrowser
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse

ROOT = Path(__file__).resolve().parent
NOTES_PATH = ROOT / "notes.json"
ATHLETE_PATH = ROOT / "athlete.json"

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


def load_athlete() -> dict:
    if ATHLETE_PATH.exists():
        return json.loads(ATHLETE_PATH.read_text(encoding="utf-8"))
    return {}


def athlete_name() -> str:
    meta = load_athlete()
    if meta.get("athlete"):
        return str(meta["athlete"])
    if NOTES_PATH.exists():
        notes = json.loads(NOTES_PATH.read_text(encoding="utf-8"))
        if notes.get("athlete"):
            return str(notes["athlete"])
    return "Athlete"


def empty_notes(name: str) -> dict:
    return {
        "athlete": name,
        "report": "mechanics-evaluation",
        "updatedAt": None,
        "sections": {key: "" for key in SECTION_KEYS},
        "ui": {"hiddenPriorities": []},
    }


def ensure_notes() -> None:
    name = athlete_name()
    if not NOTES_PATH.exists():
        NOTES_PATH.write_text(json.dumps(empty_notes(name), indent=2) + "\n", encoding="utf-8")
        return
    current = json.loads(NOTES_PATH.read_text(encoding="utf-8"))
    sections = current.setdefault("sections", {})
    changed = False
    for key in SECTION_KEYS:
        if key not in sections:
            sections[key] = ""
            changed = True
    current.setdefault("ui", {"hiddenPriorities": []})
    if not current.get("athlete"):
        current["athlete"] = name
        changed = True
    if changed:
        NOTES_PATH.write_text(json.dumps(current, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def resolve_port(preferred: int) -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        try:
            sock.bind(("127.0.0.1", preferred))
            return preferred
        except OSError:
            pass
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


class Handler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(ROOT), **kwargs)

    def end_headers(self):
        self.send_header("Cache-Control", "no-store")
        super().end_headers()

    def do_GET(self):
        path = urlparse(self.path).path
        if path in ("/notes.json", "/api/notes"):
            ensure_notes()
            data = NOTES_PATH.read_bytes()
            self.send_response(200)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
            return
        if path in ("/", "/index.html"):
            self.path = "/index.html"
        return super().do_GET()

    def do_POST(self):
        path = urlparse(self.path).path
        if path not in ("/notes.json", "/api/notes"):
            self.send_error(404)
            return
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length)
        try:
            payload = json.loads(raw.decode("utf-8"))
        except json.JSONDecodeError:
            self.send_error(400, "Invalid JSON")
            return

        ensure_notes()
        current = json.loads(NOTES_PATH.read_text(encoding="utf-8"))
        schema = empty_notes(athlete_name())
        sections = current.setdefault("sections", {})
        incoming = payload.get("sections", payload)
        if isinstance(incoming, dict):
            for key, value in incoming.items():
                if key in schema["sections"] or key.startswith(
                    ("cover.", "phase.", "priority.", "signoff.", "closing.")
                ):
                    sections[key] = value if isinstance(value, str) else str(value)
        current["athlete"] = payload.get("athlete", current.get("athlete", athlete_name()))
        current["report"] = payload.get("report", current.get("report", "mechanics-evaluation"))
        current["updatedAt"] = payload.get("updatedAt")
        if isinstance(payload.get("ui"), dict):
            current["ui"] = payload["ui"]
        NOTES_PATH.write_text(json.dumps(current, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

        body = json.dumps({"ok": True, "path": str(NOTES_PATH), "updatedAt": current["updatedAt"]}).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt, *args):
        print("[%s] %s" % (self.log_date_time_string(), fmt % args), flush=True)


def main():
    ensure_notes()
    meta = load_athlete()
    preferred = int(meta.get("port") or 8765)
    port = resolve_port(preferred)
    server = ThreadingHTTPServer(("127.0.0.1", port), Handler)
    url = f"http://127.0.0.1:{port}/index.html"
    print(f"Serving {ROOT}", flush=True)
    print(f"Athlete: {athlete_name()}", flush=True)
    print(f"Notes file: {NOTES_PATH}", flush=True)
    print(f"Open: {url}", flush=True)
    try:
        threading.Thread(target=webbrowser.open, args=(url,), daemon=True).start()
    except Exception as exc:
        print(f"Could not auto-open browser ({exc}). Open the URL manually.", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nStopped.", flush=True)


if __name__ == "__main__":
    main()
