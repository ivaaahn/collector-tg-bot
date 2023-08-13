package debts_calculation

import (
	"collector/internal"
	"fmt"
)

type Service struct {
	log  internal.Logger
	repo Repo
}

func NewService(log internal.Logger, repo Repo) *Service {
	return &Service{log: log, repo: repo}
}
func (uc *Service) GetAllDebts(chatID int64) (map[string]map[string]*Debt, error) {
	session, err := uc.repo.GetSessionByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	products, err := uc.repo.GetProducts(session.ID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	productsMap := make(map[int64]Product)
	for _, product := range products {
		productsMap[product.ID] = *product
	}

	expenses, err := uc.repo.GetExpenses(session.ID)
	if err != nil {
		return nil, fmt.Errorf("SessionUsecase->GetAllDebts: %w", err)
	}

	usersMap := make(map[int64]*User, 0)
	for _, expense := range expenses {
		_, ok := usersMap[expense.EaterID]
		if !ok {
			usersMap[expense.EaterID] = &User{
				ID:       expense.EaterID,
				Username: expense.EaterUsername,
				Expenses: make([]Expense, 0),
			}
		}

		usersMap[expense.EaterID].Expenses = append(usersMap[expense.EaterID].Expenses, *expense)
	}

	users := make([]User, 0)
	for _, user := range usersMap {
		users = append(users, *user)
	}

	calc := NewCalculator(users, productsMap)

	return calc.Calculate(), nil
}
