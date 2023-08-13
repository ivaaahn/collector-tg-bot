package handler

import (
	"collector/internal"
	"collector/internal/bot_setup/fsm"
	"collector/internal/bot_setup/sessionctx"
	"collector/internal/debts_calculation"
	"fmt"
	tele "gopkg.in/telebot.v3"
)

type GetDebtsHandler struct {
	log              internal.Logger
	debtsCalcService *debts_calculation.Service
	fsm              *fsm.FSM
}

func NewGetDebtsHandler(log internal.Logger, debtsCalcService *debts_calculation.Service, fsm *fsm.FSM) *GetDebtsHandler {
	return &GetDebtsHandler{log: log, debtsCalcService: debtsCalcService, fsm: fsm}
}

func (h *GetDebtsHandler) Execute(c tele.Context) error {
	if !sessionctx.IsSessionExist(c) {
		return c.Send(SessionNotActiveMessage)
	}

	res, err := h.debtsCalcService.GetAllDebts(c.Chat().ID)
	if err != nil {
		h.log.Warnf("Handler->GetAllDebts: %w", err)
		return c.Send("Извини, техническая ошибка :(")
	}

	if len(res) == 0 {
		return c.Send("Долгов нет")
	}

	signMap := map[debts_calculation.PaymentKind]string{
		debts_calculation.INCOME:  "+",
		debts_calculation.OUTCOME: "–",
	}

	responseText := "<b>Все долги на текущий момент</b>\n\n"
	for username, userCreditors := range res {
		for creditor, debt := range userCreditors {
			responseText += fmt.Sprintf("@%s ➡️ @%s – %d руб.\n", username, creditor, debt.TotalDebt)
			for _, historyItem := range debt.History {
				responseText += fmt.Sprintf(
					"<i>%s%d руб. – %s</i>\n",
					signMap[historyItem.Kind],
					historyItem.Price,
					historyItem.ProductTitle,
				)
			}
		}
		responseText += "\n\n"
	}

	return c.Send(responseText, &tele.SendOptions{ParseMode: tele.ModeHTML})
}
