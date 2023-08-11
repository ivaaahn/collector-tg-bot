package sessions

import (
	"collector/internal"
	"collector/internal/calculator"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type SessionUsecase struct {
	log  internal.Logger
	repo Repo
}

var UserNotFoundErr error = errors.New("user not found")

func NewUsecase(log internal.Logger, repo Repo) *SessionUsecase {
	return &SessionUsecase{log: log, repo: repo}
}

func (uc *SessionUsecase) initUser(userID int64, username string, sessionID uuid.UUID) (*User, *Member, error) {
	user := NewUser(userID, username)
	if err := uc.repo.UpsertUser(user); err != nil {
		return nil, nil, fmt.Errorf("SessionUsecase->CreateSession: %w", err)
	}

	member := NewMember(user.ID, sessionID)
	if err := uc.repo.UpsertMember(member); err != nil {
		return nil, nil, fmt.Errorf("SessionUsecase->CreateSession: %w", err)
	}

	return user, member, nil
}

func (uc *SessionUsecase) CreateSession(info CreateSessionDTO) error {
	user := NewUser(info.UserID, info.Username)
	if err := uc.repo.UpsertUser(user); err != nil {
		return fmt.Errorf("SessionUsecase->CreateSession: %w", err)
	}

	session := NewSession(uuid.New(), user.ID, info.ChatID, info.SessionName)
	if err := uc.repo.CreateSession(session); err != nil {
		return fmt.Errorf("SessionUsecase->CreateSession: %w", err)
	}

	member := NewMember(user.ID, session.ID)
	if err := uc.repo.AddMember(member); err != nil {
		return fmt.Errorf("SessionUsecase->CreateSession: %w", err)
	}

	return nil
}

func (uc *SessionUsecase) GetPurchases(chatID int64) ([]*Purchase, error) {
	session, err := uc.repo.GetSessionByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetPurchases: %w", err)
	}

	purchases, err := uc.repo.GetPurchases(session.ID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetPurchases: %w", err)
	}

	return purchases, nil
}

func (uc *SessionUsecase) AddExpenses(info AddExpensesDTO) error {
	// TODO: Сделать идемпотентным

	session, err := uc.repo.GetSessionByChatID(info.ChatID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddExpenses: %w", err)
	}

	_, _, err = uc.initUser(info.AuthorID, info.AuthorUsername, session.ID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddExpenses: %w", err)
	}

	for _, expense := range info.Expenses {
		if err = uc.addExpense(session, info.AuthorID, expense); err != nil {
			return err
		}
	}

	return nil
}

func (uc *SessionUsecase) collectEatersIDs(session *Session, authorID int64, eaters []string) ([]int64, error) {
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
				return eatersIDs, fmt.Errorf("SessionUsecase->addExpense: %w", err)
			}
			for _, user := range users {
				eatersIDs = append(eatersIDs, user.ID)
			}
		default:
			eaterUser, err := uc.repo.GetUserByUsername(eaterUsername)
			if errors.Is(err, UserNotFoundDBErr) {
				return eatersIDs, fmt.Errorf("SessionUsecase->AddExpenses: %w", UserNotFoundErr)
			} else if err != nil {
				return eatersIDs, fmt.Errorf("SessionUsecase->AddExpenses: %w", err)
			}
			eatersIDs = append(eatersIDs, eaterUser.ID)
		}
	}

	return eatersIDs, nil
}

func (uc *SessionUsecase) addExpense(session *Session, authorID int64, info *ExpenseDTO) error {
	eatersIDs, err := uc.collectEatersIDs(session, authorID, info.Eaters)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddExpenses: %w", err)
	}

	for _, eaterID := range eatersIDs {
		expense := NewExpense(info.PurchaseID, eaterID, 1, session.ID)
		if err := uc.repo.AddExpense(expense); err != nil {
			return fmt.Errorf("SessionUsecase->AddExpenses: %w", err)
		}
	}

	return nil
}

