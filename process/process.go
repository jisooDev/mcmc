package process

import (
	"encoding/csv"
	"fmt"
	"io"
	"mcmc/database"
	"mcmc/models"
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

	// ประมวลผลตามประเภทไฟล์
	var recordCount int

	switch fileType {
	case "header":
		headers, err := readHeaderData(reader, header, fileName)
		if err != nil {
			return 0, err
		}
		recordCount, _ = db.SaveSaleOrderHeader(headers)

	case "item":
		items, err := readItemData(reader, header, fileName)
		if err != nil {
			return 0, err
		}
		recordCount, _ = db.SaveSaleOrderItem(items)

	case "summary":
		summaries, err := readSummaryData(reader, header, fileName)
		if err != nil {
			return 0, err
		}
		recordCount, _ = db.SaveSaleOrderSummary(summaries)
	}

	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถบันทึกข้อมูลได้: %v", err)
	}

	return recordCount, nil
}

// readHeaderData อ่านข้อมูลจากไฟล์ header
func readHeaderData(reader *csv.Reader, headers []string, sourceFile string) ([]models.SaleOrderHeader, error) {
	var result []models.SaleOrderHeader

	// ตรวจสอบ header
	expectedHeaders := []string{"DocNo", "OnDate", "DeliveryDate", "SOCustomerId", "CustomerName", "Status", "TerritoryCode", "TotalAmount", "TotalVat", "Remark"}
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("จำนวน header ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(headers))
	}

	// อ่านแต่ละแถว
	lineCount := 0
	for {
		lineCount++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("เกิดข้อผิดพลาดในการอ่านไฟล์บรรทัดที่ %d: %v", lineCount, err)
		}

		// ตรวจสอบจำนวนฟิลด์
		if len(record) < len(expectedHeaders) {
			return nil, fmt.Errorf("จำนวนฟิลด์ไม่ถูกต้องที่บรรทัด %d: คาดหวัง %d แต่เจอ %d", lineCount, len(expectedHeaders), len(record))
		}

		// แปลงข้อมูลและเพิ่มลงในผลลัพธ์
		onDate := parseDate(record[1])
		deliveryDate := parseDate(record[2])
		
		// ตรวจสอบค่าว่าง
		var soCustomerId, customerName, status, territoryCode, remark string
		if len(record) > 3 && record[3] != "" {
			soCustomerId = record[3]
		}
		if len(record) > 4 && record[4] != "" {
			customerName = record[4]
		}
		if len(record) > 5 && record[5] != "" {
			status = record[5]
		}
		if len(record) > 6 && record[6] != "" {
			territoryCode = record[6]
		}
		if len(record) > 9 && record[9] != "" {
			remark = record[9]
		}

		header := models.SaleOrderHeader{
			DocNo:         record[0],
			SOCustomerID:  soCustomerId,
			CustomerName:  customerName, 
			Status:        status,
			TerritoryCode: territoryCode,
			TotalAmount:   parseFloat(record[7]),
			TotalVat:      parseFloat(record[8]),
			Remark:        remark,
			SourceFile:    sourceFile,
			CreatedAt:     time.Now(),
		}
		
		// ตรวจสอบว่ามีค่าวันที่หรือไม่ก่อนกำหนดให้กับ pointer
		if !onDate.IsZero() {
			header.OnDate = &onDate
		}
		if !deliveryDate.IsZero() {
			header.DeliveryDate = &deliveryDate
		}

		result = append(result, header)
	}

	return result, nil
}

