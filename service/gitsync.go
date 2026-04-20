package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util/common"
)

type GitSyncService struct {
	ConfigService
	SettingService
}

type GitProvider interface {
	GetFile(path string) ([]byte, error)
	CreateOrUpdateFile(path string, content []byte, message string) error
	DeleteFile(path string, message string) error
}

type GitHubProvider struct {
	token    string
	owner    string
	repo     string
	branch   string
	apiBase  string
}

type GitLabProvider struct {
	token      string
	projectId  string
	branch     string
	apiBase    string
}

type GiteaProvider struct {
	token    string
	owner    string
	repo     string
	branch   string
	apiBase  string
}

func (s *GitSyncService) GetConfig() (*model.GitSync, error) {
	db := database.GetDB()
	var config model.GitSync
	err := db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *GitSyncService) SaveConfig(config *model.GitSync) error {
	db := database.GetDB()
	var existing model.GitSync
	err := db.First(&existing).Error
	
	if err != nil {
		return db.Create(config).Error
	}
	
	config.Id = existing.Id
	return db.Save(config).Error
}

func (s *GitSyncService) getProvider(config *model.GitSync) (GitProvider, error) {
	switch strings.ToLower(config.Provider) {
	case "github":
		return s.newGitHubProvider(config)
	case "gitlab":
		return s.newGitLabProvider(config)
	case "gitea":
		return s.newGiteaProvider(config)
	default:
		return nil, common.NewError("unsupported git provider: ", config.Provider)
	}
}

func (s *GitSyncService) newGitHubProvider(config *model.GitSync) (*GitHubProvider, error) {
	parts := strings.Split(strings.TrimPrefix(config.RepoUrl, "https://github.com/"), "/")
	if len(parts) < 2 {
		return nil, common.NewError("invalid GitHub repo URL")
	}
	
	return &GitHubProvider{
		token:   config.Token,
		owner:   parts[0],
		repo:    strings.TrimSuffix(parts[1], ".git"),
		branch:  config.Branch,
		apiBase: "https://api.github.com",
	}, nil
}

func (s *GitSyncService) newGitLabProvider(config *model.GitSync) (*GitLabProvider, error) {
	repoUrl := config.RepoUrl
	apiBase := "https://gitlab.com/api/v4"
	
	if strings.Contains(repoUrl, "gitlab.com") {
		parts := strings.Split(strings.TrimPrefix(repoUrl, "https://gitlab.com/"), "/")
		if len(parts) < 2 {
			return nil, common.NewError("invalid GitLab repo URL")
		}
		projectId := strings.TrimSuffix(parts[0]+"/"+parts[1], ".git")
		projectId = strings.ReplaceAll(projectId, "/", "%2F")
		
		return &GitLabProvider{
			token:     config.Token,
			projectId: projectId,
			branch:    config.Branch,
			apiBase:   apiBase,
		}, nil
	}
	
	parts := strings.Split(repoUrl, "/")
	if len(parts) < 5 {
		return nil, common.NewError("invalid GitLab repo URL")
	}
	
	apiBase = strings.Join(parts[:3], "/") + "/api/v4"
	projectId := strings.TrimSuffix(parts[3]+"/"+parts[4], ".git")
	projectId = strings.ReplaceAll(projectId, "/", "%2F")
	
	return &GitLabProvider{
		token:     config.Token,
		projectId: projectId,
		branch:    config.Branch,
		apiBase:   apiBase,
	}, nil
}

func (s *GitSyncService) newGiteaProvider(config *model.GitSync) (*GiteaProvider, error) {
	parts := strings.Split(config.RepoUrl, "/")
	if len(parts) < 5 {
		return nil, common.NewError("invalid Gitea repo URL")
	}
	
	apiBase := strings.Join(parts[:3], "/") + "/api/v1"
	owner := parts[3]
	repo := strings.TrimSuffix(parts[4], ".git")
	
	return &GiteaProvider{
		token:   config.Token,
		owner:   owner,
		repo:    repo,
		branch:  config.Branch,
		apiBase: apiBase,
	}, nil
}

// GitHub Provider Implementation
func (p *GitHubProvider) GetFile(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", p.apiBase, p.owner, p.repo, path, p.branch)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return nil, nil
	}
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(result.Content, "\n", ""))
	if err != nil {
		return nil, err
	}
	
	return decoded, nil
}

