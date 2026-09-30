"""MkDocs hooks: publish root Markdown files as pages and fix repo-relative links."""

import os
import posixpath
import re

REPO_URL = "https://github.com/HarryRaddatz/argus-observability"
BRANCH = "main"
DOCS_DIR = "docs"

# Site page -> repository file used as its source.
# Written into docs/ at build time so mkdocs-static-i18n can read them from disk.
ROOT_PAGES = {
    "index.md": "README.md",
    "index.pt.md": "README.pt-BR.md",
    "changelog.md": "CHANGELOG.md",
    "contributing.md": "CONTRIBUTING.md",
    "contributing.pt.md": "CONTRIBUTING.pt-BR.md",
}

SITE_FOR_ROOT = {
    "README.md": "/",
    "README.pt-BR.md": "/pt/",
    "CONTRIBUTING.md": "/contributing/",
    "CONTRIBUTING.pt-BR.md": "/pt/contributing/",
    "CHANGELOG.md": "/changelog/",
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
    docs_dir = config["docs_dir"]
    for page, source in ROOT_PAGES.items():
        dest = os.path.join(docs_dir, page)
        with open(os.path.join(_repo_root, source), encoding="utf-8") as fh:
            content = fh.read()
        with open(dest, "w", encoding="utf-8") as out:
            out.write(content)
    return config


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

        if repo_path in SITE_FOR_ROOT:
            new = SITE_FOR_ROOT[repo_path]
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
