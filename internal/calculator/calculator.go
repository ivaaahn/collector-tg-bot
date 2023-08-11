package calculator

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

type Calculator struct {
	Users    []User
	Products map[int64]Product
	resMap   map[string]map[string]*Debt
}

func New(users []User, products map[int64]Product) *Calculator {
	return &Calculator{Users: users, Products: products, resMap: make(map[string]map[string]*Debt)}
}

func (c Calculator) getProduct(id int64) Product {
	return c.Products[id]
}

/*
Можно сформировать историю. []HistoryItem

Каждый HistoryItem представляет собой информацию об одной трате
{
	kind: INCOME/OUTCOME;
	price: 1000;
	productTitle: "Пиво"
}


Result: map вида "Должник" -> ("Кредитор": []HistoryItem)


"Дима": {

}

*/

func (c Calculator) getDebt(username, creditorUsername string) *Debt {
	_, ok := c.resMap[username]
	if !ok {
		c.resMap[username] = make(map[string]*Debt)
	}

	_, ok = c.resMap[username][creditorUsername]
	if !ok {
		c.resMap[username][creditorUsername] = &Debt{
			TotalDebt: 0,
			History:   make([]HistoryItem, 0),
		}
	}

	return c.resMap[username][creditorUsername]
}

func (c Calculator) filterNoDebts() {
	for user, debtMap := range c.resMap {
		for creditor, debt := range debtMap {
			if debt.TotalDebt <= 0 {
				delete(debtMap, creditor)
			}
		}

		if len(debtMap) == 0 {
			delete(c.resMap, user)
		}
	}
}

func (c Calculator) Calculate() map[string]map[string]*Debt {
	for _, debtor := range c.Users {
		for _, expense := range debtor.Expenses {
			product := c.getProduct(expense.ProductID)
			newPrice := product.Price / product.NumOfEats

			debtorDebt := c.getDebt(debtor.Username, expense.BuyerUsername)
			debtorDebt.TotalDebt += newPrice
			debtorDebt.History = append(debtorDebt.History, HistoryItem{
				Kind:         OUTCOME,
				ProductTitle: product.Title,
				Price:        newPrice,
			})

			buyerDebt := c.getDebt(expense.BuyerUsername, debtor.Username)
			buyerDebt.TotalDebt -= newPrice
			buyerDebt.History = append(buyerDebt.History, HistoryItem{
				Kind:         INCOME,
				ProductTitle: product.Title,
				Price:        newPrice,
			})

		}
	}

	c.filterNoDebts()

	return c.resMap
}