func (uc *SessionUsecase) AddPurchases(info AddPurchasesDTO) error {
	session, err := uc.repo.GetSessionByChatID(info.ChatID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddPurchases: %w", err)
	}

	_, buyerMember, err := uc.initUser(info.BuyerID, info.BuyerUsername, session.ID)
	if err != nil {
		return fmt.Errorf("SessionUsecase->AddPurchases: %w", err)
	}

	for _, purchaseInfo := range info.Purchases {
		purchase := NewPurchase(session.ID, buyerMember.UserID, purchaseInfo.Price, purchaseInfo.ProductTitle, purchaseInfo.Quantity)
		if err := uc.repo.AddPurchase(purchase); err != nil {
			return fmt.Errorf("SessionUsecase->AddPurchases: %w", err)
		}
	}

	return err
}

//	func (uc *SessionUsecase) CollectPurchases(chatID int64) (map[string]AllPurchasesDTO, error) {
//		session, err := uc.repo.GetSessionByChatID(chatID)
//		if err != nil {
//			return nil, fmt.Errorf("SessionUsecase->CollectPurchases: %w", err)
//		}
//
//		sessionPurchases, err := uc.repo.GetSessionPurchases(session.ID)
//		if err != nil {
//			return nil, fmt.Errorf("SessionUsecase->CollectPurchases: %w", err)
//		}
//
//		var buyerToPurchases = map[string]AllPurchasesDTO{}
//		for _, purchase := range sessionPurchases {
//			buyerPurchases := buyerToPurchases[purchase.AuthorUsername]
//
//			buyerPurchases.Sum += purchase.TotalDebt
//			buyerPurchases.Eaters = append(
//				buyerPurchases.Eaters,
//				UserPurchaseDTO{
//					TotalDebt:     purchase.TotalDebt,
//					Product:     purchase.Product,
//					CreatedAt: purchase.CreatedAt,
//				})
//
//			buyerToPurchases[purchase.AuthorUsername] = buyerPurchases
//		}
//
//		return buyerToPurchases, nil
//	}

func (uc *SessionUsecase) GetAllDebts(chatID int64) (map[string]map[string]*calculator.Debt, error) {
	session, err := uc.repo.GetSessionByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	products, err := uc.repo.GetProducts(session.ID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	productsMap := make(map[int64]calculator.Product)
	for _, product := range products {
		productsMap[product.ID] = *product
	}

	expenses, err := uc.repo.GetExpenses(session.ID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	usersMap := make(map[int64]*calculator.User, 0)
	for _, expense := range expenses {
		_, ok := usersMap[expense.EaterID]
		if !ok {
			usersMap[expense.EaterID] = &calculator.User{
				ID:       expense.EaterID,
				Username: expense.EaterUsername,
				Expenses: make([]calculator.Expense, 0),
			}
		}

		usersMap[expense.EaterID].Expenses = append(usersMap[expense.EaterID].Expenses, *expense)
	}

	users := make([]calculator.User, 0)
	for _, user := range usersMap {
		users = append(users, *user)
	}

	calc := calculator.New(users, productsMap)

	return calc.Calculate(), nil
}

//	func (uc *SessionUsecase) FinishSession(chatID int64) error {
//		session, err := uc.repo.GetSessionByChatID(chatID)
//		if err != nil {
//			return fmt.Errorf("SessionUsecase->ChangeSessionStateToClosed: %w", err)
//		}
//
//		if !session.IsActive() {
//			return SessionIsNotActiveErr
//		}
//
//		return uc.repo.ChangeSessionStateToClosed(session.ID)
//	}
//
// ---
func (uc *SessionUsecase) IsSessionExists(chatID int64) (bool, error) {
	session, err := uc.repo.GetSessionByChatID(chatID)

	if err == nil {
		return session.IsActive(), nil
	}

	if errors.Is(err, SessionNotFoundDBErr) {
		return false, nil
	}

	return false, err
}
