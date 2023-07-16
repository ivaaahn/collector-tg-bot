package users

import (
	"collector/internal"
	"fmt"
)

type Usecase struct {
	log  internal.Logger
	repo Repo
}

func NewUsecase(log internal.Logger, repo Repo) *Usecase {
	return &Usecase{log: log, repo: repo}
}

func (uc *Usecase) Upsert(userID int64, username string) (int64, error) {
	user, err := uc.repo.GetById(userID)

	if err != nil {
		return 0, fmt.Errorf("usecase: %v", err)
	}

	// If user not exists, you should create user
	if user.ID == 0 {
		user.Username = username
		user.TgID = userID
		user.ID, err = uc.repo.Create(user)
		if err != nil {
			return 0, fmt.Errorf("Usecase->Upsert: %w", err)
		}
	}
	return user.ID, nil
}
