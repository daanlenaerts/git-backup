package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
	Private  bool   `json:"private"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login string `json:"login"`
}

/**
 * Get all github repositories for a given token, including private, public and archived repositories from both the user and all organizations the user is a member of.
 * @param token: string - the token to use for the github API
 * @return: []string - a list of all the repositories
 * @return: error - an error if the request fails
 */
func GetAllGithubRepos(token string) ([]string, error) {
	repos := []string{}

	// Create HTTP client with authentication
	client := &http.Client{}

	// Get current user info
	user, err := getCurrentUser(client, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %v", err)
	}

	// Get user's repositories
	userRepos, err := getUserRepos(client, token, user.Login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %v", err)
	}
	repos = append(repos, userRepos...)

	// Get organizations
	orgs, err := getUserOrgs(client, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %v", err)
	}

	// Get repositories for each organization
	for _, org := range orgs {
		orgRepos, err := getOrgRepos(client, token, org)
		if err != nil {
			// Log error but continue with other orgs
			fmt.Printf("Warning: failed to get repositories for org %s: %v\n", org, err)
			continue
		}
		repos = append(repos, orgRepos...)
	}

	return repos, nil
}

// getCurrentUser fetches the current authenticated user
func getCurrentUser(client *http.Client, token string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user GitHubUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// getUserRepos fetches all repositories for a user
func getUserRepos(client *http.Client, token, username string) ([]string, error) {
	return getReposFromURL(client, token, fmt.Sprintf("https://api.github.com/users/%s/repos?type=all&per_page=100", username))
}

// getUserOrgs fetches all organizations the user is a member of
func getUserOrgs(client *http.Client, token string) ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/orgs?per_page=100", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orgs []struct {
		Login string `json:"login"`
	}
	err = json.Unmarshal(body, &orgs)
	if err != nil {
		return nil, err
	}

	orgNames := make([]string, len(orgs))
	for i, org := range orgs {
		orgNames[i] = org.Login
	}

	return orgNames, nil
}

// getOrgRepos fetches all repositories for an organization
func getOrgRepos(client *http.Client, token, orgName string) ([]string, error) {
	return getReposFromURL(client, token, fmt.Sprintf("https://api.github.com/orgs/%s/repos?type=all&per_page=100", orgName))
}

// getReposFromURL fetches repositories from a given URL with pagination support
func getReposFromURL(client *http.Client, token, url string) ([]string, error) {
	var allRepos []string
	page := 1

	for {
		pageURL := fmt.Sprintf("%s&page=%d", url, page)
		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		err = json.Unmarshal(body, &repos)
		if err != nil {
			return nil, err
		}

		// If no repositories returned, we've reached the end
		if len(repos) == 0 {
			break
		}

		// Add repository URLs to our list
		for _, repo := range repos {
			// Always use HTTPS URLs for consistency and to avoid SSH setup issues
			allRepos = append(allRepos, repo.CloneURL)
		}

		// If we got fewer repos than the page size, we're done
		if len(repos) < 100 {
			break
		}

		page++
	}

	return allRepos, nil
}
