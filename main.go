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

	SendTelegramMessage("Starting git-backup")

	for {

		backupAll("GitHub", githubTokens, GetAllGithubRepos)
		backupAll("GitLab", gitlabTokens, GetAllGitLabRepos)

		if repeatInt == -1 {
			break
		}
		time.Sleep(time.Duration(repeatInt) * time.Minute)
	}

}

// backupAll lists the repositories visible to each token, then backs up each
// repository once, trying every token that can see it until one succeeds.
func backupAll(provider string, tokens []string, listRepos func(string) ([]string, error)) {
	var repos []string
	repoTokens := map[string][]string{}

	for _, token := range tokens {
		tokenRepos, err := listRepos(token)
		if err != nil {
			fmt.Printf("Error getting %s repositories: %v\n", provider, err)
			SendTelegramMessage(fmt.Sprintf("Error getting %s repositories: %v", provider, err))
			continue
		}
		for _, repo := range tokenRepos {
			if _, seen := repoTokens[repo]; !seen {
				repos = append(repos, repo)
			}
			repoTokens[repo] = append(repoTokens[repo], token)
		}
	}

	fmt.Printf("%s repositories: %v\n", provider, repos)

	for _, repo := range repos {
		var err error
		for _, token := range repoTokens[repo] {
			if err = BackupRepo(repo, token); err == nil {
				break
			}
		}
		if err != nil {
			fmt.Println("Error cloning repository:", err)
			SendTelegramMessage("Error cloning repository: " + err.Error())
		}
	}
}
