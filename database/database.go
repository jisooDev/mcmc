package database

import (
	"fmt"
	"log"
	"mcmc/config"
	"mcmc/models"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB คือ GORM DB connection
type DB struct {
	*gorm.DB
}

// NewDB สร้างการเชื่อมต่อใหม่กับฐานข้อมูล
func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.DBName)

	// ตั้งค่า GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(sqlserver.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อกับฐานข้อมูลได้: %v", err)
	}

	// ตั้งค่า connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถรับ underlying database connection ได้: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

// CheckFileProcessed ตรวจสอบว่าไฟล์เคยประมวลผลแล้วหรือไม่
func (db *DB) CheckFileProcessed(filename string) (bool, error) {
	var count int64
	query := "SELECT COUNT(*) FROM file_processing_logs WHERE filename = ? AND status = 'success'"
	
	err := db.DB.Raw(query, filename).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("ไม่สามารถตรวจสอบประวัติการประมวลผลไฟล์ได้: %v", err)
	}
	
	return count > 0, nil
}

// LogFileProcessing บันทึกประวัติการประมวลผลไฟล์
func (db *DB) LogFileProcessing(filename, status string, recordCount int, errorMessage string) error {
	log := models.FileProcessingLog{
		Filename:     filename,
		Status:       status,
		RecordCount:  recordCount,
		ErrorMessage: errorMessage,
		CreatedAt:    time.Now(),
	}

	result := db.Create(&log)
	if result.Error != nil {
		return fmt.Errorf("ไม่สามารถบันทึกประวัติการประมวลผลไฟล์ได้: %v", result.Error)
	}

	return nil
}

// SaveSaleOrderHeader บันทึกข้อมูล header ลงฐานข้อมูล
func (db *DB) SaveSaleOrderHeader(headers []models.SaleOrderHeader) (int, error) {
	result := db.Create(&headers)
	if result.Error != nil {
		return 0, fmt.Errorf("ไม่สามารถบันทึกข้อมูล header ได้: %v", result.Error)
	}

	return len(headers), nil
}

// SaveSaleOrderItem บันทึกข้อมูล item ลงฐานข้อมูล
func (db *DB) SaveSaleOrderItem(items []models.SaleOrderItem) (int, error) {
	result := db.Create(&items)
	if result.Error != nil {
		return 0, fmt.Errorf("ไม่สามารถบันทึกข้อมูล item ได้: %v", result.Error)
	}

	return len(items), nil
}

// SaveSaleOrderSummary บันทึกข้อมูล summary ลงฐานข้อมูล
func (db *DB) SaveSaleOrderSummary(summaries []models.SaleOrderSummary) (int, error) {
	result := db.Create(&summaries)
	if result.Error != nil {
		return 0, fmt.Errorf("ไม่สามารถบันทึกข้อมูล summary ได้: %v", result.Error)
	}

	return len(summaries), nil
}

func (db *DB) MigrateDB() error {
	log.Println("กำลังทำ database migration...")
	
	// สร้างหรืออัปเดตตารางตาม models
	err := db.DB.AutoMigrate(
		&models.FileProcessingLog{},
		&models.SaleOrderHeader{},
		&models.SaleOrderItem{},
		&models.SaleOrderSummary{},
	)
	
	if err != nil {
		return fmt.Errorf("ไม่สามารถทำ migration ได้: %v", err)
	}
	
	log.Println("ทำ database migration สำเร็จ")
	return nil
}