package handler

import (
	"collector/internal"
	"collector/internal/bot_setup/fsm"
	"collector/internal/bot_setup/sessionctx"
	"collector/internal/flow"
	"fmt"
	tele "gopkg.in/telebot.v3"
)

type GetPurchasesHandler struct {
	log            internal.Logger
	sessionService *flow.Service
	fsm            *fsm.FSM
}

func NewGetPurchasesHandler(log internal.Logger, sessionService *flow.Service, fsm *fsm.FSM) *GetPurchasesHandler {
	return &GetPurchasesHandler{log: log, sessionService: sessionService, fsm: fsm}
}

func (h *GetPurchasesHandler) Execute(c tele.Context) error {
	if !sessionctx.IsSessionExist(c) {
		return c.Send(SessionNotActiveMessage)
	}

	purchases, err := h.sessionService.GetPurchases(c.Chat().ID)
	if err != nil {
		h.log.Warnf("GetPurchases err: %w", err)
		return c.Send("Произошла ошибка...")
	}

	var answer string

	const purchaseTemplateFmt string = "%d. %s\n"
	for _, purchase := range purchases {
		answer += fmt.Sprintf(purchaseTemplateFmt, purchase.ID, purchase.Title)
	}

	if len(answer) == 0 {
		return c.Send("Пока нет ни одной покупки")
	}

	return c.Send(answer)
}
