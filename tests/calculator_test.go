package tests

import (
	"collector/internal/calculator"
	"encoding/json"
	"reflect"
	"testing"
)

func TestCalculator(t *testing.T) {
	products := map[int64]calculator.Product{
		1: {
			1,
			"Вода",
			1,
			100,
		},
		2: {
			2,
			"Пиво",
			1,
			250,
		},
	}
	users := []calculator.User{
		{
			ID:       1,
			Username: "user_1",
			Expenses: []calculator.Expense{
				{
					ProductID:     2,
					BuyerID:       2,
					BuyerUsername: "user_2",
				},
			},
		},
		{
			ID:       2,
			Username: "user_2",
			Expenses: []calculator.Expense{
				{
					ProductID:     1,
					BuyerID:       1,
					BuyerUsername: "user_1",
				},
			},
		},
	}

	want := map[string]map[string]*calculator.Debt{
		"user_1": {
			"user_2": {
				150,
				[]calculator.HistoryItem{
					{
						calculator.OUTCOME,
						"Пиво",
						250,
					},
					{
						calculator.INCOME,
						"Вода",
						100,
					},
				},
			},
		},
	}

	calc := calculator.New(users, products)
	got := calc.Calculate()

	if !reflect.DeepEqual(got, want) {
		wantM, _ := json.Marshal(want)
		gotM, _ := json.Marshal(got)
		t.Errorf("want %s but got %s", wantM, gotM)
	}
}
