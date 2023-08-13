package bot_setup

import (
	"collector/internal/bot_setup/botmw"
	"collector/internal/bot_setup/fsm"
	"collector/internal/bot_setup/handler"
	"collector/internal/bot_setup/sessionctx"
	"collector/internal/debts_calculation"
	"collector/internal/flow"
	"collector/internal/session_manage"
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
	sessionManageRepo := session_manage.NewRepo(s.logger, s.db)
	sessionManageSvc := session_manage.NewService(s.logger, *sessionManageRepo)

	sessionFlowRepo := flow.NewRepo(s.logger, s.db)
	sessionFlowSvc := flow.NewService(s.logger, *sessionFlowRepo)

	debtsRepo := debts_calculation.NewRepo(s.logger, s.db)
	debtsService := debts_calculation.NewService(s.logger, *debtsRepo)

	fsmRepository := fsm.NewFSM(s.logger)
	router := map[string]telebot.HandlerFunc{
		newSession: handler.NewStartSessionHandler(s.logger, sessionManageSvc, fsmRepository).Execute,
		finish:     handler.NewFinishSessionHandler(s.logger, sessionManageSvc, fsmRepository).Execute,

		addPurchases:     handler.NewAddPurchasesHandler(s.logger, sessionFlowSvc, fsmRepository).Execute,
		collectPurchases: handler.NewGetPurchasesHandler(s.logger, sessionFlowSvc, fsmRepository).Execute,
		addExpenses:      handler.NewAddExpensesHandler(s.logger, sessionFlowSvc, fsmRepository).Execute,

		debts: handler.NewGetDebtsHandler(s.logger, debtsService, fsmRepository).Execute,
	}

	// Default command handlers
	s.bot.Use(sessionctx.GetCheckSessionMiddleware(s.db, s.logger))

	group := s.bot.Group()
	group.Use(botmw.GetInitFsmMiddleware(fsmRepository))
	for endpoint, handler_ := range router {
		group.Handle(endpoint, handler_)
	}

	// Common text handler
	s.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID

		if fsmRepository.IsActiveSessionExists(userID) {
			userCtx := fsmRepository.GetContext(userID)
			handler_, ok := router[userCtx.Endpoint]
			if ok {
				return handler_(c)
			}
		}

		return nil
	})

	// Common middleware
	s.bot.Use(botmw.GetLogInputMsgMiddleware(s.logger))
}
