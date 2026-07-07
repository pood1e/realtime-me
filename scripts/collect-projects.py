#!/usr/bin/env python3
"""One-time collector for the profile page's GitHub projects.

Fetches the owner's repositories from GitHub, optionally writes a short summary
for each with Claude, and stores the result as static data that the status
gateway serves at /api/profile. Run it once (re-run only when you want to
refresh); the gateway itself never calls GitHub or Claude.

Environment:
  GITHUB_TOKEN       Include private repositories and raise rate limits. Without
                     it, only public repositories of --login are collected.
  ANTHROPIC_API_KEY  Optional. When set, a one- or two-sentence summary is
                     generated per repository with Claude.

Examples:
  # Collect all owned repos into the gateway's profile config, with summaries.
  GITHUB_TOKEN=... ANTHROPIC_API_KEY=... \\
    ./scripts/collect-projects.py --config /data/profile.json

  # Only specific repos, printed to stdout (paste into the config's "projects").
  GITHUB_TOKEN=... ./scripts/collect-projects.py --repos realtime-me,dotfiles
"""
from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from typing import Any

GITHUB_API = "https://api.github.com"
DEFAULT_MODEL = "claude-opus-4-8"
README_LIMIT = 24 * 1024
SUMMARY_SYSTEM = (
    "You write concise summaries of software projects for a personal profile page. "
    "Given a repository's metadata and README, write one or two plain sentences "
    "(at most 45 words) describing what the project does and why it is interesting. "
    "Do not use marketing language, headings, bullet points, or Markdown. "
    "Respond with the summary text only."
)


def main() -> int:
    args = parse_args()
    token = os.getenv("GITHUB_TOKEN", "").strip()
    if not token and not args.login:
        print("Set GITHUB_TOKEN, or pass --login for public repositories only.", file=sys.stderr)
        return 1

    try:
        repos = select_repos(fetch_repos(token, args.login), args)
    except urllib.error.HTTPError as error:
        print(f"GitHub request failed: HTTP {error.code}", file=sys.stderr)
        return 1
    except OSError as error:
        print(f"GitHub request failed: {error.__class__.__name__}", file=sys.stderr)
        return 1

    summarize = build_summarizer(args.model)
    projects = []
    for repo in repos:
        summary = summarize(repo, fetch_readme(token, repo)) if summarize else ""
        projects.append(project_entry(repo, summary))

    emit(projects, args)
    print(f"Collected {len(projects)} project(s).", file=sys.stderr)
    return 0


def select_repos(repos: list[dict[str, Any]], args: argparse.Namespace) -> list[dict[str, Any]]:
    repos = [repo for repo in repos if include_repo(repo, args)]
    if args.repos:
        wanted = [name.strip().lower() for name in args.repos.split(",") if name.strip()]
        index = {repo["name"].lower(): repo for repo in repos}
        return [index[name] for name in wanted if name in index]
    repos.sort(key=lambda repo: repo.get("pushed_at") or "", reverse=True)
    return repos


def include_repo(repo: dict[str, Any], args: argparse.Namespace) -> bool:
    if repo.get("fork") and not args.include_forks:
        return False
    if repo.get("archived") and not args.include_archived:
        return False
    return True


def fetch_repos(token: str, login: str) -> list[dict[str, Any]]:
    if token:
        base = f"{GITHUB_API}/user/repos?affiliation=owner&visibility=all&sort=pushed"
    else:
        base = f"{GITHUB_API}/users/{login}/repos?type=owner&sort=pushed"
    repos: list[dict[str, Any]] = []
    for page in range(1, 11):
        batch = github_json(f"{base}&per_page=100&page={page}", token)
        if not isinstance(batch, list) or not batch:
            break
        repos.extend(batch)
        if len(batch) < 100:
            break
    return repos


