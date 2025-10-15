# Git Backup

A Go application that backs up GitHub and GitLab repositories.

## Docker Usage

### Build the Docker image

```bash
docker build -t git-backup .
```

### Run with Docker

```bash
# Set your tokens as environment variables
export GITHUB_TOKENS="your_github_token1,your_github_token2"
export GITLAB_TOKENS="your_gitlab_token1,your_gitlab_token2"
export REPEAT_INTERVAL=60 # every 60 minutes

# Run once
docker run --rm \
  -e GITHUB_TOKENS="$GITHUB_TOKENS" \
  -e GITLAB_TOKENS="$GITLAB_TOKENS" \
  -v $(pwd)/backups:/app/repos \
  -e REPEAT_INTERVAL="$REPEAT_INTERVAL" \
  git-backup

# Run with repeat mode (every 60 minutes)
docker run -d \
  -e GITHUB_TOKENS="$GITHUB_TOKENS" \
  -e GITLAB_TOKENS="$GITLAB_TOKENS" \
  -v $(pwd)/backups:/app/repos \
  -e REPEAT_INTERVAL="$REPEAT_INTERVAL" \
  --name git-backup \
  git-backup ./git-backup --repeat 60
```

## Environment Variables

- `GITHUB_TOKENS`: Comma-separated list of GitHub personal access tokens
- `GITLAB_TOKENS`: Comma-separated list of GitLab personal access tokens
- `REPEAT_INTERVAL`: Run the backup every N minutes (default: run once)

## Notes

- Backed up repositories are stored in the `./repos` directory (or `/repos` when using Docker)
- The application will clone new repositories and fetch updates for existing ones
- Repository names are sanitized to be filesystem-safe

## Building Docker Image and Pushing to Docker Hub

```bash
docker build -t git-backup .
docker tag git-backup docker.io/lenaertsdaan/git-backup:latest
docker push docker.io/lenaertsdaan/git-backup:latest
```