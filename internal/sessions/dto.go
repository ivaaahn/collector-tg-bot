package sessions

import "time"

type UserCostDTO struct {
	Amount      int
	Description string
	CreatedAt   time.Time
}

type AllUserCostsDTO struct {
	Sum   int
	Costs []UserCostDTO
}

type UserDebtDTO struct {
	DebtorName string
	Money      int
}

type AllUserDebtsDTO struct {
	Debts []UserDebtDTO
}

type AddExpenseDTO struct {
	Product  string
	ChatID   int64
	UserID   int64
	Username string
	Cost     int
}

type CreateSessionDTO struct {
	UserID      int64
	ChatID      int64
	Username    string
	SessionName string
}
