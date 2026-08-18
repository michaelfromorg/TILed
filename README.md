# Today I Learned

`til` is a command-line application for keeping a lightweight, append-first log of what you learn. It uses a familiar Git-like workflow: stage optional attachments, commit an entry, inspect the log, and explicitly push to configured destinations.

## Features

- Commit any number of entries per day
- Add an optional Markdown body through `$TIL_EDITOR`, `$EDITOR`, or `$VISUAL`
- Attach files up to 10 MiB
- Amend the latest entry before publishing
- Filter the log by date and search titles, bodies, IDs, or attachments
- Store entries transactionally in an embedded SQLite database
- Export the complete log as Markdown or JSON
- Create verified database backups and run integrity checks
- Create checksummed portable archives and restore them on a new device
- Generate completions for Bash, Zsh, Fish, and PowerShell
- Sync the generated log and attachments to Git
- Publish entries to a Notion database
- Run commands from the repository root or any subdirectory
- Automatically migrate repositories that use legacy YAML or Markdown storage

## Requirements

- Go 1.22 or newer to build from source
- Git when Git synchronization is enabled
- Optional Notion integration with a database containing:
  - `TIL`, a title property
  - `Attachments`, a files property

## Installation

Install the `til` binary with Go:

```bash
go install github.com/michaelfromorg/tiled/cmd/til@latest
```

Or build it from source:

```bash
git clone https://github.com/michaelfromorg/tiled.git
cd tiled
make build
```

Generate a completion script for your current shell:

```text
source <(til completion bash)
source <(til completion zsh)
til completion fish | source
til completion powershell | Out-String | Invoke-Expression
```

Run `til completion <shell> --help` for the supported shell names. Redirect the generated script into your shell's completion directory to install it persistently.

## Quick start

Initialize a repository in the directory where you want to keep its local configuration:

```bash
til init
```

`init` asks whether to configure Notion and Git. Both are optional, so you can use `til` entirely locally.

Create an entry without an attachment:

```bash
til commit -m "Learned how Go interfaces compose"
```

Or stage one or more files first:

```bash
til add example.go diagram.png
til commit -m "Explored interface embedding"
```

When `-m` is omitted, `til` opens your configured editor. The first line becomes the entry title and the remaining text becomes its body.

```bash
TIL_EDITOR="code --wait" til commit
```

Inspect local state and recent entries:

```bash
til status
til log
til log -n 5
til log --date 2025-03-30
til log --since 2025-03-01 --until 2025-03-31 --all
til log --long
til log --json
til slog "interface embedding"
```

`til log` and `til slog` show the newest 10 matches by default. Use `--all` for every match, `--reverse` for oldest-first ordering, and `--long` to include entry bodies. `til slog` performs a case-insensitive literal search over commit IDs, titles, bodies, and attachment names. Date flags use `YYYY-MM-DD`; `--until` is inclusive.

Amend the latest entry:

```bash
til add revised-example.go
til commit --amend -m "Explored interface embedding and type sets"
```

Synchronize committed entries:

```bash
til push            # all configured destinations
til push --git      # Git only
til push --notion   # Notion only
```

`commit` and `--amend` only update local TIL data. `push` is the operation that creates and pushes a Git commit or publishes entries to Notion.

## Exporting entries

Export every entry in chronological order:

```bash
til export
til export --format json
til export --format markdown --output til-export.md
til export --format json --output til-export.json
```

Exports go to standard output unless `--output` is provided. Existing output files are protected by default; pass `--force` to replace a regular file. Markdown exports preserve entry bodies, and both formats include timestamps, commit IDs, attachment names, and Notion synchronization state. Exporting references attachments by name but does not copy the attachment files.

## Storage

A repository has this layout:

```text
project/
├── .til/
│   ├── backups/
│   ├── config
│   ├── restore-backups/
│   └── staging/
└── til/
    ├── README.md
    ├── til.db
    └── files/
```

