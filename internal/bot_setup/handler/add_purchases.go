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

type AddPurchasesHandler struct {
	log            internal.Logger
	sessionService *flow.Service
	fsm            *fsm.FSM
}

func NewAddPurchasesHandler(log internal.Logger, sessionService *flow.Service, fsm *fsm.FSM) *AddPurchasesHandler {
	return &AddPurchasesHandler{log: log, sessionService: sessionService, fsm: fsm}
}

func (h *AddPurchasesHandler) Execute(c tele.Context) error {
	if !sessionctx.IsSessionExist(c) {
		return c.Send(SessionNotActiveMessage)
	}

	const prompt string = "<Название>, <Цена>[, <Кол-во>]"
	userID := c.Message().Sender.ID

	// Handle FSM for product and name
	userCtx := h.fsm.GetContext(userID)
	if userCtx.Stage == "" {
		h.fsm.Upsert(userID, nil, "waiting_for_purchase_info")
		return c.Send(prompt)
	} else {
		h.fsm.Clear(userID)
	}

	parser := PurchaseParser{}
	parsed, err := parser.Parse(c.Text())
	if err != nil {
		h.log.Warnf("AddPurchases err: %w", err)
		return c.Send(prompt)
	}

	purchases := make([]*flow.PurchaseDTO, 0)
	for _, item := range parsed {
		purchases = append(purchases, &flow.PurchaseDTO{
			ProductTitle: item.ProductTitle,
			Price:        item.Price,
			Quantity:     item.Quantity,
		})
	}

	addPurchasesDto := flow.AddPurchasesDTO{
		ChatID:        c.Chat().ID,
		BuyerID:       c.Message().Sender.ID,
		BuyerUsername: c.Message().Sender.Username,
		Purchases:     purchases,
	}

	if err = h.sessionService.AddPurchases(addPurchasesDto); err != nil {
		h.log.Warnf("AddPurchases err: %w", err)
		return c.Send("Извини, технические проблемы :(")
	}

	return c.Send("Добавлены новые покупки")
}

type PurchaseParseRes struct {
	Quantity     int
	Price        int
	ProductTitle string
}

type PurchaseParser struct {
	input string
}

func (p *PurchaseParser) Parse(input string) ([]PurchaseParseRes, error) {
	if len(input) == 0 {
		return []PurchaseParseRes{}, nil
	}

	result := make([]PurchaseParseRes, 0)
	for _, line := range strings.Split(strings.Trim(input, "\n"), "\n") {
		parsedLine, err := p.parseOne(line)
		if err != nil {
			return []PurchaseParseRes{}, err
		}

		result = append(result, parsedLine)
	}

	return result, nil
}

func (p *PurchaseParser) parseOne(item string) (PurchaseParseRes, error) {
	args := strings.Split(item, ",")
	if len(args) < 2 {
		return PurchaseParseRes{}, errors.New("<Название>, <Цена>[, <Кол-во> - default=1]")
	}

	for i := 0; i < len(args); i++ {
		args[i] = strings.Trim(args[i], " ")
	}

	product := args[0]

	cost, err := strconv.Atoi(args[1])
	if err != nil {
		return PurchaseParseRes{}, fmt.Errorf("cost format error: %q", cost)
	}

	quantity := 1
	if len(args) > 2 {
		quantity, err = strconv.Atoi(args[2])
		if err != nil {
			return PurchaseParseRes{}, fmt.Errorf("quantity format error: %q", cost)
		}
	}

	return PurchaseParseRes{
		Quantity:     quantity,
		ProductTitle: product,
		Price:        cost,
	}, nil
}
