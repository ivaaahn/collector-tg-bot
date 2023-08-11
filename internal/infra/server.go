package infra

import (
	repo "collector/internal/fsm"
	"collector/internal/sessions"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

const (
	sessionsList     = "/sessions"
	newSession       = "/new_session"
	addPurchases     = "/add_purchases"
	collectPurchases = "/collect_purchases"
	addExpenses      = "/add_expenses"
	debts            = "/debts"
	costs            = "/costs"
	costsFull        = "/costs_full"
	finish           = "/finish"
)

type Server struct {
	db     *sql.DB
	logger *logrus.Entry
	token  string
	bot    *telebot.Bot
}

func NewServer(logger *logrus.Entry, token string, db *sql.DB, bot *telebot.Bot) *Server {
	return &Server{logger: logger, token: token, db: db, bot: bot}
}

func (s *Server) Run() {
	s.setupHandlers()
	s.setupCommands()

	s.logger.Infof("Server is running!")
	s.bot.Start()
}

func (s *Server) setupCommands() {
	commands := []telebot.Command{
		{
			Text:        addPurchases,
			Description: "Добавить трату",
		},
		{
			Text:        costs,
			Description: "Траты на текущий момент",
		},
		{
			Text:        costsFull,
			Description: "Подробные траты на текущий момент",
		},
		{
			Text:        collectPurchases,
			Description: "Получить список покупок",
		},
		{
			Text:        addExpenses,
			Description: "Добавить трату",
		},
		//{
		//	Text:        sessionsList,
		//	Product: "Список сессий",
		//},
		{
			Text:        newSession,
			Description: "Начать сессию",
		},
		{
			Text:        debts,
			Description: "Мои долги",
		},
		{
			Text:        finish,
			Description: "Завершить сессию",
		},
	}
	err := s.bot.SetCommands(commands)
	if err != nil {
		s.logger.Errorf("Can't set commands: %q", err)
	}

}

func (s *Server) setupHandlers() {
	fsmRepository := repo.NewFSMRepository(s.logger)

	sessionRepo := sessions.NewRepo(s.logger, s.db)
	sessionsUsecase := sessions.NewUsecase(s.logger, *sessionRepo)
	sessionHandler := sessions.NewHandler(s.logger, sessionsUsecase, fsmRepository)

	router := map[string]telebot.HandlerFunc{
		newSession:       sessionHandler.StartSession,
		addPurchases:     sessionHandler.AddPurchases,
		collectPurchases: sessionHandler.GetPurchases,
		addExpenses:      sessionHandler.AddExpenses,
		//costsFull: sessionHandler.GetCostsFull,
		//costs:     sessionHandler.GetCosts,
		debts: sessionHandler.GetDebts,
		//finish:    sessionHandler.FinishSession,
	}

	// Default command handlers
	group := s.bot.Group()
	group.Use(getInitFsmMiddleware(fsmRepository))
	for endpoint, handler := range router {
		group.Handle(endpoint, handler)
	}

	// Common text handler
	s.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID

		if fsmRepository.IsActiveSessionExists(userID) {
			userCtx := fsmRepository.GetContext(userID)
			handler, ok := router[userCtx.Endpoint]
			if ok {
				return handler(c)
			}
		}

		return nil
	})

	// Common middleware
	s.bot.Use(getLogInputMsgMiddleware(s.logger))
}
