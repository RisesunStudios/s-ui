# Git Synchronization

## Overview

Git synchronization feature allows automatic backup and restore of SingBox configuration and database to Git repositories (GitHub, GitLab, Gitea).

## Supported Providers

- **GitHub** - https://github.com
- **GitLab** - https://gitlab.com or self-hosted
- **Gitea** - self-hosted

## Setup

### 1. Create Access Token

#### GitHub
1. Go to Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Create new token with `repo` scope (Full control of private repositories)
3. Copy the token

#### GitLab
1. Go to Settings → Access Tokens
2. Create token with `api`, `read_repository`, `write_repository` scopes
3. Copy the token

#### Gitea
1. Go to Settings → Applications → Generate New Token
2. Select `repo` scope (Full control of repositories)
3. Copy the token

### 2. Create Repository

Create a new private repository for storing configurations and backups.

### 3. Configure in S-UI

#### Create API Token (recommended)

1. Login to S-UI web interface
2. Go to Settings → API Tokens
3. Create a new token
4. Copy the token

#### API Endpoints

Two authentication options available:
- **`/apiv2/`** - uses API tokens (recommended for automation)
- **`/app/api/`** - uses session cookies (for web interface)

**Get configuration:**
```
GET /apiv2/gitSyncConfig
Header: Token: your_api_token

or

GET /app/api/gitSyncConfig
Cookie: session=...
```

**Save configuration:**
```
POST /apiv2/gitSyncConfig
Header: Token: your_api_token
Content-Type: application/json

{
  "enable": true,
  "provider": "github",
  "repoUrl": "https://github.com/username/repo",
  "branch": "main",
  "token": "your_token_here",
  "autoSync": true,
  "syncInterval": 3600,
  "syncConfig": true,
  "syncDb": true
}
```

**Parameters:**
- `enable` - enable/disable synchronization
- `provider` - provider: `github`, `gitlab`, `gitea`
- `repoUrl` - repository URL
- `branch` - branch name (default `main`)
- `token` - access token
- `autoSync` - automatic synchronization
- `syncInterval` - sync interval in seconds (default 3600 = 1 hour)
- `syncConfig` - sync SingBox configuration
- `syncDb` - sync database

**Push to Git:**
```
POST /apiv2/gitSyncPush
Header: Token: your_api_token
```

**Pull from Git:**
```
POST /apiv2/gitSyncPull
Header: Token: your_api_token
```

**Test connection:**
```
POST /apiv2/gitSyncTest
Header: Token: your_api_token
```

## Usage

### Manual Synchronization

1. Create an API token in S-UI (Settings → API Tokens)
2. Configure Git sync settings via API
3. Call `/apiv2/gitSyncPush` to push data to Git
4. Call `/apiv2/gitSyncPull` to pull data from Git

### Automatic Synchronization

When `autoSync` is enabled, the system works as follows:

1. **Push on changes** - automatically pushes configuration to Git on any database changes
2. **Periodic push** - additionally pushes data every hour (for DB sync)
3. **Check for updates** - checks Git version every 30 seconds
4. **Automatic pull** - if Git version is newer than local, automatically downloads and applies configuration

This ensures configuration synchronization between multiple S-UI servers using the same repository.

## Repository Files

After synchronization, the following files will appear in a folder named after your hostname:

```
<hostname>/
├── singbox-config.json  - SingBox configuration in raw format
├── s-ui-backup.db       - database backup
└── version.txt          - configuration version (timestamp)
```

Where `<hostname>` is the hostname from S-UI settings (Settings → Hostname), or "default" if not set.

## Restore

### Restore SingBox Configuration

Configuration is automatically applied when pulling from Git.

### Restore Database

1. Get `s-ui-backup.db` file from repository
2. Use existing API endpoint for import:
```
POST /app/api/importdb
Content-Type: multipart/form-data

db: <database file>
```

## Security

⚠️ **Important:**
- Use **private** repositories
- Keep tokens secure (both S-UI API tokens and Git tokens)
- Don't publish tokens publicly
- Regularly rotate access tokens
- Use tokens with minimal required permissions
- Set expiration dates for API tokens
- Use API tokens instead of session cookies for automation

## Repository URL Examples

**GitHub:**
```
https://github.com/username/repo
https://github.com/username/repo.git
```

**GitLab:**
```
https://gitlab.com/username/repo
https://gitlab.example.com/username/repo
```

**Gitea:**
```
https://gitea.example.com/username/repo
```

## Troubleshooting

### Authentication Error
- Verify token is correct
- Ensure token has required permissions
- Check token expiration

### Repository Access Error
- Verify repository exists
- Check URL is correct
- Ensure token has access to repository

### Files Not Appearing in Repository
- Check branch name is correct
- Verify synchronization is enabled
- Check application logs

## Logging

All sync operations are logged. Check logs for troubleshooting:
```
GET /app/api/logs?c=100&l=debug
```