- `.til/config` contains local sync settings and is written with owner-only permissions on Unix systems. It can contain your Notion token and must not be committed.
- `.til/backups` contains automatic migration backups, SQLite snapshots, and portable archives.
- `.til/restore-backups` preserves the previous database, files, and README after a forced restore.
- `.til/staging` contains attachment copies waiting for the next commit.
- `til/til.db` is the canonical SQLite entry log.
- `til/README.md` is regenerated before Git pushes.
- `til/files` stores bodies and attachments using each entry's commit ID, so multiple entries on the same day cannot overwrite one another.

## Database maintenance

Check both SQLite page integrity and all foreign-key relationships:

```bash
til db check
til db integrity  # alias for db check
```

Create a consistent, verified SQLite snapshot:

```bash
til db backup
til db backup /path/to/til-backup.db
```

When no destination is supplied, backups are written to `.til/backups` with a timestamped name. Existing files are never overwritten. On Unix systems, backup files use owner-only permissions. A database backup contains all entry metadata and bodies stored in SQLite; copy `til/files` separately when you also need an independent backup of attachment contents.

## Portable archive and restore

Create a complete portable archive containing a consistent `til.db` snapshot and every regular file under `til/files`:

```bash
til archive
til archive /path/to/my-til.tar.gz
```

When no destination is provided, the archive is written to `.til/backups`. Each archive contains a versioned manifest with the size, permissions, and SHA-256 checksum of every payload file. Local `.til/config`, Git metadata, staging files, and secrets are intentionally excluded.

To move your TIL repository to a new device:

```bash
# On the old device
til archive ~/my-til.tar.gz

# Copy my-til.tar.gz to the new device, then:
mkdir -p ~/my-til
cd ~/my-til
til restore /path/to/my-til.tar.gz
```

Restore validates archive paths and checksums, verifies SQLite and foreign-key integrity, confirms that referenced bodies and attachments exist, installs the data, regenerates `til/README.md`, and creates a local-only `.til/config` when restoring into a fresh directory. Sync credentials remain device-specific and are never stored in the archive.

Restoring over existing data is rejected by default. `til restore --force <archive>` first moves the current `til.db`, `til/files`, and generated README under `.til/restore-backups`, then installs the validated archive. An existing nested `til/.git` directory is left in place.

## Git synchronization

`til init` accepts either an empty Git remote or an existing repository. It detects the remote branch rather than assuming `main` or `master`.

When you run `til push`, the application:

1. Regenerates `til/README.md`
2. Stages changes in the nested `til` Git repository
3. Creates a Git commit when files changed
4. Pushes the current branch to `origin`

Git author configuration is still handled by Git. If your name or email is not configured, Git returns an actionable error during `til push`.

## Notion synchronization

Notion pages use the `TIL` property for the entry title and page blocks for the optional body. Local `notion_synced` state prevents already-published entries from being sent again; `til push --notion --force` deliberately republishes them.

Notion's files property requires public URLs rather than local file uploads with the API version used here. Therefore, entries with attachments require a configured GitHub remote. A normal `til push` sends Git changes first and then publishes GitHub raw-file URLs to Notion. If an attachment cannot be published safely, `til` reports an error instead of inserting a hard-coded or broken URL.

## Migrating legacy repositories

The first command run against a configured repository containing legacy `til/til.yml` or `til/til.md` storage migrates it automatically. You can also start the migration explicitly:

```bash
til migrate
```

The migration creates `til.db`, preserves entry and Notion synchronization state, assigns safe commit IDs where needed, and moves legacy body and attachment files to commit-ID-based names. Before changing the repository, it copies the original storage file to `.til/backups/til.yml.bak` or `.til/backups/til.md.bak`. Existing backups are retained with numeric suffixes.

## Development

Run the complete local verification suite:

```bash
make check
```

This checks formatting, runs `go vet`, executes unit and offline end-to-end tests, runs the race detector, and builds the release binary. The Git end-to-end test uses a temporary local bare repository and does not require network access or credentials.

## License

MIT
