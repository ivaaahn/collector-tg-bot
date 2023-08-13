package botmw

import (
	"collector/internal/bot_setup/fsm"
	"gopkg.in/telebot.v3"
	"strings"
)

func GetInitFsmMiddleware(fsmRepository *fsm.FSM) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			fsmRepository.Init(c.Message().Sender.ID, strings.Split(c.Message().Text, "@")[0])
			return next(c)
		}
	}
}
