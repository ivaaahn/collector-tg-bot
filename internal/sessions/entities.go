package sessions

import (
	"collector/internal"
	"github.com/google/uuid"
	"time"
)

const (
	SessionActiveState = "active"
	SessionClosedState = "closed"
)

type Session struct {
	UUID        uuid.UUID
	CreatorID   int64
	ChatID      int64
	SessionName string
	StartedAt   string
	State       string
}

type Cost struct {
	UserID      int64
	Money       int
	MemberID    int64
	Description string
}

type Member struct {
	ID          int64
	SessionUUID internal.UUID
	UserID      int64
}

type UserCost struct {
	Username    string
	Amount      int
	CreatedAt   time.Time
	Description string
}

type MemberCost struct {
	UserID int64
	Money  int
}

func NewEmptyMemberCost() *Cost {
	return &Cost{}
}

func NewMemberCost(userID int64, money int) *Cost {
	return &Cost{
		UserID: userID,
		Money:  money,
	}
}

func NewEmptyUserCost() *UserCost {
	return &UserCost{}
}

func NewUserCost(username string, cost int, description string) *UserCost {
	return &UserCost{
		Username:    username,
		Amount:      cost,
		Description: description,
	}
}

func NewEmptyMember() *Member {
	return &Member{}
}

func NewMember(sessionUUID internal.UUID, userID int64) *Member {
	return &Member{
		ID:          0,
		SessionUUID: sessionUUID,
		UserID:      userID,
	}
}

func NewCost(userID int64, money int, memberID int64, description string) *Cost {
	return &Cost{
		UserID:      userID,
		Money:       money,
		MemberID:    memberID,
		Description: description,
	}
}

func NewSession(UUID uuid.UUID, creatorID int64, chatID int64, sessionName string) *Session {
	return &Session{
		UUID:        UUID,
		CreatorID:   creatorID,
		ChatID:      chatID,
		SessionName: sessionName,
		State:       SessionActiveState,
	}
}

func NewEmptySession() *Session {
	return &Session{}
}

func (s *Session) IsActive() bool {
	return s.State == SessionActiveState
}
