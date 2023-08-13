package botmw

import (
	"collector/internal"
	"gopkg.in/telebot.v3"
)

func GetLogInputMsgMiddleware(logger internal.Logger) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			logger.Infof("Received message from %s, text = %s", c.Message().Sender.Username, c.Text())
			return next(c)
		}
	}
}
