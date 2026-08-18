# End-to-end tests

`e2e_test.go` builds the real `til` binary and exercises realistic workflows in temporary directories.

- `TestLocalCLIWorkflow` covers piped initialization, add, multiple same-day commits, amend, log, status, nested-directory discovery, and no-op push behavior.
- `TestGitCLIWorkflow` uses a temporary local bare Git remote to cover initialization, explicit push semantics, generated README links, and attachment publication.

Run both through the normal Go suite:

```bash
go test ./...
```

The `test.sh` and `test-git.sh` wrappers run either workflow individually. Neither test requires network access, API keys, or an `.env` file.
