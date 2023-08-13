package handler

import (
	"collector/internal"
	"collector/internal/bot_setup/fsm"
	"collector/internal/bot_setup/sessionctx"
	"collector/internal/session_manage"
	tele "gopkg.in/telebot.v3"
)

type FinishSessionHandler struct {
	log            internal.Logger
	sessionService *session_manage.Service
	fsm            *fsm.FSM
}

func NewFinishSessionHandler(log internal.Logger, sessionService *session_manage.Service, fsm *fsm.FSM) *FinishSessionHandler {
	return &FinishSessionHandler{log: log, sessionService: sessionService, fsm: fsm}
}

func (h *FinishSessionHandler) Execute(c tele.Context) error {
	if !sessionctx.IsSessionExist(c) {
		return c.Send(SessionNotActiveMessage)
	}

	sessionID, _ := sessionctx.GetSessionIDFromCtx(c)
	if err := h.sessionService.FinishSession(sessionID); err != nil {
		h.log.Warnf("Finish session: %w", err)
		return c.Send("Извини, технические проблемы")
	}

	return c.Send("Сессия успешно завершена")
}
