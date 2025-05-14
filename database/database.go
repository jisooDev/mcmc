package database

import (
	"database/sql"
	"fmt"
	"mcmc/config"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// DB คือ connection pool สำหรับฐานข้อมูล
type DB struct {
	*sql.DB
}

// NewDB สร้างการเชื่อมต่อใหม่กับฐานข้อมูล
func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.DBName)

	db, err := sql.Open(cfg.Driver, connString)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อกับฐานข้อมูลได้: %v", err)
	}

	// ทดสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถ ping ฐานข้อมูลได้: %v", err)
	}

	// ตั้งค่า connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

// LogFileProcessing บันทึกประวัติการประมวลผลไฟล์
func (db *DB) LogFileProcessing(filename, status string, recordCount int, errorMessage string) error {
	query := `
		INSERT INTO file_processing_logs
		(filename, status, record_count, error_message, created_at)
		VALUES (@p1, @p2, @p3, @p4, @p5)
	`

	_, err := db.Exec(query, filename, status, recordCount, errorMessage, time.Now())
	if err != nil {
		return fmt.Errorf("ไม่สามารถบันทึกประวัติการประมวลผลไฟล์ได้: %v", err)
	}

	return nil
}

// SaveSaleOrderHeader บันทึกข้อมูล header ลงฐานข้อมูล
func (db *DB) SaveSaleOrderHeader(headers []map[string]interface{}, sourceFile string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถเริ่ม transaction ได้: %v", err)
	}

	query := `
		INSERT INTO saleorder_header
		(doc_no, on_date, delivery_date, so_customer_id, customer_name, status,
		territory_code, total_amount, total_vat, remark, source_file, created_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12)
	`

	count := 0
	for _, header := range headers {
		_, err := tx.Exec(
			query,
			header["doc_no"],
			header["on_date"],
			header["delivery_date"],
			header["so_customer_id"],
			header["customer_name"],
			header["status"],
			header["territory_code"],
			header["total_amount"],
			header["total_vat"],
			header["remark"],
			sourceFile,
			time.Now(),
		)

		if err != nil {
			tx.Rollback()
			return count, fmt.Errorf("ไม่สามารถบันทึกข้อมูล header ได้: %v", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("ไม่สามารถ commit transaction ได้: %v", err)
	}

	return count, nil
}

// SaveSaleOrderItem บันทึกข้อมูล item ลงฐานข้อมูล
func (db *DB) SaveSaleOrderItem(items []map[string]interface{}, sourceFile string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถเริ่ม transaction ได้: %v", err)
	}

	query := `
		INSERT INTO saleorder_item
		(doc_no, item_id, product_code, so_product_id, sku_unit_type_id, quantity,
		price, amount, vat, vat_rate, item_type, order_rank, ref_item_id, io_number,
		source_file, created_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12, @p13, @p14, @p15, @p16)
	`

	count := 0
	for _, item := range items {
		_, err := tx.Exec(
			query,
			item["doc_no"],
			item["item_id"],
			item["product_code"],
			item["so_product_id"],
			item["sku_unit_type_id"],
			item["quantity"],
			item["price"],
			item["amount"],
			item["vat"],
			item["vat_rate"],
			item["item_type"],
			item["order_rank"],
			item["ref_item_id"],
			item["io_number"],
			sourceFile,
			time.Now(),
		)

		if err != nil {
			tx.Rollback()
			return count, fmt.Errorf("ไม่สามารถบันทึกข้อมูล item ได้: %v", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("ไม่สามารถ commit transaction ได้: %v", err)
	}

	return count, nil
}

// SaveSaleOrderSummary บันทึกข้อมูล summary ลงฐานข้อมูล
func (db *DB) SaveSaleOrderSummary(summaries []map[string]interface{}, sourceFile string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถเริ่ม transaction ได้: %v", err)
	}

	query := `
		INSERT INTO saleorder_summary
		(header_count, item_count, source_file, created_at)
		VALUES (@p1, @p2, @p3, @p4)
	`

	count := 0
	for _, summary := range summaries {
		_, err := tx.Exec(
			query,
			summary["header_count"],
			summary["item_count"],
			sourceFile,
			time.Now(),
		)

		if err != nil {
			tx.Rollback()
			return count, fmt.Errorf("ไม่สามารถบันทึกข้อมูล summary ได้: %v", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("ไม่สามารถ commit transaction ได้: %v", err)
	}

	return count, nil
}