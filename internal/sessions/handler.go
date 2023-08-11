package sessions

import (
	"collector/internal"
	"collector/internal/calculator"
	repo "collector/internal/fsm"
	"collector/internal/parser"
	"errors"
	"fmt"
	tele "gopkg.in/telebot.v3"
)

var sessionErr error = errors.New("session error")

type Handler struct {
	log     internal.Logger
	usecase *SessionUsecase
	fsm     *repo.FSMRepository
}

func NewHandler(log internal.Logger, usecase *SessionUsecase, fsm *repo.FSMRepository) *Handler {
	return &Handler{log: log, usecase: usecase, fsm: fsm}
}

func (h *Handler) StartSession(c tele.Context) error {
	userID := c.Message().Sender.ID
	chatID := c.Chat().ID

	if err := h.checkSession(c, false); err != nil {
		return err
	}

	// Handle FSM for session name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
		h.fsm.Upsert(userID, nil, "waiting_for_title")
		return c.Send("Введите название сессии")
	} else {
		h.fsm.Clear(userID)
	}

	info := CreateSessionDTO{
		UserID:      userID,
		ChatID:      chatID,
		Username:    c.Message().Sender.Username,
		SessionName: c.Text(),
	}

	if err := h.usecase.CreateSession(info); err != nil {
		h.log.Warnf("StartSessionHandler->Execute: %w", err)
		return c.Send("Извини, технические проблемы")
	}

	h.log.Infof("User %s was added to session %s", info.Username, info.SessionName)
	return c.Send(fmt.Sprintf("Сессия '%s' успешно создана!", info.SessionName))
}

func (h *Handler) GetPurchases(c tele.Context) error {
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	purchases, err := h.usecase.GetPurchases(c.Chat().ID)
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

func (h *Handler) AddPurchases(c tele.Context) error {
	const prompt string = "<Название>, <Цена>[, <Кол-во>]"
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	userID := c.Message().Sender.ID

	// Handle FSM for product and name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
		h.fsm.Upsert(userID, nil, "waiting_for_purchase_info")
		return c.Send(prompt)
	} else {
		h.fsm.Clear(userID)
	}

	parsed, err := parser.NewPurchaseParser().Parse(c.Text())
	if err != nil {
		h.log.Warnf("AddPurchases err: %w", err)
		return c.Send(prompt)
	}

	purchases := make([]*PurchaseDTO, 0)
	for _, item := range parsed {
		purchases = append(purchases, &PurchaseDTO{
			ProductTitle: item.ProductTitle,
			Price:        item.Price,
			Quantity:     item.Quantity,
		})
	}

	addPurchasesDto := AddPurchasesDTO{
		ChatID:        c.Chat().ID,
		BuyerID:       c.Message().Sender.ID,
		BuyerUsername: c.Message().Sender.Username,
		Purchases:     purchases,
	}

	if err = h.usecase.AddPurchases(addPurchasesDto); err != nil {
		h.log.Warnf("AddPurchases err: %w", err)
		return c.Send("Извини, технические проблемы :(")
	}

	return c.Send("Добавлены новые покупки")
}

func (h *Handler) AddExpenses(c tele.Context) error {
	const prompt string = "<Номер продукта>[<Пользователь1> ..., @me] | [@all]"
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	userID := c.Message().Sender.ID

	// Handle FSM for product and name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
		h.fsm.Upsert(userID, nil, "waiting_for_expense_info")
		c.Send(prompt)
		return h.GetPurchases(c)
	} else {
		h.fsm.Clear(userID)
	}

	parsed, err := parser.NewExpenseParser().Parse(c.Text())
	if err != nil {
		h.log.Warnf("AddExpenses err: %w", err)
		return c.Send(prompt)
	}

	addExpensesDto := make([]*ExpenseDTO, 0)
	for _, item := range parsed {
		eatersUsernames := make([]string, 0)
		for _, eater := range item.Eaters {
			eatersUsernames = append(eatersUsernames, eater[1:])
		}
		addExpensesDto = append(addExpensesDto, &ExpenseDTO{
			PurchaseID: item.ProductNo,
			Eaters:     eatersUsernames,
			Quantity:   1,
		})
	}

	err = h.usecase.AddExpenses(
		AddExpensesDTO{
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

func (h *Handler) GetDebts(c tele.Context) error {
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	res, err := h.usecase.GetAllDebts(c.Chat().ID)
	if err != nil {
		h.log.Warnf("Handler->GetAllDebts: %w", err)
		return c.Send("Извини, техническая ошибка :(")
	}

	if len(res) == 0 {
		return c.Send("Долгов нет")
	}

	signMap := map[calculator.PaymentKind]string{
		calculator.INCOME:  "+",
		calculator.OUTCOME: "–",
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

func (h *Handler) checkSession(c tele.Context, exists bool) error {
	isSessionExists, err := h.usecase.IsSessionExists(c.Chat().ID)
	if err != nil {
		h.log.Warnf("StartSessionHandler->Execute: %w", err)
		if err = c.Send("Извини, технические проблемы"); err != nil {
			return err
		}
		return sessionErr
	}
	if !exists && isSessionExists {
		if err = c.Send("Для старта новой сессии завершите текущую"); err != nil {
			return err
		}
		return sessionErr
	}

	if exists && !isSessionExists {
		if err = c.Send("Для выполнения данной команды начните сессию"); err != nil {
			return err
		}
		return sessionErr
	}

	return nil
}

//func (h *Handler) FinishSession(c tele.Context) error {
//	if err := h.checkSession(c, true); err != nil {
//		return err
//	}
//
//	chatID := c.Chat().ID
//
//	costs, err := h.usecase.CollectPurchases(chatID)
//	if err != nil {
//		h.log.Warnf("Finish session: %w", err)
//		return c.Send("Извини, технические проблемы")
//	}
//
//	if err = h.usecase.FinishSession(chatID); err != nil {
//		h.log.Warnf("Finish session: %w", err)
//		return c.Send("Извини, технические проблемы")
//	}
//
//	responseText := "Сессия завершена! Итоговые траты: \n"
//	for username, userCosts := range costs {
//		responseText += "\n\n"
//		responseText += fmt.Sprintf("Пользователь @%s \n", username)
//		responseText += fmt.Sprintf("Общая сумма: %d рублей\n\n", userCosts.Sum)
//
//		// Sorting for pretty output
//		sort.Slice(userCosts.Eaters, func(i, j int) bool {
//			return userCosts.Eaters[i].TotalDebt > userCosts.Eaters[j].TotalDebt
//		})
//
//		for _, cost := range userCosts.Eaters {
//			responseText += fmt.Sprintf("%s - %d рублей \n", cost.Product, cost.TotalDebt)
//		}
//	}
//
//	return c.Send(responseText)
//}
