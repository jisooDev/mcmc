package process

import (
	"encoding/csv"
	"fmt"
	"io"
	"mcmc/database"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ProcessFile ประมวลผลไฟล์ CSV และเตรียมข้อมูลสำหรับบันทึกลงฐานข้อมูล
func ProcessFile(db *database.DB, filePath string) (int, error) {
	// ตรวจสอบประเภทไฟล์จากชื่อ
	fileName := filepath.Base(filePath)
	fileType := ""

	if strings.Contains(fileName, "saleorder_header_") {
		fileType = "header"
	} else if strings.Contains(fileName, "saleorder_item_") {
		fileType = "item"
	} else if strings.Contains(fileName, "saleorder_summary_") {
		fileType = "summary"
	} else {
		return 0, fmt.Errorf("ไม่รู้จักประเภทไฟล์: %s", fileName)
	}

	// เปิดไฟล์ CSV
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถเปิดไฟล์ได้: %v", err)
	}
	defer file.Close()

	// สร้าง CSV reader
	reader := csv.NewReader(file)
	reader.Comma = '|' // กำหนดตัวคั่นเป็น pipe (|)

	// อ่าน header
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถอ่าน header ได้: %v", err)
	}

	// อ่านข้อมูลและแปลงเป็น maps
	var data []map[string]interface{}
	var recordCount int

	// ประมวลผลตามประเภทไฟล์
	switch fileType {
	case "header":
		data, err = readHeaderData(reader, header)
		if err != nil {
			return 0, err
		}
		recordCount, err = db.SaveSaleOrderHeader(data, fileName)

	case "item":
		data, err = readItemData(reader, header)
		if err != nil {
			return 0, err
		}
		recordCount, err = db.SaveSaleOrderItem(data, fileName)

	case "summary":
		data, err = readSummaryData(reader, header)
		if err != nil {
			return 0, err
		}
		recordCount, err = db.SaveSaleOrderSummary(data, fileName)
	}

	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถบันทึกข้อมูลได้: %v", err)
	}

	return recordCount, nil
}

// readHeaderData อ่านข้อมูลจากไฟล์ header
func readHeaderData(reader *csv.Reader, headers []string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// ตรวจสอบ header
	expectedHeaders := []string{"DocNo", "OnDate", "DeliveryDate", "SOCustomerId", "CustomerName", "Status", "TerritoryCode", "TotalAmount", "TotalVat", "Remark"}
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("จำนวน header ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(headers))
	}

	// อ่านแต่ละแถว
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("เกิดข้อผิดพลาดในการอ่านไฟล์: %v", err)
		}

		// ตรวจสอบจำนวนฟิลด์
		if len(record) < len(expectedHeaders) {
			return nil, fmt.Errorf("จำนวนฟิลด์ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(record))
		}

		// แปลงข้อมูลและเพิ่มลงในผลลัพธ์
		header := map[string]interface{}{
			"doc_no":         record[0],
			"on_date":        parseDate(record[1]),
			"delivery_date":  parseDate(record[2]),
			"so_customer_id": record[3],
			"customer_name":  record[4],
			"status":         record[5],
			"territory_code": record[6],
			"total_amount":   parseFloat(record[7]),
			"total_vat":      parseFloat(record[8]),
			"remark":         record[9],
		}

		result = append(result, header)
	}

	return result, nil
}

// readItemData อ่านข้อมูลจากไฟล์ item
func readItemData(reader *csv.Reader, headers []string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// ตรวจสอบ header
	expectedHeaders := []string{"DocNo", "ItemId", "ProductCode", "SOProductId", "SKUUnitTypeId", "Quantity", "Price", "Amount", "Vat", "VatRate", "ItemType", "OrderRank", "RefItemId", "IO_Number"}
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("จำนวน header ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(headers))
	}

	// อ่านแต่ละแถว
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("เกิดข้อผิดพลาดในการอ่านไฟล์: %v", err)
		}

		// ตรวจสอบจำนวนฟิลด์
		if len(record) < len(expectedHeaders) {
			return nil, fmt.Errorf("จำนวนฟิลด์ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(record))
		}

		// แปลงข้อมูลและเพิ่มลงในผลลัพธ์
		item := map[string]interface{}{
			"doc_no":           record[0],
			"item_id":          record[1],
			"product_code":     record[2],
			"so_product_id":    record[3],
			"sku_unit_type_id": record[4],
			"quantity":         parseFloat(record[5]),
			"price":            parseFloat(record[6]),
			"amount":           parseFloat(record[7]),
			"vat":              parseFloat(record[8]),
			"vat_rate":         parseFloat(record[9]),
			"item_type":        record[10],
			"order_rank":       parseInt(record[11]),
			"ref_item_id":      record[12],
			"io_number":        record[13],
		}

		result = append(result, item)
	}

	return result, nil
}

// readSummaryData อ่านข้อมูลจากไฟล์ summary
func readSummaryData(reader *csv.Reader, headers []string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// ตรวจสอบ header
	expectedHeaders := []string{"HeaderCount", "ItemCount"}
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("จำนวน header ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(headers))
	}

	// อ่านแต่ละแถว
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("เกิดข้อผิดพลาดในการอ่านไฟล์: %v", err)
		}

		// ตรวจสอบจำนวนฟิลด์
		if len(record) < len(expectedHeaders) {
			return nil, fmt.Errorf("จำนวนฟิลด์ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(record))
		}

		// แปลงข้อมูลและเพิ่มลงในผลลัพธ์
		summary := map[string]interface{}{
			"header_count": parseInt(record[0]),
			"item_count":   parseInt(record[1]),
		}

		result = append(result, summary)
	}

	return result, nil
}

// ฟังก์ชันช่วยสำหรับแปลงข้อมูล
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}

	return date
}

func parseFloat(numStr string) float64 {
	if numStr == "" {
		return 0
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	return num
}

func parseInt(numStr string) int {
	if numStr == "" {
		return 0
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	return num
}