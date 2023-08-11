package sessions

import (
	"time"
)

type UserPurchaseDTO struct {
	Price     int
	Title     string
	CreatedAt time.Time
}

type AllPurchasesDTO struct {
	Sum       int
	Purchases []UserPurchaseDTO
}

type UserDebtDTO struct {
	DebtorName string
	Money      int
}

type AllUserDebtsDTO struct {
	Debts []UserDebtDTO
}

type ExpenseDTO struct {
	PurchaseID int64
	Quantity   int
	Eaters     []string
}

type PurchaseDTO struct {
	ProductTitle string
	Price        int
	Quantity     int
}

type AddPurchasesDTO struct {
	ChatID        int64
	BuyerID       int64
	BuyerUsername string
	Purchases     []*PurchaseDTO
}

type AddExpensesDTO struct {
	ChatID         int64
	AuthorID       int64
	AuthorUsername string
	Expenses       []*ExpenseDTO
}

type CreateSessionDTO struct {
	UserID      int64
	ChatID      int64
	Username    string
	SessionName string
}
