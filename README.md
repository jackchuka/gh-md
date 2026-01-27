# gh-md

[![Test](https://github.com/jackchuka/gh-md/workflows/Test/badge.svg)](https://github.com/jackchuka/gh-md/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jackchuka/gh-md)](https://goreportcard.com/report/github.com/jackchuka/gh-md)

A GitHub CLI extension that syncs GitHub Issues, Pull Requests, and Discussions to local markdown files with bidirectional support.

## Features

- **Smart context detection** - Commands auto-detect your git repo and branch
- **Pull** GitHub data as markdown files with YAML frontmatter
- **Push** local changes back to GitHub (title, body, state, comments)
- **Browse** local files interactively with FZF and CEL filtering
- **Prune** delete closed/merged items to keep your workspace clean
- **Conflict detection** prevents overwriting newer remote changes
- **AI-friendly** format ideal for use with coding assistants and local tools

## Installation

```bash
gh extension install jackchuka/gh-md
```

Requires [GitHub CLI](https://cli.github.com/) with authentication (`gh auth login`).

## Quickstart

```bash
# Pull issues from a repo
gh md pull owner/repo --issues

# Browse and select a file
gh md

# Edit in your editor, then push changes
gh md push ~/.gh-md/owner/repo/issues/123.md
```

## Smart Context Detection

When run inside a git repository, commands automatically detect your repo and branch:

| Command       | On feature branch with PR | On main/master       | Outside git repo      |
| ------------- | ------------------------- | -------------------- | --------------------- |
| `gh md`       | FZF pre-filtered to PR    | FZF filtered to repo | Show all items        |
| `gh md pull`  | Pull that PR with reviews | Pull all repo items  | Requires `owner/repo` |
| `gh md push`  | FZF filtered to repo      | FZF filtered to repo | Requires file path    |
| `gh md prune` | Prune current repo        | Prune current repo   | Prune all repos       |

```bash
# Inside a git repo on a feature branch with an open PR:
gh md pull    # Pulls your PR with all review comments
gh md         # Opens FZF pre-filtered to your PR
gh md push    # Opens FZF to select file from current repo
gh md prune   # Prunes closed items from current repo
```

## Usage

### Browse (default)

Interactively browse local files with [FZF](https://github.com/junegunn/fzf).

When run inside a git repo, FZF is pre-filtered to your current repo (or PR if on a feature branch). Clear the query to see all items.

```bash
# Smart: pre-filtered to current repo/PR
gh md

# Browse within a specific repository
gh md owner/repo

# Filter by type
gh md --issues
gh md --prs
gh md --discussions

# Filter by status
gh md --new          # Items updated since last pull
gh md --assigned     # Items assigned to you

# Advanced filtering with CEL expressions
gh md --filter 'state == "open"'
gh md --filter 'labels.exists(l, l == "bug")'
gh md --filter 'created > now - duration("168h")'  # Last 7 days

# Non-interactive list mode
gh md --list
gh md --list --format=json    # Output as JSON
gh md --list --format=yaml    # Output as YAML
```

**Actions after selection:**

- Open in `$EDITOR`
- Push changes to GitHub
- View in browser
- Copy file path
- Pull fresh from GitHub

Requires [FZF](https://github.com/junegunn/fzf) to be installed (`brew install fzf`).

### Pull

Fetch GitHub data and save as local markdown files. Uses incremental sync by default.

When run without arguments inside a git repo:

- On a feature branch with PR: pulls that PR with all review comments
- On main/master: pulls all items for the current repository

```bash
# Smart: pull based on git context
gh md pull

# Pull all issues, PRs, and discussions from a repository
gh md pull owner/repo

# Pull only issues
gh md pull owner/repo --issues

# Pull only pull requests
gh md pull owner/repo --prs

# Pull only discussions
gh md pull owner/repo --discussions

# Limit the number of items
gh md pull owner/repo --issues --limit 10

# Pull only open items
gh md pull owner/repo --open-only

# Force full sync (ignore last sync timestamp)
gh md pull owner/repo --full

# Pull a specific item by URL
gh md pull https://github.com/owner/repo/issues/123

# Pull all previously synced repositories
gh md pull --all
```

### Push

Push local markdown changes back to GitHub.

When run without arguments inside a git repo, opens FZF to select a file from the current repo.

```bash
# Smart: FZF selector for current repo
gh md push

# Push changes from a local file
gh md push owner/repo/issues/123.md

# Preview changes without pushing
gh md push --dry-run owner/repo/issues/123.md

# Force push even if remote has newer changes
gh md push --force owner/repo/issues/123.md
```

**What you can push:**

- Title and body changes
- State changes (open/closed)
- New comments
- Edited comments

### Prune

Delete local files for closed issues and merged/closed PRs.

When run inside a git repo, defaults to pruning only the current repository.

```bash
# Smart: prune current repo (dry-run)
gh md prune

# Actually delete files
gh md prune --confirm

# Prune a specific repository
gh md prune owner/repo --confirm

# Output as JSON or YAML
gh md prune --format=json
gh md prune --format=yaml
```

### Repos

List all repositories that have been synced with gh-md.

```bash
# List all managed repositories
gh md repos

# Output as JSON or YAML
gh md repos --format=json
gh md repos --format=yaml
```

## File Format

Files are stored as markdown with YAML frontmatter:

```markdown
---
id: I_abc123
url: https://github.com/owner/repo/issues/123
number: 123
owner: owner
repo: repo
title: Example Issue
state: open
labels: [bug, help wanted]
assignees: [octocat]
created: 2026-01-01T00:00:00Z
updated: 2026-01-24T12:00:00Z
last_pulled: 2026-01-24T12:30:00Z
---

<!-- gh-md:content -->

# Example Issue

Issue body content here.

<!-- /gh-md:content -->

---

<!-- gh-md:comment
id: IC_def456
author: octocat
created: 2026-01-02T00:00:00Z
-->

### @octocat (2026-01-02)

This is a comment.

<!-- /gh-md:comment -->

<!-- gh-md:new-comment -->

Add new comments here.

<!-- /gh-md:new-comment -->
```

## Storage Location

Files are stored in `~/.gh-md/` by default:

```
~/.gh-md/
  owner/
    repo/
      issues/
        123.md
      pulls/
        456.md
      discussions/
        789.md
```

Override with the `GH_MD_ROOT` environment variable:

```bash
export GH_MD_ROOT=/path/to/custom/directory
```

## Use Cases

- **AI Assistants**: Provide context from GitHub issues and PRs to coding assistants
- **Offline Access**: Browse and edit GitHub content without internet
- **Bulk Editing**: Make changes to multiple items locally, then push
- **Backup**: Keep local copies of important discussions

## License

[MIT License](LICENSE)
