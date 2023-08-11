package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ExpenseParseRes struct {
	Eaters    []string
	ProductNo int64
}

type ExpenseParser struct {
	input string
}

func NewExpenseParser() *ExpenseParser {
	return &ExpenseParser{}
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
