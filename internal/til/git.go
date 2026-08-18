package til

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNoChanges = errors.New("no staged changes")

type GitManager struct {
	WorkDir string
}

func NewGitManager(workDir string) *GitManager {
	return &GitManager{WorkDir: workDir}
}

func (gm *GitManager) IsInitialized() bool {
	info, err := os.Stat(filepath.Join(gm.WorkDir, ".git"))
	return err == nil && info.IsDir()
}

func (gm *GitManager) Init(remoteURL string) (retErr error) {
	if gm.IsInitialized() {
		return errors.New("git repository already initialized")
	}
	if strings.TrimSpace(remoteURL) == "" {
		return errors.New("git remote URL cannot be empty")
	}
	if err := os.MkdirAll(gm.WorkDir, 0755); err != nil {
		return fmt.Errorf("create Git working directory: %w", err)
	}

	gitDir := filepath.Join(gm.WorkDir, ".git")
	defer func() {
		if retErr != nil {
			_ = os.RemoveAll(gitDir)
		}
	}()

	if _, err := gm.run("init", "-b", "main"); err != nil {
		if _, fallbackErr := gm.run("init"); fallbackErr != nil {
			return fallbackErr
		}
		if _, fallbackErr := gm.run("symbolic-ref", "HEAD", "refs/heads/main"); fallbackErr != nil {
			return fallbackErr
		}
	}
	if err := gm.SetRemote(remoteURL); err != nil {
		return err
	}
	if _, err := gm.run("fetch", "origin"); err != nil {
		return fmt.Errorf("fetch Git remote: %w", err)
	}

	branch, found, err := gm.remoteBranch()
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	if _, err := gm.run("checkout", "-B", branch, "origin/"+branch); err != nil {
		return fmt.Errorf("check out remote branch %s: %w", branch, err)
	}
	if _, err := gm.run("branch", "--set-upstream-to", "origin/"+branch, branch); err != nil {
		return fmt.Errorf("track remote branch %s: %w", branch, err)
	}
	return nil
}

func (gm *GitManager) Status() (string, error) {
	if !gm.IsInitialized() {
		return "", errors.New("git repository not initialized")
	}
	output, err := gm.run("status", "--porcelain")
	if err != nil {
		return "", err
	}
	return output, nil
}

func (gm *GitManager) SetRemote(remoteURL string) error {
	if !gm.IsInitialized() {
		return errors.New("git repository not initialized")
	}
	if strings.TrimSpace(remoteURL) == "" {
		return errors.New("git remote URL cannot be empty")
	}

	if _, err := gm.run("remote", "get-url", "origin"); err == nil {
		_, err = gm.run("remote", "set-url", "origin", remoteURL)
		return err
	}
	_, err := gm.run("remote", "add", "origin", remoteURL)
	return err
}

func (gm *GitManager) Add(files ...string) error {
	if !gm.IsInitialized() {
		return errors.New("git repository not initialized")
	}
	if len(files) == 0 {
		return errors.New("at least one file is required")
	}

	args := append([]string{"add", "--"}, files...)
	_, err := gm.run(args...)
	return err
}

func (gm *GitManager) AddAll() error {
	return gm.Add(".")
}

func (gm *GitManager) HasStagedChanges() (bool, error) {
	if !gm.IsInitialized() {
		return false, errors.New("git repository not initialized")
	}

	cmd := exec.Command("git", "diff", "--cached", "--quiet", "--exit-code")
	cmd.Dir = gm.WorkDir
	err := cmd.Run()
	if err == nil {
		return false, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("inspect staged Git changes: %w", err)
}

func (gm *GitManager) Commit(message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("Git commit message cannot be empty")
	}
	hasChanges, err := gm.HasStagedChanges()
	if err != nil {
		return err
	}
	if !hasChanges {
		return ErrNoChanges
	}

	_, err = gm.run("commit", "-m", message)
	return err
}

func (gm *GitManager) Push() error {
	if !gm.IsInitialized() {
		return errors.New("git repository not initialized")
	}
	if _, err := gm.run("remote", "get-url", "origin"); err != nil {
		return errors.New("Git remote 'origin' is not configured")
	}

	branch, err := gm.CurrentBranch()
	if err != nil {
		return err
	}
	if _, err := gm.run("push", "--set-upstream", "origin", branch); err != nil {
		return fmt.Errorf("push branch %s: %w", branch, err)
	}
	return nil
}

