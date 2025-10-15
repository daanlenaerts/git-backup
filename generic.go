package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func BackupRepo(repo string, token string) error {
	var repoName string

	repoName = strings.TrimPrefix(strings.TrimPrefix(repo, "https://"), "git@")

	// Remove .git extension if present
	if filepath.Ext(repoName) == ".git" {
		repoName = repoName[:len(repoName)-4]
	}

	// Replace all non-alphanumeric characters in the repo name with underscores with regex
	repoName = strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9\-]`).ReplaceAllString(repoName, "_"))

	// Create the command to clone the repository
	// If the target directory does not exist, clone, otherwise fetch --all
	var cmd *exec.Cmd
	targetDir := os.Getenv("REPOS_DIR")
	if targetDir == "" {
		targetDir = "./repos"
	}
	repoPath := filepath.Join(targetDir, repoName)

	// Prepare authenticated URL if token is provided
	authRepo := repo
	if token != "" && strings.HasPrefix(repo, "https://") {
		// Insert token into HTTPS URL for authentication
		// Format: https://token@github.com/user/repo.git
		parts := strings.Split(repo, "://")
		if len(parts) == 2 {
			authRepo = fmt.Sprintf("%s://%s@%s", parts[0], token, parts[1])
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		fmt.Printf("Cloning repository %s to %s\n", repo, repoPath)
		// Clone the repository
		cmd = exec.Command("git", "clone", authRepo, repoPath)
	} else {
		fmt.Printf("Fetching updates from repository %s to %s\n", repo, repoPath)
		// Pull updates from the existing repository
		cmd = exec.Command("git", "fetch", "--all")
		cmd.Dir = repoPath
	}

	// Run the command and capture output and errors
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %v\nGit output: %s", repo, err, string(output))
	}

	fmt.Printf("Successfully updated %s in %s\n", repo, repoName)
	return nil
}
