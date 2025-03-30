package til

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitManager handles Git operations
type GitManager struct {
	WorkDir string
}

// NewGitManager creates a new GitManager
func NewGitManager(workDir string) *GitManager {
	return &GitManager{
		WorkDir: workDir,
	}
}

// IsInitialized checks if the Git repository is initialized
func (gm *GitManager) IsInitialized() bool {
	gitDir := filepath.Join(gm.WorkDir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Init initializes a local copy of the user's Git repository
func (gm *GitManager) Init(url string) error {
	if gm.IsInitialized() {
		return errors.New("git repository already initialized")
	}

	// 1. Initialize a git repository in the directory
	initCmd := exec.Command("git", "init")
	initCmd.Dir = gm.WorkDir
	if err := initCmd.Run(); err != nil {
		return err
	}

	// 2. Add the remote
	remoteCmd := exec.Command("git", "remote", "add", "origin", url)
	remoteCmd.Dir = gm.WorkDir
	if err := remoteCmd.Run(); err != nil {
		return err
	}

	// 3. Fetch the remote repository
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = gm.WorkDir
	if err := fetchCmd.Run(); err != nil {
		return err
	}

	// 4. Create a new branch tracking the remote
	branchCmd := exec.Command("git", "checkout", "-B", "main", "origin/main", "--force")
	branchCmd.Dir = gm.WorkDir
	err := branchCmd.Run()
	if err != nil {
		// Try with master branch if main fails
		branchCmd = exec.Command("git", "checkout", "-B", "master", "origin/master", "--force")
		branchCmd.Dir = gm.WorkDir
		if err := branchCmd.Run(); err != nil {
			return err
		}
	}

	// 5. Clean up any untracked files that weren't overwritten
	// You can optionally add this if you want to keep only what's in the repo
	// cleanCmd := exec.Command("git", "clean", "-fd")
	// cleanCmd.Dir = gm.WorkDir
	// return cleanCmd.Run()

	return nil
}

func (gm *GitManager) Status() (string, error) {
	if !gm.IsInitialized() {
		return "", errors.New("git repository not initialized")
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = gm.WorkDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// SetRemote sets the remote origin URL
func (gm *GitManager) SetRemote(url string) error {
	if !gm.IsInitialized() {
		return errors.New("git repository not initialized")
	}

	// Check if remote already exists
	checkCmd := exec.Command("git", "remote")
	checkCmd.Dir = gm.WorkDir
	output, err := checkCmd.Output()
	if err != nil {
		return err
	}

	if len(output) > 0 {
		// Remote exists, update it
		cmd := exec.Command("git", "remote", "set-url", "origin", url)
		cmd.Dir = gm.WorkDir
		return cmd.Run()
	} else {
		// Remote doesn't exist, add it
		cmd := exec.Command("git", "remote", "add", "origin", url)
		cmd.Dir = gm.WorkDir
		return cmd.Run()
	}
}

// Add stages files for commit
func (gm *GitManager) Add(files ...string) error {
	if !gm.IsInitialized() {
		return errors.New("git repository not initialized")
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = gm.WorkDir
	return cmd.Run()
}

// AddAll stages all files for commit
func (gm *GitManager) AddAll() error {
	return gm.Add(".")
}

// Commit commits staged files
func (gm *GitManager) Commit(message string) error {
	if !gm.IsInitialized() {
		fmt.Println("Git repository not initialized")
		return errors.New("git repository not initialized")
	}

	cmd := exec.Command("git", "commit", "-m", message)
	fmt.Println("Running commit command:", cmd.String())
	cmd.Dir = gm.WorkDir
	fmt.Println("Commit command directory:", cmd.Dir)
	return cmd.Run()
}

// Push pushes commits to the remote repository
func (gm *GitManager) Push() error {
	if !gm.IsInitialized() {
		return errors.New("git repository not initialized")
	}

	cmd := exec.Command("git", "push", "origin", "master")
	cmd.Dir = gm.WorkDir
	err := cmd.Run()
	if err != nil {
		// Try pushing to main instead
		cmd = exec.Command("git", "push", "origin", "main")
		cmd.Dir = gm.WorkDir
		return cmd.Run()
	}
	return nil
}

// GetFileURL returns the URL of a file in the Git repository
func (gm *GitManager) GetFileURL(remoteURL, filePath string) string {
	// Ensure no trailing .git
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Convert SSH URL to HTTPS URL if needed
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL[4:], ":")
		if len(parts) == 2 {
			remoteURL = fmt.Sprintf("https://%s/%s", parts[0], parts[1])
		}
	}

	// Get the relative path within the repository
	relPath, err := filepath.Rel(gm.WorkDir, filePath)
	if err != nil {
		return ""
	}

	// Construct the URL
	return fmt.Sprintf("%s/blob/main/%s", remoteURL, relPath)
}
