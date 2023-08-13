package debts_calculation

import (
	"github.com/google/uuid"
)

type Session struct {
	ID uuid.UUID
}

type Product struct {
	ID        int64
	Title     string
	NumOfEats int
	Price     int
}

type Expense struct {
	ProductID     int64
	BuyerID       int64
	BuyerUsername string
	EaterUsername string
	EaterID       int64
}

type PaymentKind int8

const INCOME PaymentKind = 1
const OUTCOME PaymentKind = -1

type HistoryItem struct {
	Kind         PaymentKind
	ProductTitle string
	Price        int
}

type Debt struct {
	TotalDebt int
	History   []HistoryItem
}

type User struct {
	ID       int64
	Username string
	Expenses []Expense
}
