package models

import "testing"

func TestMangaSystemParameter_TableName(t *testing.T) {
	param := MangaSystemParameter{}
	expected := "manga_system_parameter"

	got := param.TableName()

	if got != expected {
		t.Errorf("TableName() = %s; expected %s", got, expected)
	}
}

func TestMangaSystemParameter_StructFields(t *testing.T) {
	param := MangaSystemParameter{
		ParameterId: 1,
		GroupCode:   "SYSTEM",
		Code:        "MAX_LIMIT",
		Value:       "100",
		OrderNumber: 1,
	}

	if param.ParameterId != 1 {
		t.Errorf("expected ParameterId = 1 , got %d", param.ParameterId)
	}

	if param.GroupCode != "SYSTEM" {
		t.Errorf("expected GroupCode = SYSTEM , got %s", param.GroupCode)
	}

	if param.Code != "MAX_LIMIT" {
		t.Errorf("expected Code = MAX_LIMIT , got %s", param.Code)
	}

	if param.Value != "100" {
		t.Errorf("expected Value = 100 , got %s", param.Value)
	}

	if param.OrderNumber != 1 {
		t.Errorf("expected OrderNumber = 1 , got %d", param.OrderNumber)
	}
}
