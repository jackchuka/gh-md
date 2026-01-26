# gh-md

[![Test](https://github.com/jackchuka/gh-md/workflows/Test/badge.svg)](https://github.com/jackchuka/gh-md/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jackchuka/gh-md)](https://goreportcard.com/report/github.com/jackchuka/gh-md)

A GitHub CLI extension that syncs GitHub Issues, Pull Requests, and Discussions to local markdown files with bidirectional support.

## Features

- **Pull** GitHub data as markdown files with YAML frontmatter
- **Push** local changes back to GitHub (title, body, state, comments)
- **Search** local files interactively with FZF
- **Conflict detection** prevents overwriting newer remote changes
- **AI-friendly** format ideal for use with coding assistants and local tools

## Installation

```bash
gh extension install jackchuka/gh-md
```

Requires [GitHub CLI](https://cli.github.com/) with authentication (`gh auth login`).

## Usage

### Pull

Fetch GitHub data and save as local markdown files.

```bash
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

# Pull a specific item by URL
gh md pull https://github.com/owner/repo/issues/123
```

### Push

Push local markdown changes back to GitHub.

```bash
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

### Search

Interactively search local files with [FZF](https://github.com/junegunn/fzf).

```bash
# Search all local files
gh md search

# Search within a specific repository
gh md search owner/repo

# Filter by type
gh md search --issues
gh md search --prs
gh md search --discussions

# Non-interactive list mode
gh md search --list
```

**Actions after selection:**

- Open in `$EDITOR`
- Push changes to GitHub
- View in browser
- Copy file path
- Pull fresh from GitHub

Requires [FZF](https://github.com/junegunn/fzf) to be installed (`brew install fzf`).

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
