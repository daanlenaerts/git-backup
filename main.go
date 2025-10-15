package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Get GITHUB_TOKENS from environment variable
	githubTokens := strings.Split(os.Getenv("GITHUB_TOKENS"), ",")

	// Get GITLAB_TOKENS from environment variable
	gitlabTokens := strings.Split(os.Getenv("GITLAB_TOKENS"), ",")

	if len(gitlabTokens) == 0 && len(githubTokens) == 0 {
		fmt.Println("GITLAB_TOKENS or GITHUB_TOKENS is not set")
		return
	}

	// Get REPEAT_INTERVAL from environment variable
	repeat := os.Getenv("REPEAT_INTERVAL")
	if repeat == "" {
		repeat = "-1"
	}
	repeatInt, err := strconv.Atoi(repeat)
	if err != nil {
		fmt.Println("Error converting REPEAT_INTERVAL to int:", err)
		return
	}

	for {

		for _, token := range githubTokens {
			err := backupGithub(token)
			if err != nil {
				fmt.Println("Error backing up github:", err)
			}
		}

		for _, token := range gitlabTokens {
			err := backupGitLab(token)
			if err != nil {
				fmt.Println("Error backing up gitlab:", err)
			}
		}

		if repeatInt == -1 {
			break
		}
		time.Sleep(time.Duration(repeatInt) * time.Minute)
	}

}

func backupGithub(token string) error {
	// Get all github repositories
	repos, err := GetAllGithubRepos(token)
	if err != nil {
		fmt.Println("Error getting github repositories:", err)
		return err
	}

	fmt.Println("Github repositories:", repos)

	// Clone the repositories
	for _, repo := range repos {
		err := BackupRepo(repo, token)
		if err != nil {
			fmt.Println("Error cloning repository:", err)
		}
	}

	return nil
}

func backupGitLab(token string) error {
	// Get all gitlab repositories
	repos, err := GetAllGitLabRepos(token)
	if err != nil {
		fmt.Println("Error getting gitlab repositories:", err)
		return err
	}

	fmt.Println("GitLab repositories:", repos)

	// Clone the repositories
	for _, repo := range repos {
		err := BackupRepo(repo, token)
		if err != nil {
			fmt.Println("Error cloning repository:", err)
		}
	}

	return nil
}
