package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnumRole string

const (
	USER  EnumRole = "USER"
	ADMIN EnumRole = "ADMIN"
)

type User struct {
	ID            uint      `gorm:"primaryKey;column:id"`
	PublicID      string    `gorm:"uniqueIndex;not null;column:public_id"`
	Name          string    `gorm:"not null;column:name"`
	Email         string    `gorm:"uniqueIndex;not null;column:email"`
	EmailVerified bool      `gorm:"default:false;column:email_verified"`
	Image         *string   `gorm:"column:image"`
	Role          EnumRole  `gorm:"type:enum('USER','ADMIN');default:'USER';column:role;index"`
	Password      string    `gorm:"not null;column:password"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at;index"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

// BeforeCreate hook to generate PublicID
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.PublicID == "" {
		u.PublicID = uuid.New().String()
	}
	return
}

func (User) TableName() string {
	return "tbl_users"
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		PublicID:      u.PublicID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
		Role:          u.Role,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

type UserResponse struct {
	ID            uint      `json:"id"`
	PublicID      string    `json:"public_id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Image         *string   `json:"image"`
	Role          EnumRole  `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
