package cronjob

import (
	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/service"
)

type GitSyncJob struct {
	gitSyncService service.GitSyncService
}

func NewGitSyncJob() *GitSyncJob {
	return &GitSyncJob{}
}

func (j *GitSyncJob) Run() {
	config, err := j.gitSyncService.GetConfig()
	if err != nil {
		db := database.GetDB()
		var count int64
		db.Model(&model.GitSync{}).Count(&count)
		if count == 0 {
			return
		}
		logger.Debug("Git sync job: failed to get config:", err)
		return
	}

	if !config.Enable || !config.AutoSync {
		return
	}

	logger.Debug("Running Git sync job")
	err = j.gitSyncService.PushToGit()
	if err != nil {
		logger.Error("Git sync job failed:", err)
	}
}

type GitSyncCheckJob struct {
	gitSyncService service.GitSyncService
}

func NewGitSyncCheckJob() *GitSyncCheckJob {
	return &GitSyncCheckJob{}
}

func (j *GitSyncCheckJob) Run() {
	config, err := j.gitSyncService.GetConfig()
	if err != nil {
		return
	}

	if !config.Enable || !config.SyncConfig {
		return
	}

	err = j.gitSyncService.CheckAndPullIfNewer()
	if err != nil {
		logger.Debug("Git sync check:", err)
	}
}