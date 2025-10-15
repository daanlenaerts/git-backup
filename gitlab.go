package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GitLabRepo represents a GitLab repository
type GitLabRepo struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	Visibility        string `json:"visibility"`
	Archived          bool   `json:"archived"`
}

// GitLabUser represents a GitLab user
type GitLabUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// GitLabGroup represents a GitLab group
type GitLabGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

/**
 * Get all GitLab repositories for a given token, including private, public and archived repositories from both the user and all groups the user is a member of.
 * @param token: string - the token to use for the GitLab API
 * @return: []string - a list of all the repositories
 * @return: error - an error if the request fails
 */
func GetAllGitLabRepos(token string) ([]string, error) {
	repos := []string{}

	// Create HTTP client with authentication
	client := &http.Client{}

	// Get current user info
	user, err := getCurrentGitLabUser(client, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %v", err)
	}

	// Get user's repositories
	userRepos, err := getUserGitLabRepos(client, token, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %v", err)
	}
	repos = append(repos, userRepos...)

	// Get groups
	groups, err := getUserGitLabGroups(client, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %v", err)
	}

	// Get repositories for each group
	for _, group := range groups {
		groupRepos, err := getGroupGitLabRepos(client, token, group.ID)
		if err != nil {
			// Log error but continue with other groups
			fmt.Printf("Warning: failed to get repositories for group %s: %v\n", group.Name, err)
			continue
		}
		repos = append(repos, groupRepos...)
	}

	return repos, nil
}

// getCurrentGitLabUser fetches the current authenticated user
func getCurrentGitLabUser(client *http.Client, token string) (*GitLabUser, error) {
	req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user GitLabUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// getUserGitLabRepos fetches all repositories for a user
func getUserGitLabRepos(client *http.Client, token string, userID int) ([]string, error) {
	return getGitLabReposFromURL(client, token, fmt.Sprintf("https://gitlab.com/api/v4/users/%d/projects?membership=true&per_page=100", userID))
}

// getUserGitLabGroups fetches all groups the user is a member of
func getUserGitLabGroups(client *http.Client, token string) ([]GitLabGroup, error) {
	req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/groups?membership=true&per_page=100", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var groups []GitLabGroup
	err = json.Unmarshal(body, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// getGroupGitLabRepos fetches all repositories for a group
func getGroupGitLabRepos(client *http.Client, token string, groupID int) ([]string, error) {
	return getGitLabReposFromURL(client, token, fmt.Sprintf("https://gitlab.com/api/v4/groups/%d/projects?per_page=100", groupID))
}

// getGitLabReposFromURL fetches repositories from a given URL with pagination support
func getGitLabReposFromURL(client *http.Client, token, url string) ([]string, error) {
	var allRepos []string
	page := 1

	for {
		pageURL := fmt.Sprintf("%s&page=%d", url, page)
		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var repos []GitLabRepo
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
			allRepos = append(allRepos, repo.HTTPURLToRepo)
		}

		// If we got fewer repos than the page size, we're done
		if len(repos) < 100 {
			break
		}

		page++
	}

	return allRepos, nil
}
