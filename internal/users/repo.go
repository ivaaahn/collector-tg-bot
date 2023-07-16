package users

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

func (r Repo) GetById(ID int64) (*User, error) {
	var (
		user = NewUser()
		err  error
	)
	queryString := `SELECT id, tg_id, username, requisites FROM users WHERE id = $1;`

	rows, err := r.Conn.Query(queryString, ID)
	if err == nil {
		for rows.Next() {
			err = rows.Scan(&user.ID, &user.TgID, &user.Username, &user.CreatedAt, &user.Requisites)
		}
	}
	return user, err
}

func (r Repo) GetBySessionID(sessionUUID internal.UUID) ([]*User, error) {
	result := make([]*User, 0)

	query := `
	select U.id, U.tg_id, U.username, U.created_at, U.requisites 
	from users as U 
	    join members as M on U.id = M.user_id 
	where M.session_id = $1`

	rows, err := r.Conn.Query(query, sessionUUID)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user = &User{}
		err = rows.Scan(&user.ID, &user.TgID, &user.Username, &user.CreatedAt, &user.Requisites)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, err
}

func (r Repo) Create(user *User) (int64, error) {
	query := `
		INSERT INTO users (tg_id, username, created_at, requisites) 
		VALUES ($1, $2, current_timestamp, $3) 
		returning id;`

	var id int64
	row := r.Conn.QueryRow(query, user.TgID, user.Username, user.Requisites)
	err := row.Scan(&id)
	return id, err
}
