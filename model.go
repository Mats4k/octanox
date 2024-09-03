package octanox

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username string
	Email    string
	Avatar   *string
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Token     string    `gorm:"type:text;not null"`
	IsRevoked bool      `gorm:"type:boolean;default:false"`
}
