package sessions

import (
	"collector/internal"
	"collector/internal/users"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type DebtsMatrix map[int64]map[int64]int

type UserUsecase interface {
	Upsert(userID int64, username string) (int64, error)
}

type UserRepo interface {
	GetById(ID int64) (*users.User, error)
	GetBySessionID(sessionUUID internal.UUID) ([]*users.User, error)
}

type SessionUsecase struct {
	log         internal.Logger
	repo        Repo
	userUsecase UserUsecase
	userRepo    UserRepo
}

func NewUsecase(log internal.Logger, userUsecase UserUsecase, repo Repo, usersRepo UserRepo) *SessionUsecase {
	return &SessionUsecase{log: log, repo: repo, userUsecase: userUsecase, userRepo: usersRepo}
}

func (uc *SessionUsecase) IsSessionExists(chatID int64) (bool, error) {
	session, err := uc.repo.GetByChatID(chatID)
	if err != nil {
		return false, err
	}

	// todo: drop when rewrite repo
	if session.IsActive() {
		return true, nil
	}

	return false, nil
}

func (uc *SessionUsecase) CreateSession(info CreateSessionDTO) error {
	userID, err := uc.userUsecase.Upsert(info.UserID, info.Username)
	if err != nil {
		return fmt.Errorf("GroupUsecase->CreateSession: %w", err)
	}

	session := NewSession(uuid.New(), userID, info.ChatID, info.SessionName)
	err = uc.repo.InsertSession(session)
	if err != nil {
		return fmt.Errorf("usecase: %v", err.Error())
	}

	// Add creator to members
	_, err = uc.repo.AddMember(session.UUID, userID)
	return err
}

func (uc *SessionUsecase) getOrCreateMember(sessionUUID internal.UUID, userID int64) (*Member, error) {
	member, err := uc.repo.GetMember(sessionUUID, userID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->getOrCreateMember: %w", err)
	}

	if member.ID == 0 {
		member.ID, err = uc.repo.AddMember(sessionUUID, userID)
		if err != nil {
			return nil, fmt.Errorf("SessionUsecase->getOrCreateMember: %w", err)
		}
	}

	return member, nil
}

func (uc *SessionUsecase) AddCost(info AddExpenseDTO) error {
	session, err := uc.repo.GetByChatID(info.ChatID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddCost: %w", err)
	}
	if session.IsActive() == false {
		return errors.New("SessionUsecase->AddCost: session is not active")
	}

	userID, err := uc.userUsecase.Upsert(info.UserID, info.Username)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddCost: %w", err)
	}

	member, err := uc.getOrCreateMember(session.UUID, userID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddCost: %w", err)
	}

	cost := NewCost(userID, info.Cost, member.ID, info.Product)

	return uc.repo.AddMemberCost(cost)
}

func (uc *SessionUsecase) CollectCosts(chatID int64) (map[string]AllUserCostsDTO, error) {
	session, err := uc.repo.GetByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->CollectCosts: %w", err)
	}

	costs, err := uc.repo.SelectUsersCosts(session.UUID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->CollectCosts: %w", err)
	}

	var usersCosts = map[string]AllUserCostsDTO{}
	for _, cost := range costs {
		userResult := usersCosts[cost.Username]

		userResult.Sum += cost.Amount
		userResult.Costs = append(
			userResult.Costs,
			UserCostDTO{
				Amount:      cost.Amount,
				Description: cost.Description,
				CreatedAt:   cost.CreatedAt,
			})

		usersCosts[cost.Username] = userResult
	}

	return usersCosts, nil
}

func (uc *SessionUsecase) makeDebtsMatrix(sessionUUID uuid.UUID) (DebtsMatrix, error) {
	sessionUsers, err := uc.userRepo.GetBySessionID(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->makeDebtsMatrix: %w", err)
	}

	var debtsMatrix = make(DebtsMatrix)
	for _, user := range sessionUsers {
		debtor := make(map[int64]int)
		for _, tmpUser := range sessionUsers {
			debtor[tmpUser.ID] = 0
		}
		debtsMatrix[user.ID] = debtor
	}

	memberCosts, err := uc.repo.GetMemberCosts(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->makeDebtsMatrix: %w", err)
	}

	for _, memberCost := range memberCosts {
		for memberDebtorID := range debtsMatrix[memberCost.UserID] {
			if memberDebtorID != memberCost.UserID {
				avgDebt := memberCost.Money / len(sessionUsers)
				debtsMatrix[memberCost.UserID][memberDebtorID] += avgDebt
			}
		}
	}

	for user, debtors := range debtsMatrix {
		for debtor := range debtors {
			if debtsMatrix[user][debtor] == 0 || debtsMatrix[debtor][user] == 0 {
				continue
			}

			if debtsMatrix[user][debtor] > debtsMatrix[debtor][user] {
				debtsMatrix[user][debtor] -= debtsMatrix[debtor][user]
				debtsMatrix[debtor][user] = 0
			} else {
				debtsMatrix[debtor][user] -= debtsMatrix[user][debtor]
				debtsMatrix[user][debtor] = 0
			}
		}
	}
	return debtsMatrix, nil
}

func (uc *SessionUsecase) GetAllDebts(chatID int64) (map[string]AllUserDebtsDTO, error) {
	session, err := uc.repo.GetByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	debtsMatrix, err := uc.makeDebtsMatrix(session.UUID)
	if err != nil {
		return nil, err
	}

	debts := map[string]AllUserDebtsDTO{}
	for userID, debtorsIDs := range debtsMatrix {
		for debtorID := range debtorsIDs {
			if debtsMatrix[userID][debtorID] != 0 {
				creditor, _ := uc.userRepo.GetById(userID)
				debtor, _ := uc.userRepo.GetById(debtorID)
				debt := debtsMatrix[userID][debtorID]

				userDebts := debts[creditor.Username]
				userDebts.Debts = append(
					userDebts.Debts,
					UserDebtDTO{
						DebtorName: debtor.Username,
						Money:      debt,
					})
				debts[creditor.Username] = userDebts
			}
		}
	}

	return debts, nil
}

func (uc *SessionUsecase) FinishSession(chatID int64) error {
	session, err := uc.repo.GetByChatID(chatID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->ChangeSessionStateToClosed: %w", err)
	}

	if session.IsActive() {
		return SessionIsNotActiveErr
	}

	return uc.repo.ChangeSessionStateToClosed(session.UUID)
}
