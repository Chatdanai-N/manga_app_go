package database

import (
	"fmt"
	"log"
	"manga_app/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(cfg *config.Config) (*gorm.DB, error) {
	// 1. ประกอบ DSN จาก Config struct
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable search_path=%s TimeZone=Asia/Bangkok",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DbSearchPath,
	)

	// 2. Open Connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 3. ตั้งค่า Connection Pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)               // จำนวน connection สำรองสูงสุด
	sqlDB.SetMaxOpenConns(100)              // จำนวน connection สูงสุดที่ยอมให้เปิดพร้อมกัน
	sqlDB.SetConnMaxLifetime(1 * time.Hour) // อายุสูงสุดของ connection

	log.Println("Database connection successfully established")
	return db, nil
}
