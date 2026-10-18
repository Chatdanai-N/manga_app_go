package models

import (
	"testing"
	"time"
)

func TestMangaUsers_TableName(t *testing.T) {
	param := MangaUsers{}
	expected := "manga_users"

	got := param.TableName()
	if got != expected {
		t.Errorf("TableName(): %s expected: %v", got, expected)
	}
}

func TestMangaUsers_StructFields(t *testing.T) {
	param := MangaUsers{
		Id:           1,
		Username:     "username",
		Email:        "email",
		PasswordHash: "password",
		CreatedDate:  time.Now(),
		UpdatedDate:  time.Now(),
	}

	if param.Id != 1 {
		t.Errorf("expected id = %d, got %d", 1, param.Id)
	}

	if param.Username != "username" {
		t.Errorf("expected username = %s, got %s", "username", param.Username)
	}

	if param.Email != "email" {
		t.Errorf("expected email = %s, got %s", "email", param.Email)
	}

	if param.PasswordHash != "password" {
		t.Errorf("expected password = %s, got %s", "password", param.PasswordHash)
	}

	if param.CreatedDate != time.Now() {
		t.Errorf("expected created = %s, got %s", time.Now(), param.CreatedDate)
	}

	if param.UpdatedDate != time.Now() {
		t.Errorf("expected updated = %s, got %s", time.Now(), param.UpdatedDate)
	}
}
