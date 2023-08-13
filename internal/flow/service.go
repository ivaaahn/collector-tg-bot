package flow

import (
	"collector/internal"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type Service struct {
	log  internal.Logger
	repo Repo
}

var UserNotFoundErr = errors.New("user not found")

func NewService(log internal.Logger, repo Repo) *Service {
	return &Service{log: log, repo: repo}
}

func (uc *Service) initUser(userID int64, username string, sessionID uuid.UUID) (*User, *Member, error) {
	user := NewUser(userID, username)
	if err := uc.repo.UpsertUser(user); err != nil {
		return nil, nil, fmt.Errorf("Service->CreateSession: %w", err)
	}

	member := NewMember(user.ID, sessionID)
	if err := uc.repo.UpsertMember(member); err != nil {
		return nil, nil, fmt.Errorf("Service->CreateSession: %w", err)
	}

	return user, member, nil
}

func (uc *Service) GetPurchases(chatID int64) ([]*Purchase, error) {
	session, err := uc.repo.GetSessionByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("Service->GetPurchases: %w", err)
	}

	purchases, err := uc.repo.GetPurchases(session.ID)
	if err != nil {
		return nil, fmt.Errorf("Service->GetPurchases: %w", err)
	}

	return purchases, nil
}

func (uc *Service) AddExpenses(info AddExpensesDTO) error {
	// TODO: Сделать идемпотентным

	session, err := uc.repo.GetSessionByChatID(info.ChatID)
	if err != nil {
		return fmt.Errorf("Service->AddExpenses: %w", err)
	}

	_, _, err = uc.initUser(info.AuthorID, info.AuthorUsername, session.ID)
	if err != nil {
		return fmt.Errorf("Service->AddExpenses: %w", err)
	}

	for _, expense := range info.Expenses {
		if err = uc.addExpense(session, info.AuthorID, expense); err != nil {
			return err
		}
	}

	return nil
}

func (uc *Service) collectEatersIDs(session *Session, authorID int64, eaters []string) ([]int64, error) {
	if len(eaters) == 0 {
		return []int64{authorID}, nil
	}

	eatersIDs := make([]int64, 0)
	for _, eaterUsername := range eaters {
		switch eaterUsername {
		case "me":
			eatersIDs = append(eatersIDs, authorID)
		case "all":
			users, err := uc.repo.GetUsersBySessionID(session.ID)
			if err != nil {
				return eatersIDs, fmt.Errorf("Service->addExpense: %w", err)
			}
			for _, user := range users {
				eatersIDs = append(eatersIDs, user.ID)
			}
		default:
			eaterUser, err := uc.repo.GetUserByUsername(eaterUsername)
			if errors.Is(err, UserNotFoundDBErr) {
				return eatersIDs, fmt.Errorf("Service->AddExpenses: %w", UserNotFoundErr)
			} else if err != nil {
				return eatersIDs, fmt.Errorf("Service->AddExpenses: %w", err)
			}
			eatersIDs = append(eatersIDs, eaterUser.ID)
		}
	}

	return eatersIDs, nil
}

func (uc *Service) addExpense(session *Session, authorID int64, info *ExpenseDTO) error {
	eatersIDs, err := uc.collectEatersIDs(session, authorID, info.Eaters)
	if err != nil {
		return fmt.Errorf("Service->AddExpenses: %w", err)
	}

	for _, eaterID := range eatersIDs {
		expense := NewExpense(info.PurchaseID, eaterID, 1, session.ID)
		if err := uc.repo.AddExpense(expense); err != nil {
			return fmt.Errorf("Service->AddExpenses: %w", err)
		}
	}

	return nil
}

func (uc *Service) AddPurchases(info AddPurchasesDTO) error {
	session, err := uc.repo.GetSessionByChatID(info.ChatID)
	if err != nil {
		return fmt.Errorf("Service->AddPurchases: %w", err)
	}

	_, buyerMember, err := uc.initUser(info.BuyerID, info.BuyerUsername, session.ID)
	if err != nil {
		return fmt.Errorf("Service->AddPurchases: %w", err)
	}

	for _, purchaseInfo := range info.Purchases {
		purchase := NewPurchase(session.ID, buyerMember.UserID, purchaseInfo.Price, purchaseInfo.ProductTitle, purchaseInfo.Quantity)
		if err := uc.repo.AddPurchase(purchase); err != nil {
			return fmt.Errorf("Service->AddPurchases: %w", err)
		}
	}

	return err
}
