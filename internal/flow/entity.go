package flow

import (
	"collector/internal"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        int64
	Username  string
	CreatedAt time.Time
}

type Session struct {
	ID         uuid.UUID
	Title      string
	CreatorID  int64
	ChatID     int64
	CreatedAt  time.Time
	FinishedAt sql.NullTime
}

type Member struct {
	UserID    int64
	SessionID internal.UUID
}

type Purchase struct {
	ID        int
	Title     string
	BuyerID   int64
	SessionID uuid.UUID
	Price     int
	CreatedAt time.Time
	Quantity  int
}

type Expense struct {
	PurchaseID int64
	EaterID    int64
	SessionID  uuid.UUID
	Quantity   int
}

func NewPurchase(sessionID uuid.UUID, buyerID int64, price int, title string, quantity int) *Purchase {
	return &Purchase{
		BuyerID:   buyerID,
		SessionID: sessionID,
		Price:     price,
		Title:     title,
		Quantity:  quantity,
	}
}

func NewExpense(purchaseID int64, eaterID int64, qty int, sessionID uuid.UUID) *Expense {
	return &Expense{
		PurchaseID: purchaseID,
		EaterID:    eaterID,
		SessionID:  sessionID,
		Quantity:   qty,
	}
}

func NewUser(ID int64, username string) *User {
	return &User{
		ID:       ID,
		Username: username,
	}
}

func NewMember(userID int64, sessionID uuid.UUID) *Member {
	return &Member{
		UserID:    userID,
		SessionID: sessionID,
	}
}
