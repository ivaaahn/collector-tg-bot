package tests

import (
	"collector/internal/debts_calculation"
	"encoding/json"
	"reflect"
	"testing"
)

func TestCalculator(t *testing.T) {
	products := map[int64]debts_calculation.Product{
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
	users := []debts_calculation.User{
		{
			ID:       1,
			Username: "user_1",
			Expenses: []debts_calculation.Expense{
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
			Expenses: []debts_calculation.Expense{
				{
					ProductID:     1,
					BuyerID:       1,
					BuyerUsername: "user_1",
				},
			},
		},
	}

	want := map[string]map[string]*debts_calculation.Debt{
		"user_1": {
			"user_2": {
				150,
				[]debts_calculation.HistoryItem{
					{
						debts_calculation.OUTCOME,
						"Пиво",
						250,
					},
					{
						debts_calculation.INCOME,
						"Вода",
						100,
					},
				},
			},
		},
	}

	calc := debts_calculation.NewCalculator(users, products)
	got := calc.Calculate()

	if !reflect.DeepEqual(got, want) {
		wantM, _ := json.Marshal(want)
		gotM, _ := json.Marshal(got)
		t.Errorf("want %s but got %s", wantM, gotM)
	}
}
