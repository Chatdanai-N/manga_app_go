package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	// 1.Arrange
	envContent := `DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=secret
DB_NAME=testdb
DB_SEARCH_PATH=public
PORT=8080`
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp env file: %v", err)
	}

	// ย้าย Working Directory ชั่วคราวไปที่ Directory ที่มีไฟล์ .env
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	// 2.Act
	cfg, err := LoadConfig()

	// 3.Assert
	if err != nil {
		t.Fatalf("expected no error , got %v", err)
	}

	if cfg == nil {
		t.Fatal("expected cfg not to be nil")
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost to be localhost, got %s", cfg.DBHost)
	}

	if cfg.DBPort != "5432" {
		t.Errorf("expected DBPort to be 5432, got %s", cfg.DBPort)
	}

	if cfg.DBUser != "root" {
		t.Errorf("expected DBUser to be root, got %s", cfg.DBUser)
	}

	if cfg.DBPassword != "secret" {
		t.Errorf("expected DBPassword to be secret, got %s", cfg.DBPassword)
	}

	if cfg.DBName != "testdb" {
		t.Errorf("expected DBName to be testdb, got %s", cfg.DBName)
	}

	if cfg.DbSearchPath != "public" {
		t.Errorf("expected DBSearchPath to be public, got %s", cfg.DbSearchPath)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected Port to be 8080, got %s", cfg.Port)
	}

}

// TestLoadConfig_FileNotFound ทดสอบกรณีไม่มีไฟล์ .env (ต้อง return error)
func TestLoadConfig_FileNotFound(t *testing.T) {

	// arrange
	tmpDir := t.TempDir()

	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// act
	cfg, err := LoadConfig()

	// assert
	if err != nil {
		t.Error("expected error when .env file  does not exist, got nil")
	}

	if cfg != nil {
		t.Errorf("expected config to be nil on error, got %v", cfg)
	}
}
