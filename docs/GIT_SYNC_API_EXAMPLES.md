# Git Sync API Examples

## Authentication

### Option 1: Using API Token (Recommended for automation)

First, create an API token in S-UI:
1. Login to S-UI web interface
2. Go to Settings → API Tokens
3. Create a new token
4. Copy the token

Then use it with `/apiv2/` endpoints:

```bash
curl -X GET http://localhost:2095/apiv2/gitSyncConfig \
  -H "Token: your_api_token_here"
```

### Option 2: Using Session Cookie (for web interface)

Login first:

```bash
curl -X POST http://localhost:2095/app/api/login \
  -d "user=admin&pass=admin" \
  -c cookies.txt
```

Then use with `/app/api/` endpoints:

```bash
curl -X GET http://localhost:2095/app/api/gitSyncConfig \
  -b cookies.txt
```

## Configuration Examples

### GitHub Configuration

**Using API Token (recommended):**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncConfig \
  -H "Content-Type: application/json" \
  -H "Token: your_api_token_here" \
  -d '{
    "enable": true,
    "provider": "github",
    "repoUrl": "https://github.com/username/s-ui-backup",
    "branch": "main",
    "token": "ghp_xxxxxxxxxxxxxxxxxxxx",
    "autoSync": true,
    "syncInterval": 3600,
    "syncConfig": true,
    "syncDb": true
  }'
```

**Using Session Cookie:**
```bash
curl -X POST http://localhost:2095/app/api/gitSyncConfig \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "enable": true,
    "provider": "github",
    "repoUrl": "https://github.com/username/s-ui-backup",
    "branch": "main",
    "token": "ghp_xxxxxxxxxxxxxxxxxxxx",
    "autoSync": true,
    "syncInterval": 3600,
    "syncConfig": true,
    "syncDb": true
  }'
```

### GitLab Configuration

**Using API Token:**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncConfig \
  -H "Content-Type: application/json" \
  -H "Token: your_api_token_here" \
  -d '{
    "enable": true,
    "provider": "gitlab",
    "repoUrl": "https://gitlab.com/username/s-ui-backup",
    "branch": "main",
    "token": "glpat-xxxxxxxxxxxxxxxxxxxx",
    "autoSync": true,
    "syncInterval": 3600,
    "syncConfig": true,
    "syncDb": true
  }'
```

### Gitea Configuration

**Using API Token:**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncConfig \
  -H "Content-Type: application/json" \
  -H "Token: your_api_token_here" \
  -d '{
    "enable": true,
    "provider": "gitea",
    "repoUrl": "https://gitea.example.com/username/s-ui-backup",
    "branch": "main",
    "token": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "autoSync": true,
    "syncInterval": 3600,
    "syncConfig": true,
    "syncDb": true
  }'
```

## Get Current Configuration

**Using API Token:**
```bash
curl -X GET http://localhost:2095/apiv2/gitSyncConfig \
  -H "Token: your_api_token_here"
```

**Using Session Cookie:**
```bash
curl -X GET http://localhost:2095/app/api/gitSyncConfig \
  -b cookies.txt
```

Response:
```json
{
  "success": true,
  "obj": {
    "id": 1,
    "enable": true,
    "provider": "github",
    "repoUrl": "https://github.com/username/s-ui-backup",
    "branch": "main",
    "token": "***",
    "autoSync": true,
    "syncInterval": 3600,
    "syncConfig": true,
    "syncDb": true,
    "lastSync": 1713643744
  }
}
```

## Manual Push to Git

**Using API Token:**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncPush \
  -H "Token: your_api_token_here"
```

**Using Session Cookie:**
```bash
curl -X POST http://localhost:2095/app/api/gitSyncPush \
  -b cookies.txt
```

Response:
```json
{
  "success": true,
  "msg": ""
}
```

## Manual Pull from Git

**Using API Token:**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncPull \
  -H "Token: your_api_token_here"
```

**Using Session Cookie:**
```bash
curl -X POST http://localhost:2095/app/api/gitSyncPull \
  -b cookies.txt
```

Response:
```json
{
  "success": true,
  "msg": ""
}
```

## Test Connection

**Using API Token:**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncTest \
  -H "Token: your_api_token_here"
```

**Using Session Cookie:**
```bash
curl -X POST http://localhost:2095/app/api/gitSyncTest \
  -b cookies.txt
```

Success Response:
```json
{
  "success": true,
  "msg": ""
}
```

Error Response:
```json
{
  "success": false,
  "msg": "GitHub API error: 401 - Bad credentials"
}
```

