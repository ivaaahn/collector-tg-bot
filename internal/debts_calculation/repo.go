package debts_calculation

import (
	"collector/internal"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/sqlscan"
)

var SessionNotFoundDBErr = errors.New("session not found error")

type Repo struct {
	log  internal.Logger
	Conn *sql.DB
}

func NewRepo(log internal.Logger, conn *sql.DB) *Repo {
	return &Repo{log: log, Conn: conn}
}

func (r *Repo) GetSessionByChatID(chatID int64) (*Session, error) {
	q := `SELECT id FROM sessions WHERE chat_id = $1 and finished_at is NULL;`

	var session Session
	err := sqlscan.Get(context.Background(), r.Conn, &session, q, chatID)

	if err == nil {
		return &session, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, SessionNotFoundDBErr
	}

	return nil, fmt.Errorf("Repo->GetSessionByChatID: %w", err)
}

func (r *Repo) GetExpenses(sessionID internal.UUID) ([]*Expense, error) {
	q := `
SELECT eu.id as eater_id, 
    	eu.username as eater_username, 
    	purchase_id as product_id, 
    	bu.username as buyer_username,
    	bu.id as buyer_id
FROM expenses e
	JOIN users eu on e.eater_id = eu.id
	JOIN purchases p on e.purchase_id = p.id
	JOIN users bu on p.buyer_id = bu.id
WHERE e.session_id = $1;`

	var expenses []*Expense
	err := sqlscan.Select(context.Background(), r.Conn, &expenses, q, sessionID)
	if err != nil {
		return nil, err
	}

	return expenses, err
}

func (r *Repo) GetProducts(sessionID internal.UUID) ([]*Product, error) {
	q := `
SELECT id, title, price, count(p.id) as num_of_eats
FROM purchases p
	JOIN expenses e on p.id = e.purchase_id
WHERE p.session_id = $1
GROUP BY e.purchase_id, p.id, title, price;`

	var products []*Product
	err := sqlscan.Select(context.Background(), r.Conn, &products, q, sessionID)
	if err != nil {
		return nil, err
	}

	return products, err
}