func (p *GitHubProvider) CreateOrUpdateFile(path string, content []byte, message string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", p.apiBase, p.owner, p.repo, path)
	
	var sha string
	existing, _ := p.GetFile(path)
	if existing != nil {
		req, _ := http.NewRequest("GET", url+"?ref="+p.branch, nil)
		req.Header.Set("Authorization", "token "+p.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		
		client := &http.Client{Timeout: 30 * time.Second}
		resp, _ := client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
			var result struct {
				Sha string `json:"sha"`
			}
			json.NewDecoder(resp.Body).Decode(&result)
			sha = result.Sha
		}
	}
	
	payload := map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
		"branch":  p.branch,
	}
	
	if sha != "" {
		payload["sha"] = sha
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (p *GitHubProvider) DeleteFile(path string, message string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", p.apiBase, p.owner, p.repo, path)
	
	req, _ := http.NewRequest("GET", url+"?ref="+p.branch, nil)
	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, _ := client.Do(req)
	if resp == nil {
		return common.NewError("failed to get file SHA")
	}
	defer resp.Body.Close()
	
	var result struct {
		Sha string `json:"sha"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	payload := map[string]string{
		"message": message,
		"sha":     result.Sha,
		"branch":  p.branch,
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ = http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GitLab Provider Implementation
func (p *GitLabProvider) GetFile(path string) ([]byte, error) {
	encodedPath := strings.ReplaceAll(path, "/", "%2F")
	url := fmt.Sprintf("%s/projects/%s/repository/files/%s?ref=%s", p.apiBase, p.projectId, encodedPath, p.branch)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("PRIVATE-TOKEN", p.token)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return nil, nil
	}
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	decoded, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return nil, err
	}
	
	return decoded, nil
}

func (p *GitLabProvider) CreateOrUpdateFile(path string, content []byte, message string) error {
	encodedPath := strings.ReplaceAll(path, "/", "%2F")
	url := fmt.Sprintf("%s/projects/%s/repository/files/%s", p.apiBase, p.projectId, encodedPath)
	
	existing, _ := p.GetFile(path)
	action := "create"
	if existing != nil {
		action = "update"
	}
	
	payload := map[string]string{
		"branch":         p.branch,
		"content":        base64.StdEncoding.EncodeToString(content),
		"commit_message": message,
		"encoding":       "base64",
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	method := "POST"
	if action == "update" {
		method = "PUT"
	}
	
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("PRIVATE-TOKEN", p.token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (p *GitLabProvider) DeleteFile(path string, message string) error {
	encodedPath := strings.ReplaceAll(path, "/", "%2F")
	url := fmt.Sprintf("%s/projects/%s/repository/files/%s", p.apiBase, p.projectId, encodedPath)
	
	payload := map[string]string{
		"branch":         p.branch,
		"commit_message": message,
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
	req.Header.Set("PRIVATE-TOKEN", p.token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Gitea Provider Implementation
func (p *GiteaProvider) GetFile(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", p.apiBase, p.owner, p.repo, path, p.branch)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+p.token)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return nil, nil
	}
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gitea API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(result.Content, "\n", ""))
	if err != nil {
		return nil, err
	}
	
	return decoded, nil
}

func (p *GiteaProvider) CreateOrUpdateFile(path string, content []byte, message string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", p.apiBase, p.owner, p.repo, path)
	
	var sha string
	existing, _ := p.GetFile(path)
	if existing != nil {
		req, _ := http.NewRequest("GET", url+"?ref="+p.branch, nil)
		req.Header.Set("Authorization", "token "+p.token)
		
		client := &http.Client{Timeout: 30 * time.Second}
		resp, _ := client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
			var result struct {
				Sha string `json:"sha"`
			}
			json.NewDecoder(resp.Body).Decode(&result)
			sha = result.Sha
		}
	}
	
	payload := map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
		"branch":  p.branch,
	}
	
	if sha != "" {
		payload["sha"] = sha
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Gitea API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (p *GiteaProvider) DeleteFile(path string, message string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", p.apiBase, p.owner, p.repo, path)
	
	req, _ := http.NewRequest("GET", url+"?ref="+p.branch, nil)
	req.Header.Set("Authorization", "token "+p.token)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, _ := client.Do(req)
	if resp == nil {
		return common.NewError("failed to get file SHA")
	}
	defer resp.Body.Close()
	
	var result struct {
		Sha string `json:"sha"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	payload := map[string]string{
		"message": message,
		"sha":     result.Sha,
		"branch":  p.branch,
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ = http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Gitea API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (s *GitSyncService) getHostname() (string, error) {
	hostname, err := s.SettingService.getString("hostname")
	if err != nil || hostname == "" {
		hostname = "default"
	}
	return hostname, nil
}

func (s *GitSyncService) getConfigHash(config []byte) string {
	return fmt.Sprintf("%x", time.Now().Unix())
}

// Sync Operations
func (s *GitSyncService) PushToGit() error {
	config, err := s.GetConfig()
	if err != nil || !config.Enable {
		return err
	}
	
	provider, err := s.getProvider(config)
	if err != nil {
		return err
	}
	
	hostname, err := s.getHostname()
	if err != nil {
		hostname = "default"
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	version := time.Now().Unix()
	
	if config.SyncConfig {
		rawConfig, err := s.ConfigService.GetConfig("")
		if err != nil {
			logger.Error("Failed to get SingBox config:", err)
		} else {
			configPath := fmt.Sprintf("%s/singbox-config.json", hostname)
			err = provider.CreateOrUpdateFile(configPath, *rawConfig, "Update SingBox config - "+timestamp)
			if err != nil {
				logger.Error("Failed to push SingBox config:", err)
			} else {
				logger.Info("SingBox config pushed to Git")
				
				versionPath := fmt.Sprintf("%s/version.txt", hostname)
				versionContent := []byte(fmt.Sprintf("%d", version))
				provider.CreateOrUpdateFile(versionPath, versionContent, "Update version - "+timestamp)
			}
		}
	}
	
	if config.SyncDb {
		db, err := database.GetDb("stats,changes")
		if err != nil {
			logger.Error("Failed to get database:", err)
		} else {
			dbPath := fmt.Sprintf("%s/s-ui-backup.db", hostname)
			err = provider.CreateOrUpdateFile(dbPath, db, "Update database backup - "+timestamp)
			if err != nil {
				logger.Error("Failed to push database:", err)
			} else {
				logger.Info("Database pushed to Git")
			}
		}
	}
	
	db := database.GetDB()
	config.LastSync = time.Now().Unix()
	db.Save(config)
	
	return nil
}

func (s *GitSyncService) PullFromGit() error {
	config, err := s.GetConfig()
	if err != nil || !config.Enable {
		return err
	}
	
	provider, err := s.getProvider(config)
	if err != nil {
		return err
	}
	
	hostname, err := s.getHostname()
	if err != nil {
		hostname = "default"
	}
	
	if config.SyncConfig {
		configPath := fmt.Sprintf("%s/singbox-config.json", hostname)
		content, err := provider.GetFile(configPath)
		if err != nil {
			logger.Error("Failed to pull SingBox config:", err)
		} else if content != nil {
			logger.Info("SingBox config pulled from Git")
			
			db := database.GetDB()
			tx := db.Begin()
			err = s.SettingService.SaveConfig(tx, content)
			if err != nil {
				tx.Rollback()
				logger.Error("Failed to save pulled config:", err)
			} else {
				tx.Commit()
				go func() { _ = s.ConfigService.RestartCore() }()
				logger.Info("Config applied and core restarted")
			}
		}
	}
	
	if config.SyncDb {
		dbPath := fmt.Sprintf("%s/s-ui-backup.db", hostname)
		content, err := provider.GetFile(dbPath)
		if err != nil {
			logger.Error("Failed to pull database:", err)
		} else if content != nil {
			logger.Info("Database pulled from Git (manual import required)")
		}
	}
	
	db := database.GetDB()
	config.LastSync = time.Now().Unix()
	db.Save(config)
	
	return nil
}

func (s *GitSyncService) CheckAndPullIfNewer() error {
	config, err := s.GetConfig()
	if err != nil || !config.Enable || !config.SyncConfig {
		return err
	}
	
	provider, err := s.getProvider(config)
	if err != nil {
		return err
	}
	
	hostname, err := s.getHostname()
	if err != nil {
		hostname = "default"
	}
	
	versionPath := fmt.Sprintf("%s/version.txt", hostname)
	remoteVersion, err := provider.GetFile(versionPath)
	if err != nil || remoteVersion == nil {
		return err
	}
	
	remoteVersionInt, err := strconv.ParseInt(strings.TrimSpace(string(remoteVersion)), 10, 64)
	if err != nil {
		return err
	}
	
	if config.LastSync < remoteVersionInt {
		logger.Info("Remote config is newer, pulling...")
		return s.PullFromGit()
	}
	
	return nil
}

func (s *GitSyncService) TestConnection() error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	
	provider, err := s.getProvider(config)
	if err != nil {
		return err
	}
	
	_, err = provider.GetFile("README.md")
	if err != nil {
		return err
	}
	
	return nil
}