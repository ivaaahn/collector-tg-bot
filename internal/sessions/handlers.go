package sessions

import (
	"collector/internal"
	repo "collector/internal/repository"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"sort"
	"strconv"
	"strings"
)

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

func (h *Handler) AddCost(c tele.Context) error {
	userID := c.Message().Sender.ID

	// Handle FSM for product and name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
		h.fsm.Upsert(userID, nil, "waiting_for_info")
		return c.Send("Формат: <Цена> <Описание>")
	} else {
		h.fsm.Clear(userID)
	}

	args := strings.Split(c.Text(), " ")
	if len(args) < 2 {
		return c.Send("<Цена> <Описание>")
	}

	cost, err := strconv.Atoi(args[0])
	if err != nil {
		return c.Send("Цена должна быть целым числом!")
	}

	info := AddExpenseDTO{
		ChatID:   c.Chat().ID,
		Product:  strings.Join(args[1:], " "),
		Cost:     cost,
		UserID:   c.Message().Sender.ID,
		Username: c.Message().Sender.Username,
	}
	err = h.usecase.AddCost(info)
	if err != nil {
		h.log.Warnf("Add expense err: %w", err)
		return c.Send("Извини, технические проблемы :(")
	}

	return c.Send("Добавлена новая трата!")
}

func (h *Handler) GetCostsFull(c tele.Context) error {
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	allCosts, err := h.usecase.CollectCosts(c.Chat().ID)
	if err != nil {
		h.log.Warnf("Handler->GetCostsFull: %w", err)
		return c.Send("Извини, техническая ошибка :(")
	}

	if len(allCosts) == 0 {
		return c.Send("Трат пока еще не было :(")
	}

	responseText := "<b>Все траты на текущий момент</b>"
	for username, allUserCosts := range allCosts {
		responseText += "\n\n"
		responseText += fmt.Sprintf("<u>@%s – %d руб.</u>\n\n", username, allUserCosts.Sum)

		// Sorting for pretty output
		sort.Slice(allUserCosts.Costs, func(i, j int) bool {
			return allUserCosts.Costs[i].Amount > allUserCosts.Costs[j].Amount
		})

		for _, cost := range allUserCosts.Costs {
			responseText += fmt.Sprintf("<i>%s – %d руб (%s).</i>\n", cost.Description, cost.Amount, cost.CreatedAt.Local().Format("2 Jan, 15:02"))
		}
	}

	return c.Send(responseText, &tele.SendOptions{ParseMode: tele.ModeHTML})
}

func (h *Handler) GetCosts(c tele.Context) error {
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	allCosts, err := h.usecase.CollectCosts(c.Chat().ID)
	if err != nil {
		h.log.Warnf("Handler->GetCostsFull: %w", err)
		return c.Send("Извини, техническая ошибка :(")
	}

	if len(allCosts) == 0 {
		return c.Send("Трат пока еще не было :(")
	}

	responseText := "<b>Все траты на текущий момент</b>\n"
	for username, allUserCosts := range allCosts {
		responseText += fmt.Sprintf("@%s – %d руб.\n", username, allUserCosts.Sum)
	}

	return c.Send(responseText, &tele.SendOptions{ParseMode: tele.ModeHTML})
}

func (h *Handler) GetDebts(c tele.Context) error {
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	debts, err := h.usecase.GetAllDebts(c.Chat().ID)
	if err != nil {
		h.log.Warnf("Handler->GetAllDebts: %w", err)
		return c.Send("Извини, техническая ошибка :(")
	}

	if len(debts) == 0 {
		return c.Send("Долгов нет")
	}

	responseText := "<b>Все долги на текущий момент</b>\n"
	for username, userDebts := range debts {
		responseText += "\n\n"
		responseText += fmt.Sprintf("<u>@%s – %d руб.</u>\n\n", username, userDebts)

		// Sorting for pretty output
		sort.Slice(userDebts.Debts, func(i, j int) bool {
			return userDebts.Debts[i].Money > userDebts.Debts[j].Money
		})

		for _, cost := range userDebts.Debts {
			responseText += fmt.Sprintf("<i>%s – %d руб.</i>\n", cost.DebtorName, cost.Money)
		}
	}

	return c.Send(responseText, &tele.SendOptions{ParseMode: tele.ModeHTML})
}

func (h *Handler) checkSession(c tele.Context, exists bool) error {
	isSessionExists, err := h.usecase.IsSessionExists(c.Chat().ID)
	if err != nil {
		h.log.Warnf("StartSessionHandler->Execute: %w", err)
		return c.Send("Извини, технические проблемы")
	}
	if !exists && isSessionExists {
		return c.Send("Для старта новой сессии завершите текущую")
	}

	if exists && !isSessionExists {
		return c.Send("Для выполнения данной команды начните сессию")
	}

	return nil
}

func (h *Handler) FinishSession(c tele.Context) error {
	if err := h.checkSession(c, true); err != nil {
		return err
	}

	chatID := c.Chat().ID

	costs, err := h.usecase.CollectCosts(chatID)
	if err != nil {
		h.log.Warnf("Finish session: %w", err)
		return c.Send("Извини, технические проблемы")
	}

	if err = h.usecase.FinishSession(chatID); err != nil {
		h.log.Warnf("Finish session: %w", err)
		return c.Send("Извини, технические проблемы")
	}

	responseText := "Сессия завершена! Итоговые траты: \n"
	for username, userCosts := range costs {
		responseText += "\n\n"
		responseText += fmt.Sprintf("Пользователь @%s \n", username)
		responseText += fmt.Sprintf("Общая сумма: %d рублей\n\n", userCosts.Sum)

		// Sorting for pretty output
		sort.Slice(userCosts.Costs, func(i, j int) bool {
			return userCosts.Costs[i].Amount > userCosts.Costs[j].Amount
		})

		for _, cost := range userCosts.Costs {
			responseText += fmt.Sprintf("%s - %d рублей \n", cost.Description, cost.Amount)
		}
	}

	return c.Send(responseText)
}
