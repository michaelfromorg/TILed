# TIL - Today I Learned

TIL (Today I Learned) is a command-line application for tracking what you learned today. It provides a git-like interface for adding entries and syncing them with Notion.

## Features

- `til init` - Initialize a new TIL repository or sync with an existing one
- `til add <files>` - Stage files for the current item
- `til commit -m "<message>"` - Add a new item to the log (message is required, files are optional)
- `til commit --amend` - Amend the previous commit
- `til push` - Submit all outstanding commits to Notion
- `til log -n <number>` - View a one-line log of past learnings (with filenames)

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

### Initialize a Repository

```bash
til init
```

This will prompt you to configure Notion sync. If you enable Notion sync, you'll need to provide your Notion API key and database ID.

### Add Files

```bash
til add file1.txt file2.txt
```

This stages the files for the next commit.

### Commit an Entry

```bash
til commit -m "Learned about Go interfaces today"
```

### Amend a Commit

```bash
til commit --amend -m "Learned about Go interfaces and embedding today"
```

### Push to Notion

```bash
til push
```

### View the Log

```bash
til log
til log -n 5  # Show only the last 5 entries
```

## Notion Integration

TIL can sync your entries with a Notion database. The database should have the following columns:

- TIL (title) - The message of your TIL entry
- Created (date) - Automatically added by Notion
- Attachment (files) - Files attached to the entry

## Data Structure

TIL maintains its data in the following structure:

- `data/til.md` - The log file
- `data/files/<yyyy-mm-dd>_<filename.ext>` - Files attached to entries
- `.til/config` - Configuration file
- `.til/staging/` - Staged files

## License

MIT

