package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	Email        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         string         `gorm:"type:varchar(20);default:'member'" json:"role"` // owner, admin, member
	TeamID       *uint          `gorm:"index" json:"team_id"`
	Avatar       string         `gorm:"type:varchar(500)" json:"avatar"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

func (u *User) TableName() string {
	return "users"
}

type Team struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID   uint           `gorm:"not null" json:"owner_id"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Owner   *User  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []User `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

func (t *Team) TableName() string {
	return "teams"
}

type Invitation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TeamID    uint      `gorm:"not null;index" json:"team_id"`
	Email     string    `gorm:"type:varchar(100);not null" json:"email"`
	Token     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	Status    string    `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, accepted, expired
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`

	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

func (i *Invitation) TableName() string {
	return "invitations"
}
