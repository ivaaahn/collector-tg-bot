package infra

import (
	"collector/internal"
	repo "collector/internal/repository"
	"gopkg.in/telebot.v3"
)

func getInitFsmMiddleware(fsmRepository *repo.FSMRepository) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			fsmRepository.Init(c.Message().Sender.ID, c.Message().Text)
			return next(c)
		}
	}
}

func getLogInputMsgMiddleware(logger internal.Logger) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			logger.Infof("Received message from %s, text = %s", c.Message().Sender.Username, c.Text())
			return next(c)
		}
	}
}
