package flow

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
