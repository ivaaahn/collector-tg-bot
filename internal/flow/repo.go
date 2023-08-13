package flow

import (
	"collector/internal"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/sqlscan"
)

var SessionNotFoundDBErr = errors.New("session not found error")
var UserNotFoundDBErr = errors.New("user not found error")
var MemberNotFoundDBErr = errors.New("member not found error")

type Repo struct {
	log  internal.Logger
	Conn *sql.DB
}

func NewRepo(log internal.Logger, conn *sql.DB) *Repo {
	return &Repo{log: log, Conn: conn}
}

func (r *Repo) UpsertUser(newUser *User) error {
	_, err := r.GetUserById(newUser.ID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, UserNotFoundDBErr) {
		return fmt.Errorf("Repo->UpsertUser: %w", err)
	}

	if err = r.CreateUser(newUser); err != nil {
		return fmt.Errorf("Repo->UpsertUser: %w", err)
	}

	return nil
}

func (r *Repo) UpsertMember(newMember *Member) error {
	_, err := r.GetMemberByUserId(newMember.SessionID, newMember.UserID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, MemberNotFoundDBErr) {
		return fmt.Errorf("Repo->UpsertMember: %w", err)
	}

	if err = r.AddMember(newMember); err != nil {
		return fmt.Errorf("Repo->UpsertMember: %w", err)
	}

	return nil
}
func (r *Repo) GetUserById(ID int64) (*User, error) {
	q := `SELECT id, username FROM users WHERE id = $1;`

	var user User
	err := sqlscan.Get(context.Background(), r.Conn, &user, q, ID)
	if err == nil {
		return &user, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, UserNotFoundDBErr
	}

	return nil, fmt.Errorf("Repo->GetUserById: %w", err)
}

func (r *Repo) GetUserByUsername(username string) (*User, error) {
	q := `SELECT id, username, created_at FROM users WHERE username = $1;`

	var user User
	err := sqlscan.Get(context.Background(), r.Conn, &user, q, username)
	if err == nil {
		return &user, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, UserNotFoundDBErr
	}

	return nil, fmt.Errorf("Repo->GetUserByUsername: %w", err)
}

func (r *Repo) CreateUser(user *User) error {
	q := `
			INSERT INTO users (id, username, created_at)
			VALUES ($1, $2, current_timestamp)
		`

	_, err := r.Conn.Exec(q, user.ID, user.Username)
	if err != nil {
		return fmt.Errorf("Repo->CreateUser: %w", err)
	}

	return nil
}

func (r *Repo) GetSessionByChatID(chatID int64) (*Session, error) {
	var session Session

	q := `
		SELECT id, title, creator_id, chat_id, created_at, finished_at
		FROM sessions WHERE chat_id = $1 and finished_at is NULL;`

	err := sqlscan.Get(context.Background(), r.Conn, &session, q, chatID)

	if err == nil {
		return &session, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, SessionNotFoundDBErr
	}

	return nil, fmt.Errorf("Repo->GetSessionByChatID: %w", err)
}

func (r *Repo) AddMember(member *Member) error {
	q := `INSERT INTO members(session_id, user_id) VALUES ($1, $2)`

	_, err := r.Conn.Exec(q, member.SessionID, member.UserID)
	if err != nil {
		return fmt.Errorf("Repo->AddMember: %w", err)
	}

	return nil
}

func (r *Repo) GetMemberByUserId(sessionID internal.UUID, userID int64) (*Member, error) {
	q := ` SELECT session_id, user_id FROM members WHERE session_id = $1 AND user_id = $2;`

	var member Member
	err := sqlscan.Get(context.Background(), r.Conn, &member, q, sessionID, userID)
	if err == nil {
		return &member, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, MemberNotFoundDBErr
	}

	return nil, err
}

func (r *Repo) GetMembers(sessionID internal.UUID) ([]*Member, error) {
	q := `SELECT session_id, user_id FROM members WHERE session_id = $1;`

	var members []*Member
	err := sqlscan.Select(context.Background(), r.Conn, &members, q, sessionID)
	if err != nil {
		return nil, err
	}

	return members, err
}

func (r *Repo) GetUsersBySessionID(sessionID internal.UUID) ([]*User, error) {
	q := `
SELECT id, username, created_at
FROM users 
    JOIN members on users.id = members.user_id
WHERE session_id = $1;`

	var users []*User
	err := sqlscan.Select(context.Background(), r.Conn, &users, q, sessionID)
	if err != nil {
		return nil, err
	}

	return users, err
}

func (r *Repo) GetPurchases(sessionID internal.UUID) ([]*Purchase, error) {
	q := `
SELECT id, buyer_id, session_id, price, created_at, title, quantity 
FROM purchases 
WHERE session_id = $1;`

	var purchases []*Purchase
	err := sqlscan.Select(context.Background(), r.Conn, &purchases, q, sessionID)
	if err != nil {
		return nil, err
	}

	return purchases, err
}

func (r *Repo) AddPurchase(purchase *Purchase) error {
	q := `INSERT INTO purchases (buyer_id, session_id, price, title, quantity) VALUES ($1, $2, $3, $4, $5);`
	_, err := r.Conn.Exec(q, purchase.BuyerID, purchase.SessionID, purchase.Price, purchase.Title, purchase.Quantity)
	return err
}

func (r *Repo) AddExpense(expense *Expense) error {
	q := `INSERT INTO expenses (purchase_id, eater_id, session_id, qty) VALUES ($1, $2, $3, $4);`
	_, err := r.Conn.Exec(q, expense.PurchaseID, expense.EaterID, expense.SessionID, expense.Quantity)
	return err
}
