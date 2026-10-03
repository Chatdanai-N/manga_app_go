package database

import (
	"manga_app/config"
	"testing"
)

func TestConnect_SuccessConnected(t *testing.T) {

	// 1.Arrange
	// load config datasource with .env
	cfg, err := config.LoadConfig("../.env")

	// 2.Act
	db, err := ConnectDatabase(cfg)

	// 3.Assert
	if db == nil {
		t.Errorf("expected db not to be nil : got %v", db)
	}

	if err != nil {
		t.Errorf("failed to connect to database : got %v", err)
	}
}

// TestConnectDatabase_InvalidConfig
func TestConnectDatabase_InvalidConfig(t *testing.T) {

	// 1. Arrange
	cfg := config.Config{
		DBPort:       "5432",
		DBUser:       "root",
		DBPassword:   "secret",
		DBName:       "testdb",
		DbSearchPath: "public",
	}

	// 2.Act
	db, err := ConnectDatabase(&cfg)

	// 3.Assert
	if err != nil {
		t.Errorf("expected error when open connect error = %v", err)
	}

	if db == nil {
		t.Errorf("expected db to be nil on error = %v", err)
	}
}
