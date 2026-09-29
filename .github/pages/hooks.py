"""MkDocs hooks: publish root Markdown files as pages and fix repo-relative links."""

import os
import posixpath
import re

from mkdocs.structure.files import File

REPO_URL = "https://github.com/HarryRaddatz/argus-observability"
BRANCH = "main"
DOCS_DIR = "docs"

# Site page -> repository file used as its source.
ROOT_PAGES = {
    "index.md": "README.md",
    "changelog.md": "CHANGELOG.md",
    "contributing.md": "CONTRIBUTING.md",
}

LINK_RE = re.compile(r"(!?\[[^\]]*\]\()([^)\s]+)(\))")
SKIP_PREFIXES = ("http://", "https://", "mailto:", "#", "/")

_repo_root = ""


def _in_docs(repo_path, is_dir):
    if not repo_path.startswith(DOCS_DIR + "/"):
        return False
    if is_dir:
        return os.path.isfile(os.path.join(_repo_root, repo_path, "README.md"))
    return True


def on_config(config):
    global _repo_root
    _repo_root = os.path.dirname(os.path.abspath(config["config_file_path"]))
    return config


def on_files(files, config):
    for page, source in ROOT_PAGES.items():
        with open(os.path.join(_repo_root, source), encoding="utf-8") as fh:
            content = fh.read()
        files.append(File.generated(config, page, content=content))
    return files


def on_page_markdown(markdown, page, config, files):
    src_uri = page.file.src_uri
    source = ROOT_PAGES.get(src_uri)
    base_dir = posixpath.dirname(source) if source else posixpath.join(DOCS_DIR, posixpath.dirname(src_uri))
    page_dir = posixpath.dirname(src_uri)
    if source:
        page.edit_url = f"{REPO_URL}/edit/{BRANCH}/{source}"

    def rewrite(match):
        target = match.group(2)
        if target.startswith(SKIP_PREFIXES):
            return match.group(0)
        path, sep, anchor = target.partition("#")
        repo_path = posixpath.normpath(posixpath.join(base_dir, path))
        is_dir = path.endswith("/") or os.path.isdir(os.path.join(_repo_root, repo_path))

        if repo_path in ROOT_PAGES.values():
            page_uri = next(k for k, v in ROOT_PAGES.items() if v == repo_path)
            new = posixpath.relpath(page_uri, page_dir or ".")
        elif _in_docs(repo_path, is_dir):
            if is_dir:
                repo_path = posixpath.join(repo_path, "README.md")
            docs_rel = posixpath.relpath(repo_path, DOCS_DIR)
            new = posixpath.relpath(docs_rel, page_dir or ".")
        else:
            kind = "tree" if is_dir else "blob"
            new = f"{REPO_URL}/{kind}/{BRANCH}/{repo_path}"

        return f"{match.group(1)}{new}{sep}{anchor}{match.group(3)}"

    return LINK_RE.sub(rewrite, markdown)