// readItemData อ่านข้อมูลจากไฟล์ item
func readItemData(reader *csv.Reader, headers []string, sourceFile string) ([]models.SaleOrderItem, error) {
	var result []models.SaleOrderItem

	// ตรวจสอบ header
	expectedHeaders := []string{"DocNo", "ItemId", "ProductCode", "SOProductId", "SKUUnitTypeId", "Quantity", "Price", "Amount", "Vat", "VatRate", "ItemType", "OrderRank", "RefItemId", "IO_Number"}
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("จำนวน header ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(headers))
	}

	// อ่านแต่ละแถว
	lineCount := 0
	for {
		lineCount++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("เกิดข้อผิดพลาดในการอ่านไฟล์บรรทัดที่ %d: %v", lineCount, err)
		}

		// ตรวจสอบจำนวนฟิลด์
		if len(record) < len(expectedHeaders) {
			return nil, fmt.Errorf("จำนวนฟิลด์ไม่ถูกต้องที่บรรทัด %d: คาดหวัง %d แต่เจอ %d", lineCount, len(expectedHeaders), len(record))
		}

		// ตรวจสอบค่าว่าง
		var itemId, productCode, soProductId, skuUnitTypeId, itemType, refItemId, ioNumber string
		if len(record) > 1 && record[1] != "" {
			itemId = record[1]
		}
		if len(record) > 2 && record[2] != "" {
			productCode = record[2]
		}
		if len(record) > 3 && record[3] != "" {
			soProductId = record[3]
		}
		if len(record) > 4 && record[4] != "" {
			skuUnitTypeId = record[4]
		}
		if len(record) > 10 && record[10] != "" {
			itemType = record[10]
		}
		if len(record) > 12 && record[12] != "" {
			refItemId = record[12]
		}
		if len(record) > 13 && record[13] != "" {
			ioNumber = record[13]
		}

		// แปลงข้อมูลและเพิ่มลงในผลลัพธ์
		item := models.SaleOrderItem{
			DocNo:         record[0],
			ItemID:        itemId,
			ProductCode:   productCode,
			SOProductID:   soProductId,
			SKUUnitTypeID: skuUnitTypeId,
			Quantity:      parseFloat(record[5]),
			Price:         parseFloat(record[6]),
			Amount:        parseFloat(record[7]),
			Vat:           parseFloat(record[8]),
			VatRate:       parseFloat(record[9]),
			ItemType:      itemType,
			OrderRank:     parseInt(record[11]),
			RefItemID:     refItemId,
			IONumber:      ioNumber,
			SourceFile:    sourceFile,
			CreatedAt:     time.Now(),
		}

		result = append(result, item)
	}

	return result, nil
}

// readSummaryData อ่านข้อมูลจากไฟล์ summary
func readSummaryData(reader *csv.Reader, headers []string, sourceFile string) ([]models.SaleOrderSummary, error) {
	var result []models.SaleOrderSummary

	// ตรวจสอบ header
	expectedHeaders := []string{"HeaderCount", "ItemCount"}
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("จำนวน header ไม่ถูกต้อง: คาดหวัง %d แต่เจอ %d", len(expectedHeaders), len(headers))
	}

	// อ่านแต่ละแถว
	lineCount := 0
	for {
		lineCount++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("เกิดข้อผิดพลาดในการอ่านไฟล์บรรทัดที่ %d: %v", lineCount, err)
		}

		// ตรวจสอบจำนวนฟิลด์
		if len(record) < len(expectedHeaders) {
			return nil, fmt.Errorf("จำนวนฟิลด์ไม่ถูกต้องที่บรรทัด %d: คาดหวัง %d แต่เจอ %d", lineCount, len(expectedHeaders), len(record))
		}

		// แปลงข้อมูลและเพิ่มลงในผลลัพธ์
		summary := models.SaleOrderSummary{
			HeaderCount: parseInt(record[0]),
			ItemCount:   parseInt(record[1]),
			SourceFile:  sourceFile,
			CreatedAt:   time.Now(),
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

	// ลองรูปแบบต่างๆ ของวันที่
	layouts := []string{
		"2006-01-02 15:04:05", 
        "2006-01-02T15:04:05",
		"2006-01-02",
		"02/01/2006",
		"2/1/2006",
		"2006/01/02",
	}

	for _, layout := range layouts {
		date, err := time.Parse(layout, dateStr)
		if err == nil {
			return date
		}
	}

	// ถ้าแปลงไม่สำเร็จในทุกรูปแบบ
	return time.Time{}
}

func parseFloat(numStr string) float64 {
	if numStr == "" {
		return 0
	}

	// ลบตัวอักษรที่ไม่เกี่ยวข้องกับตัวเลข (เช่น สัญลักษณ์สกุลเงิน)
	cleanStr := strings.ReplaceAll(numStr, ",", "")
	cleanStr = strings.TrimSpace(cleanStr)

	num, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0
	}

	return num
}

func parseInt(numStr string) int {
	if numStr == "" {
		return 0
	}

	// ลบตัวอักษรที่ไม่เกี่ยวข้องกับตัวเลข
	cleanStr := strings.ReplaceAll(numStr, ",", "")
	cleanStr = strings.TrimSpace(cleanStr)

	num, err := strconv.Atoi(cleanStr)
	if err != nil {
		// ลองแปลงเป็น float ก่อนแล้วค่อยแปลงเป็น int
		floatNum, floatErr := strconv.ParseFloat(cleanStr, 64)
		if floatErr != nil {
			return 0
		}
		return int(floatNum)
	}

	return num
}