package session_manage

import (
	"collector/internal"
	"fmt"
	"github.com/google/uuid"
)

type Service struct {
	log  internal.Logger
	repo Repo
}

func NewService(log internal.Logger, repo Repo) *Service {
	return &Service{log: log, repo: repo}
}

func (uc *Service) CreateSession(info CreateSessionDTO) error {
	user := NewUser(info.UserID, info.Username)
	if err := uc.repo.UpsertUser(user); err != nil {
		return fmt.Errorf("Service->CreateSession: %w", err)
	}

	session := NewSession(uuid.New(), user.ID, info.ChatID, info.SessionName)
	if err := uc.repo.CreateSession(session); err != nil {
		return fmt.Errorf("Service->CreateSession: %w", err)
	}

	member := NewMember(user.ID, session.ID)
	if err := uc.repo.AddMember(member); err != nil {
		return fmt.Errorf("Service->CreateSession: %w", err)
	}

	return nil
}

func (uc *Service) FinishSession(sessionID uuid.UUID) error {
	session, err := uc.repo.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("Service->SetFinishTimeToNow: %w", err)
	}

	return uc.repo.SetFinishTimeToNow(session.ID)
}