func (gm *GitManager) CurrentBranch() (string, error) {
	if !gm.IsInitialized() {
		return "", errors.New("git repository not initialized")
	}
	branch, err := gm.run("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("determine current Git branch: %w", err)
	}
	if branch == "" {
		return "", errors.New("Git repository is not on a branch")
	}
	return branch, nil
}

func (gm *GitManager) GetFileURL(remoteURL, filePath string) string {
	webURL, err := remoteWebURL(remoteURL)
	if err != nil {
		return ""
	}
	relativePath, err := filepath.Rel(gm.WorkDir, filePath)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return ""
	}
	branch, err := gm.CurrentBranch()
	if err != nil {
		branch = "main"
	}
	return fmt.Sprintf(
		"%s/blob/%s/%s",
		webURL,
		url.PathEscape(branch),
		escapeURLPath(filepath.ToSlash(relativePath)),
	)
}

func GitHubRawFileURL(remoteURL, branch, filePath string) (string, error) {
	webURL, err := remoteWebURL(remoteURL)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(webURL)
	if err != nil {
		return "", fmt.Errorf("parse Git remote URL: %w", err)
	}
	if !strings.EqualFold(parsed.Hostname(), "github.com") {
		return "", fmt.Errorf("attachments require a GitHub remote, got %s", parsed.Hostname())
	}
	if strings.TrimSpace(branch) == "" {
		branch = "main"
	}

	repositoryPath := strings.Trim(parsed.Path, "/")
	if repositoryPath == "" {
		return "", errors.New("GitHub remote does not include a repository path")
	}
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s",
		repositoryPath,
		url.PathEscape(branch),
		escapeURLPath(filepath.ToSlash(filePath)),
	), nil
}

func RedactGitRemoteURL(remoteURL string) string {
	parsed, err := url.Parse(remoteURL)
	if err != nil || parsed.User == nil {
		return remoteURL
	}
	parsed.User = nil
	return parsed.String()
}

func (gm *GitManager) remoteBranch() (string, bool, error) {
	if branch, err := gm.run("symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil {
		return strings.TrimPrefix(branch, "origin/"), true, nil
	}

	for _, branch := range []string{"main", "master"} {
		if _, err := gm.run("show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch); err == nil {
			return branch, true, nil
		}
	}

	output, err := gm.run("for-each-ref", "--format=%(refname:short)", "refs/remotes/origin")
	if err != nil {
		return "", false, err
	}
	for _, ref := range strings.Split(output, "\n") {
		ref = strings.TrimSpace(ref)
		if ref != "" && ref != "origin/HEAD" {
			return strings.TrimPrefix(ref, "origin/"), true, nil
		}
	}
	return "", false, nil
}

func (gm *GitManager) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = gm.WorkDir
	output, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(output))
	if err != nil {
		displayArgs := make([]string, len(args))
		for i, arg := range args {
			displayArgs[i] = RedactGitRemoteURL(arg)
			if displayArgs[i] != arg {
				trimmed = strings.ReplaceAll(trimmed, arg, displayArgs[i])
			}
		}
		if trimmed == "" {
			return "", fmt.Errorf("git %s: %w", strings.Join(displayArgs, " "), err)
		}
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(displayArgs, " "), trimmed, err)
	}
	return trimmed, nil
}

func remoteWebURL(remoteURL string) (string, error) {
	remoteURL = strings.TrimSpace(strings.TrimSuffix(remoteURL, ".git"))
	if remoteURL == "" {
		return "", errors.New("Git remote URL cannot be empty")
	}

	if strings.HasPrefix(remoteURL, "git@") {
		hostAndPath := strings.TrimPrefix(remoteURL, "git@")
		host, path, ok := strings.Cut(hostAndPath, ":")
		if !ok || host == "" || path == "" {
			return "", fmt.Errorf("invalid SSH Git remote: %s", remoteURL)
		}
		return fmt.Sprintf("https://%s/%s", host, strings.TrimPrefix(path, "/")), nil
	}

	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return "", fmt.Errorf("parse Git remote URL: %w", err)
	}
	switch parsed.Scheme {
	case "http", "https":
		parsed.User = nil
		return strings.TrimSuffix(parsed.String(), "/"), nil
	case "ssh":
		if parsed.Hostname() == "" || parsed.Path == "" {
			return "", fmt.Errorf("invalid SSH Git remote: %s", remoteURL)
		}
		return fmt.Sprintf("https://%s/%s", parsed.Hostname(), strings.TrimPrefix(parsed.Path, "/")), nil
	default:
		return "", fmt.Errorf("Git remote does not have a web URL: %s", remoteURL)
	}
}

func escapeURLPath(path string) string {
	parts := strings.Split(path, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