## Disable Sync

**Using API Token:**
```bash
curl -X POST http://localhost:2095/apiv2/gitSyncConfig \
  -H "Content-Type: application/json" \
  -H "Token: your_api_token_here" \
  -d '{
    "enable": false,
    "provider": "github",
    "repoUrl": "https://github.com/username/s-ui-backup",
    "branch": "main",
    "token": "ghp_xxxxxxxxxxxxxxxxxxxx",
    "autoSync": false,
    "syncInterval": 3600,
    "syncConfig": true,
    "syncDb": true
  }'
```

## JavaScript/Fetch Examples

### Configure Git Sync

**Using API Token:**
```javascript
fetch('/apiv2/gitSyncConfig', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Token': 'your_api_token_here'
  },
  body: JSON.stringify({
    enable: true,
    provider: 'github',
    repoUrl: 'https://github.com/username/s-ui-backup',
    branch: 'main',
    token: 'ghp_xxxxxxxxxxxxxxxxxxxx',
    autoSync: true,
    syncInterval: 3600,
    syncConfig: true,
    syncDb: true
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

**Using Session Cookie (in browser):**
```javascript
fetch('/app/api/gitSyncConfig', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    enable: true,
    provider: 'github',
    repoUrl: 'https://github.com/username/s-ui-backup',
    branch: 'main',
    token: 'ghp_xxxxxxxxxxxxxxxxxxxx',
    autoSync: true,
    syncInterval: 3600,
    syncConfig: true,
    syncDb: true
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

### Push to Git

**Using API Token:**
```javascript
fetch('/apiv2/gitSyncPush', {
  method: 'POST',
  headers: {
    'Token': 'your_api_token_here'
  }
})
.then(response => response.json())
.then(data => {
  if (data.success) {
    console.log('Successfully pushed to Git');
  } else {
    console.error('Push failed:', data.msg);
  }
});
```

### Test Connection

**Using API Token:**
```javascript
fetch('/apiv2/gitSyncTest', {
  method: 'POST',
  headers: {
    'Token': 'your_api_token_here'
  }
})
.then(response => response.json())
.then(data => {
  if (data.success) {
    console.log('Connection successful');
  } else {
    console.error('Connection failed:', data.msg);
  }
});
```

## Python Examples

### Configure Git Sync

**Using API Token:**
```python
import requests

url = 'http://localhost:2095/apiv2/gitSyncConfig'
headers = {
    'Content-Type': 'application/json',
    'Token': 'your_api_token_here'
}

data = {
    'enable': True,
    'provider': 'github',
    'repoUrl': 'https://github.com/username/s-ui-backup',
    'branch': 'main',
    'token': 'ghp_xxxxxxxxxxxxxxxxxxxx',
    'autoSync': True,
    'syncInterval': 3600,
    'syncConfig': True,
    'syncDb': True
}

response = requests.post(url, json=data, headers=headers)
print(response.json())
```

**Using Session Cookie:**
```python
import requests

url = 'http://localhost:2095/app/api/gitSyncConfig'
headers = {'Content-Type': 'application/json'}
cookies = {'session': 'your_session_cookie'}

data = {
    'enable': True,
    'provider': 'github',
    'repoUrl': 'https://github.com/username/s-ui-backup',
    'branch': 'main',
    'token': 'ghp_xxxxxxxxxxxxxxxxxxxx',
    'autoSync': True,
    'syncInterval': 3600,
    'syncConfig': True,
    'syncDb': True
}

response = requests.post(url, json=data, headers=headers, cookies=cookies)
print(response.json())
```

### Push to Git

**Using API Token:**
```python
import requests

url = 'http://localhost:2095/apiv2/gitSyncPush'
headers = {'Token': 'your_api_token_here'}

response = requests.post(url, headers=headers)
result = response.json()

if result['success']:
    print('Successfully pushed to Git')
else:
    print(f'Push failed: {result["msg"]}')
```

## Notes

- Replace `localhost:2095` with your actual S-UI server address
- **Recommended:** Use API tokens (`/apiv2/` endpoints) for automation and scripts
- **Alternative:** Use session cookies (`/app/api/` endpoints) for web interface integration
- Create API tokens in S-UI: Settings → API Tokens
- Replace `your_api_token_here` with your actual API token
- Replace Git tokens (e.g., `ghp_xxx`) with your actual Git provider tokens
- Git tokens are masked (shown as `***`) in GET responses for security
- API tokens can have expiration dates for better security
