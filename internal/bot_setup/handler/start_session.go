package handler

import (
	"collector/internal"
	"collector/internal/bot_setup/fsm"
	"collector/internal/bot_setup/sessionctx"
	"collector/internal/session_manage"
	"fmt"
	tele "gopkg.in/telebot.v3"
)

type StartSessionHandler struct {
	log            internal.Logger
	sessionService *session_manage.Service
	fsm            *fsm.FSM
}

func NewStartSessionHandler(log internal.Logger, sessionService *session_manage.Service, fsm *fsm.FSM) *StartSessionHandler {
	return &StartSessionHandler{log: log, sessionService: sessionService, fsm: fsm}
}

func (h *StartSessionHandler) Execute(c tele.Context) error {
	if sessionctx.IsSessionExist(c) {
		return c.Send(SessionActiveMessage)
	}

	userID := c.Message().Sender.ID
	chatID := c.Chat().ID

	// Handle FSM for session name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
		h.fsm.Upsert(userID, nil, "waiting_for_title")
		return c.Send("Введите название сессии")
	} else {
		h.fsm.Clear(userID)
	}

	info := session_manage.CreateSessionDTO{
		UserID:      userID,
		ChatID:      chatID,
		Username:    c.Message().Sender.Username,
		SessionName: c.Text(),
	}

	if err := h.sessionService.CreateSession(info); err != nil {
		h.log.Warnf("StartSessionHandler->Execute: %w", err)
		return c.Send("Извини, технические проблемы")
	}

	h.log.Infof("User %s was added to session %s", info.Username, info.SessionName)
	return c.Send(fmt.Sprintf("Сессия '%s' успешно создана!", info.SessionName))
}
