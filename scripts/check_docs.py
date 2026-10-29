#!/usr/bin/env python3
"""Check current local Markdown links and collapsed details; no external requests."""

from pathlib import Path
import re
import sys
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
# Immutable snapshots retain links in their original repository-root context.
SNAPSHOTS = {
    ROOT / "docs/history/README-before-step-74.md",
    ROOT / "docs/history/DESIGN-before-step-74.md",
}


def prose(text):
    text = re.sub(r"^```.*?^```[^\n]*", "", text, flags=re.M | re.S)
    return re.sub(r"`[^`\n]*`", "", text)


def anchors(text):
    result = set()
    counts = {}
    for heading in re.findall(r"^#{1,6}\s+(.+)$", prose(text), re.M):
        slug = re.sub(r"[^\w\- ]", "", heading.lower()).replace(" ", "-")
        count = counts.get(slug, 0)
        counts[slug] = count + 1
        result.add(f"{slug}-{count}" if count else slug)
    result.update(re.findall(r'\bid=["\']([^"\']+)["\']', text))
    return result


def main():
    errors = []
    files = [*sorted(ROOT.glob("*.md")), *sorted((ROOT / "docs").rglob("*.md"))]
    for path in files:
        text = prose(path.read_text())
        label = path.relative_to(ROOT)
        depth = 0
        for tag in re.findall(r"</?details\b[^>]*>", text):
            depth += -1 if tag.startswith("</") else 1
            if depth < 0:
                errors.append(f"{label}: unmatched </details>")
            if re.search(r"\bopen\b", tag):
                errors.append(f"{label}: details must be collapsed by default")
        if depth or text.count("<summary>") != text.count("</summary>") or text.count("<summary>") != len(re.findall(r"<details\b", text)):
            errors.append(f"{label}: unbalanced details/summary sections")
        if path in SNAPSHOTS:
            continue
        for target in re.findall(r"\[[^\]]*\]\(([^)]+)\)", text):
            url = urlsplit(target)
            if url.scheme or url.netloc:
                continue
            destination = (path.parent / unquote(url.path)).resolve() if url.path else path
            if not destination.exists():
                errors.append(f"{label}: missing link target {target}")
            elif url.fragment and destination.suffix == ".md" and unquote(url.fragment) not in anchors(destination.read_text()):
                errors.append(f"{label}: missing anchor {target}")
    if errors:
        print("\n".join(errors), file=sys.stderr)
        return 1
    print(f"Checked {len(files)} Markdown files; current local links and collapsed sections passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