def fetch_readme(token: str, repo: dict[str, Any]) -> str:
    full_name = repo.get("full_name")
    if not full_name:
        return ""
    request = urllib.request.Request(
        f"{GITHUB_API}/repos/{full_name}/readme",
        headers=github_headers(token, accept="application/vnd.github.raw"),
    )
    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            return response.read(README_LIMIT).decode("utf-8", "replace")
    except (urllib.error.HTTPError, OSError):
        return ""


def github_json(url: str, token: str) -> Any:
    request = urllib.request.Request(url, headers=github_headers(token))
    with urllib.request.urlopen(request, timeout=15) as response:
        return json.loads(response.read().decode())


def github_headers(token: str, accept: str = "application/vnd.github+json") -> dict[str, str]:
    headers = {
        "Accept": accept,
        "User-Agent": "realtime-me-collect-projects",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def project_entry(repo: dict[str, Any], summary: str) -> dict[str, Any]:
    return {
        "display_name": repo.get("name", ""),
        "description": repo.get("description") or "",
        "summary": summary,
        "visibility": "private" if repo.get("private") else "public",
        "primary_language": repo.get("language") or "",
        "topics": repo.get("topics") or [],
        "star_count": int(repo.get("stargazers_count") or 0),
        "repository_url": repo.get("html_url") or "",
        "homepage_url": repo.get("homepage") or "",
        "last_push_time": repo.get("pushed_at") or "",
    }


def build_summarizer(model: str):
    key = os.getenv("ANTHROPIC_API_KEY", "").strip()
    if not key:
        return None
    try:
        from anthropic import Anthropic
    except ImportError:
        print("ANTHROPIC_API_KEY set but the `anthropic` package is missing; run `pip install anthropic`.", file=sys.stderr)
        return None

    client = Anthropic(api_key=key)

    def summarize(repo: dict[str, Any], readme: str) -> str:
        message = client.messages.create(
            model=model,
            max_tokens=256,
            system=SUMMARY_SYSTEM,
            messages=[{"role": "user", "content": summary_prompt(repo, readme)}],
        )
        return "".join(block.text for block in message.content if block.type == "text").strip()

    return summarize


def summary_prompt(repo: dict[str, Any], readme: str) -> str:
    lines = [f"Repository: {repo.get('full_name') or repo.get('name', '')}"]
    if repo.get("description"):
        lines.append(f"Description: {repo['description']}")
    if repo.get("language"):
        lines.append(f"Primary language: {repo['language']}")
    if repo.get("topics"):
        lines.append(f"Topics: {', '.join(repo['topics'])}")
    if readme:
        lines.append(f"\nREADME:\n{readme}")
    return "\n".join(lines)


def emit(projects: list[dict[str, Any]], args: argparse.Namespace) -> None:
    if args.config:
        config: dict[str, Any] = {}
        if os.path.exists(args.config):
            with open(args.config, encoding="utf-8") as handle:
                config = json.load(handle)
        config["projects"] = projects
        with open(args.config, "w", encoding="utf-8") as handle:
            json.dump(config, handle, ensure_ascii=False, indent=2)
            handle.write("\n")
        return
    text = json.dumps(projects, ensure_ascii=False, indent=2)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as handle:
            handle.write(text + "\n")
    else:
        print(text)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Collect GitHub projects for the profile page (one-time).")
    parser.add_argument("--login", default=os.getenv("GITHUB_LOGIN", ""), help="GitHub login for public-only mode (no token).")
    parser.add_argument("--repos", default="", help="Comma-separated repository names to include, in order.")
    parser.add_argument("--config", default="", help="Profile config file to update in place (sets its \"projects\").")
    parser.add_argument("--output", default="", help="Write the projects JSON array to this file instead of stdout.")
    parser.add_argument("--model", default=os.getenv("ANTHROPIC_MODEL", DEFAULT_MODEL), help="Claude model for summaries.")
    parser.add_argument("--include-forks", action="store_true", help="Include forked repositories.")
    parser.add_argument("--include-archived", action="store_true", help="Include archived repositories.")
    return parser.parse_args()


if __name__ == "__main__":
    sys.exit(main())
