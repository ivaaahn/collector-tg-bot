package session_manage

import (
	"collector/internal"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
)

var UserNotFoundDBErr = errors.New("user not found error")
var SessionNotFoundErr = errors.New("session not found error")

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

func (r *Repo) CreateSession(session *Session) error {
	q := `
	INSERT INTO sessions (id, creator_id, chat_id, title, created_at)
	VALUES ($1, $2, $3, $4, current_timestamp);`

	_, err := r.Conn.Exec(
		q, session.ID, session.CreatorID, session.ChatID, session.Title,
	)
	if err != nil {
		return fmt.Errorf("Repo->CreateSession: %w", err)
	}

	return nil
}

func (r *Repo) GetSession(ID uuid.UUID) (*Session, error) {
	q := `SELECT id, finished_at FROM sessions WHERE id = $1`

	var session Session
	if err := sqlscan.Get(context.Background(), r.Conn, &session, q, ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, SessionNotFoundErr
		}
		return nil, fmt.Errorf("Repo->GetSession: %w", err)
	}

	return &session, nil
}

func (r *Repo) SetFinishTimeToNow(sessionID uuid.UUID) error {
	q := `
	UPDATE sessions 
		SET finished_at = now()
	WHERE id = $1;`

	_, err := r.Conn.Exec(q, sessionID)
	if err != nil {
		return fmt.Errorf("Repo->SetFinishTimeToNow: %w", err)
	}

	return nil
}

func (r *Repo) AddMember(member *Member) error {
	q := `INSERT INTO members(session_id, user_id) VALUES ($1, $2)`

	_, err := r.Conn.Exec(q, member.SessionID, member.UserID)
	if err != nil {
		return fmt.Errorf("Repo->AddMember: %w", err)
	}

	return nil
}
