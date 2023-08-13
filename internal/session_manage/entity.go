package session_manage

import (
	"collector/internal"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        int64
	Username  string
	CreatedAt time.Time
}

type Session struct {
	ID         uuid.UUID
	Title      string
	CreatorID  int64
	ChatID     int64
	CreatedAt  time.Time
	FinishedAt sql.NullTime
}

type Member struct {
	UserID    int64
	SessionID internal.UUID
}

func NewSession(UUID uuid.UUID, creatorID int64, chatID int64, title string) *Session {
	return &Session{
		ID:        UUID,
		CreatorID: creatorID,
		ChatID:    chatID,
		Title:     title,
	}
}

func NewUser(ID int64, username string) *User {
	return &User{
		ID:       ID,
		Username: username,
	}
}

func NewMember(userID int64, sessionID uuid.UUID) *Member {
	return &Member{
		UserID:    userID,
		SessionID: sessionID,
	}
}
