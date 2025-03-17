# Today I Learned

TIL (Today I Learned) is a command-line application for tracking what you learned today.

The intended audience is software developers. In that vein, it provides a `git`-like interface for adding entries, and the option of syncing them with version control or external sources (e.g., Notion).

- [Today I Learned](#today-i-learned)
  - [Features](#features)
  - [Interface](#interface)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Initialize a repository](#initialize-a-repository)
    - [Add files](#add-files)
    - [Commit an entry](#commit-an-entry)
    - [Status](#status)
    - [Amend a commit](#amend-a-commit)
    - [Push to Notion](#push-to-notion)
    - [View the log](#view-the-log)
  - [License](#license)

## Features

- Append new entries daily with a familiar, `git`-like interface
- Provide file attachments to your entries (e.g., code snippets, images)
- Sync data to an external source; GitHub at minimum is highly recommended, and you may also do Notion

## Interface

- `til init`, initialize a new TIL repository or sync with an existing one
- `til add <files>`, stage files for the current log entry
- `til commit -m "<message>"`, add a new log entry (message is required, files are optional; more than one commit is allowed per day)
- `til status`, see what is staged and what commits are outstanding
- `til commit --amend`, amend the previous commit
- `til push`, submit all outstanding commits to external sources
- `til log -n <number>`, view a one-line log of past learnings (with filenames)

The interface is intentionally limited. Want to edit a previous entry? You can't. Though you're welcome to edit the synced entry (e.g., the listing in Notion), and your updates will never be overwritten.

The goal is to have your `til` log be effectively write-on-the-day-of only, whereas your "published" version can be tidied up.

## Prerequisites

- Create an empty repository on GitHub or similar
- (optionally) create a Notion database with a schema of `TIL` (title) and `Attachments` (files)

## Installation

```bash
go install github.com/michaelfromorg/tiled@latest
```

Or build from source:

```bash
git clone https://github.com/michaelfromorg/tiled.git
cd til
go build
```

## Usage

Here's a walk-through of all the available commands.

### Initialize a repository

```bash
til init
```

This step requires a remote URL to a repository.

This will prompt you to configure Notion sync. If you enable Notion sync, you'll need to provide your Notion API key and database ID.

### Add files

```bash
til add file1.txt file2.txt
```

This stages the files for the next commit.

See what's staged, or what commits are outstanding.

### Commit an entry

```bash
til commit -m "Learned about Go interfaces today"
```

### Status

```bash
til status
```

### Amend a commit

```bash
til commit --amend -m "Learned about Go interfaces and embedding today"
```

### Push to Notion

```bash
til push
```

### View the log

```bash
til log
til log -n 5  # Show only the last 5 entries
```

## License

MIT
