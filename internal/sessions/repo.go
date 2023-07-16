package sessions

import "C"
import (
	"collector/internal"
	"database/sql"
)

type Repo struct {
	log  internal.Logger
	Conn *sql.DB
}

func NewRepo(log internal.Logger, conn *sql.DB) *Repo {
	return &Repo{log: log, Conn: conn}
}

func (r *Repo) InsertSession(session *Session) error {
	queryString := `
		INSERT INTO sessions (uuid, creator_id, chat_id, session_name, started_at, state) 
		VALUES ($1, $2, $3, $4, current_timestamp, $5);`

	_, err := r.Conn.Exec(queryString, session.UUID, session.CreatorID, session.ChatID,
		session.SessionName, session.State)
	return err
}

func (r *Repo) GetByChatID(chatID int64) (*Session, error) {
	var (
		session = NewEmptySession()
		err     error
	)
	query := `
	SELECT uuid, creator_id, chat_id, session_name, state
	FROM sessions WHERE chat_id = $1;`

	rows, err := r.Conn.Query(query, chatID)
	if err == nil {
		for rows.Next() {
			err = rows.Scan(&session.UUID, &session.CreatorID, &session.ChatID, &session.SessionName,
				&session.State)
		}
	}
	return session, err
}

func (r *Repo) AddMember(sessionUUID internal.UUID, userID int64) (int64, error) {
	var id int64
	query := `INSERT INTO members(session_id, user_id) VALUES ($1, $2) returning id;`
	row := r.Conn.QueryRow(query, sessionUUID, userID)
	err := row.Scan(&id)
	return id, err
}

func (r *Repo) GetMember(sessionUUID internal.UUID, userID int64) (*Member, error) {
	var (
		member = NewEmptyMember()
		err    error
	)
	query := `
	SELECT id, session_id, user_id
	FROM members 
	WHERE session_id = $1 AND user_id = $2;`

	rows, err := r.Conn.Query(query, sessionUUID, userID)
	if err == nil {
		for rows.Next() {
			err = rows.Scan(&member.ID, &member.SessionUUID, &member.UserID)
		}
	}
	return member, err
}

func (r *Repo) AddMemberCost(cost *Cost) error {
	queryString := `INSERT INTO costs (member_id, money, description) VALUES ($1, $2, $3);`
	_, err := r.Conn.Exec(queryString, cost.MemberID, cost.Money, cost.Description)
	return err
}

func (r *Repo) SelectUsersCosts(sessionUUID internal.UUID) ([]*UserCost, error) {
	result := make([]*UserCost, 0)

	queryString := `
	SELECT U.username, C.money, C.description, C.created_at
	FROM members as M 
	    JOIN costs C on M.id = C.member_id
		JOIN users as U on M.user_id = U.id
	WHERE M.session_id = $1;`

	rows, err := r.Conn.Query(queryString, sessionUUID)
	if err == nil {
		for rows.Next() {
			var tmpExpenses = &UserCost{}
			err = rows.Scan(&tmpExpenses.Username, &tmpExpenses.Amount, &tmpExpenses.Description, &tmpExpenses.CreatedAt)
			if err == nil {
				result = append(result, tmpExpenses)
			}
		}
	}

	return result, err
}

func (r *Repo) GetMemberCosts(sessionUUID internal.UUID) ([]*MemberCost, error) {
	result := make([]*MemberCost, 0)

	query := `
	SELECT M.user_id, C.money
	FROM members as M 
	    JOIN costs as C on M.id = C.member_id
	WHERE M.session_id = $1`

	rows, err := r.Conn.Query(query, sessionUUID)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var costs = &MemberCost{}
		err = rows.Scan(&costs.UserID, &costs.Money)
		if err != nil {
			return nil, err
		}
		result = append(result, costs)
	}

	return result, err
}

func (r *Repo) ChangeSessionStateToClosed(sessionUUID internal.UUID) error {
	query := `UPDATE sessions SET state = $1 WHERE uuid = $2`

	_, err := r.Conn.Exec(query, SessionClosedState, sessionUUID)
	return err
}
