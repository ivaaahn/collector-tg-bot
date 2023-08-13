package sessionctx

import (
	"collector/internal"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"gopkg.in/telebot.v3"
)

type SessionDefault struct {
	ID         uuid.UUID
	FinishedAt sql.NullTime
}

var SessionNotExistErr = errors.New("session not exist")

const SessionIDKey = "session_id"

func GetSessionIDFromCtx(c telebot.Context) (uuid.UUID, error) {
	sessionIdRaw := c.Get(SessionIDKey)
	if sessionIdRaw == -1 {
		return uuid.New(), SessionNotExistErr
	}

	sessionID, ok := sessionIdRaw.(uuid.UUID)
	if !ok {
		return uuid.New(), SessionNotExistErr
	}

	return sessionID, nil
}

func IsSessionExist(c telebot.Context) bool {
	id, err := GetSessionIDFromCtx(c)
	fmt.Printf("%s", id)
	return err == nil
}

func GetCheckSessionMiddleware(conn *sql.DB, log internal.Logger) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			q := `
SELECT id, finished_at
FROM sessions 
WHERE chat_id = $1 and finished_at is NULL;`

			var activeSession SessionDefault
			if err := sqlscan.Get(context.Background(), conn, &activeSession, q, c.Chat().ID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.Set(SessionIDKey, -1)
					return next(c)
				}

				log.Warnf("GetCheckSessionMiddleware: %w", err)
				return c.Send("Извини, технические проблемы")
			}

			c.Set(SessionIDKey, activeSession.ID)
			return next(c)
		}
	}
}
