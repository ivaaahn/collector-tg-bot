package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type PurchaseParseRes struct {
	Quantity     int
	Price        int
	ProductTitle string
}

type PurchaseParser struct {
	input string
}

func NewPurchaseParser() *PurchaseParser {
	return &PurchaseParser{}
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
