package session_manage

type CreateSessionDTO struct {
	UserID      int64
	ChatID      int64
	Username    string
	SessionName string
}
