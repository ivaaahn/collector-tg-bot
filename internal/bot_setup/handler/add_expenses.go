package handler

import (
	"collector/internal"
	"collector/internal/bot_setup/fsm"
	"collector/internal/bot_setup/sessionctx"
	"collector/internal/flow"
	"errors"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"strconv"
	"strings"
)

type AddExpensesHandler struct {
	log            internal.Logger
	sessionService *flow.Service
	fsm            *fsm.FSM
}

func NewAddExpensesHandler(log internal.Logger, sessionService *flow.Service, fsm *fsm.FSM) *AddExpensesHandler {
	return &AddExpensesHandler{log: log, sessionService: sessionService, fsm: fsm}
}

func (h *AddExpensesHandler) Execute(c tele.Context) error {
	if !sessionctx.IsSessionExist(c) {
		return c.Send(SessionNotActiveMessage)
	}

	const prompt string = "<Номер продукта>[<Пользователь1> ..., @me] | [@all]"
	userID := c.Message().Sender.ID

	// Handle FSM for product and name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
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

		h.fsm.Upsert(userID, nil, "waiting_for_expense_info")
		c.Send(prompt)
		return c.Send(answer)
	} else {
		h.fsm.Clear(userID)
	}

	expenseParser := ExpenseParser{}
	parsed, err := expenseParser.Parse(c.Text())
	if err != nil {
		h.log.Warnf("AddExpenses err: %w", err)
		return c.Send(prompt)
	}

	addExpensesDto := make([]*flow.ExpenseDTO, 0)
	for _, item := range parsed {
		eatersUsernames := make([]string, 0)
		for _, eater := range item.Eaters {
			eatersUsernames = append(eatersUsernames, eater[1:])
		}
		addExpensesDto = append(addExpensesDto, &flow.ExpenseDTO{
			PurchaseID: item.ProductNo,
			Eaters:     eatersUsernames,
			Quantity:   1,
		})
	}

	err = h.sessionService.AddExpenses(
		flow.AddExpensesDTO{
			ChatID:         c.Chat().ID,
			AuthorID:       c.Sender().ID,
			AuthorUsername: c.Sender().Username,
			Expenses:       addExpensesDto,
		})
	if err != nil {
		h.log.Warnf("AddPurchases err: %w", err)
		return c.Send("Извини, технические проблемы :(")
	}

	return c.Send("Добавлена новая трата!")
}

type ExpenseParseRes struct {
	Eaters    []string
	ProductNo int64
}

type ExpenseParser struct {
	input string
}

func (p *ExpenseParser) Parse(input string) ([]ExpenseParseRes, error) {
	if len(input) == 0 {
		return []ExpenseParseRes{}, nil
	}

	result := make([]ExpenseParseRes, 0)
	for _, line := range strings.Split(strings.Trim(input, "\n"), "\n") {
		parsedLine, err := p.parseOne(line)
		if err != nil {
			return []ExpenseParseRes{}, err
		}

		result = append(result, parsedLine)
	}

	return result, nil
}

func (p *ExpenseParser) parseOne(item string) (ExpenseParseRes, error) {
	args := strings.Split(item, " ")
	if len(args) < 1 {
		return ExpenseParseRes{}, errors.New("<Номер продукта>[, <username1>, <username2>, ..., @me, @all]")
	}

	for i := 0; i < len(args); i++ {
		args[i] = strings.Trim(args[i], " ")
	}

	productNo, err := strconv.Atoi(args[0])
	if err != nil {
		return ExpenseParseRes{}, fmt.Errorf("productNo format error: %q", productNo)
	}

	usernames := make([]string, 0)

	if len(args) == 1 {
		usernames = append(usernames, "@me")
	} else if len(args) == 2 && args[1] == "@all" {
		usernames = append(usernames, "@all")
	} else {
		meExists := false
		for _, username := range args[1:] {
			usernames = append(usernames, username)

			if strings.Compare("@me", username) == 0 {
				if meExists {
					return ExpenseParseRes{}, fmt.Errorf("can't use @me more than 1 time: %q", productNo)
				} else {
					meExists = true
				}
			} else if strings.Compare("@all", username) == 0 {
				return ExpenseParseRes{}, fmt.Errorf("can't use @all with other usernames: %q", productNo)
			}
		}
	}

	return ExpenseParseRes{
		Eaters:    usernames,
		ProductNo: int64(productNo),
	}, nil
}
