package tests

import (
	"collector/internal/parser"
	"reflect"
	"testing"
)

func TestAddPurchases(t *testing.T) {
	// Arrange
	input := `Пиво, 100
Водка синяя, 250
Огурцы, 400, 2
`
	want := []parser.PurchaseParseRes{
		{
			ProductTitle: "Пиво",
			Price:        100,
			Quantity:     1,
		},
		{
			ProductTitle: "Водка синяя",
			Price:        250,
			Quantity:     1,
		},
		{
			ProductTitle: "Огурцы",
			Price:        400,
			Quantity:     2,
		},
	}
	parserObj := parser.NewPurchaseParser()

	// Act
	got, err := parserObj.Parse(input)

	// Assert
	if err != nil {
		t.Fatalf("got error: %q", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v but got %v", want, got)
	}
}

func TestAddExpensesPurchase(t *testing.T) {
	// Arrange
	input := `1
2 @me
3 @all
4 @ivaaahn @me
`
	want := []parser.ExpenseParseRes{
		{
			ProductNo: 1,
			Eaters:    []string{"@me"},
		},
		{
			ProductNo: 2,
			Eaters:    []string{"@me"},
		},
		{
			ProductNo: 3,
			Eaters:    []string{"@all"},
		},
		{
			ProductNo: 4,
			Eaters:    []string{"@ivaaahn", "@me"},
		},
	}
	parserObj := parser.NewExpenseParser()

	// Act
	got, err := parserObj.Parse(input)

	// Assert
	if err != nil {
		t.Fatalf("got error: %q", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v but got %v", want, got)
	}
}
