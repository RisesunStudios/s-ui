package model

type GitSync struct {
	Id           uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable       bool   `json:"enable" form:"enable" gorm:"default:false;not null"`
	Provider     string `json:"provider" form:"provider"` // github, gitlab, gitea
	RepoUrl      string `json:"repoUrl" form:"repoUrl"`
	Branch       string `json:"branch" form:"branch" gorm:"default:main"`
	Token        string `json:"token" form:"token"`
	AutoSync     bool   `json:"autoSync" form:"autoSync" gorm:"default:false;not null"`
	SyncInterval int    `json:"syncInterval" form:"syncInterval" gorm:"default:3600"` // seconds
	LastSync     int64  `json:"lastSync" form:"lastSync" gorm:"default:0"`
	SyncConfig   bool   `json:"syncConfig" form:"syncConfig" gorm:"default:true;not null"`
	SyncDb       bool   `json:"syncDb" form:"syncDb" gorm:"default:true;not null"`
}