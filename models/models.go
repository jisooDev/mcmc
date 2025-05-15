package models

import (
	"time"
)

// FileProcessingLog เก็บประวัติการประมวลผลไฟล์
type FileProcessingLog struct {
	ID           uint      `gorm:"primaryKey"`
	Filename     string    `gorm:"uniqueIndex;type:nvarchar(255);not null"` 
	Status       string    `gorm:"type:nvarchar(10);not null"`
	RecordCount  int       `gorm:"not null;default:0"`
	ErrorMessage string    `gorm:"type:nvarchar(max)"`
	CreatedAt    time.Time `gorm:"type:datetime2;not null"`
}

// SaleOrderHeader เก็บข้อมูลหัวเอกสารการขาย
type SaleOrderHeader struct {
	ID            uint       `gorm:"primaryKey"`
	DocNo         string     `gorm:"type:nvarchar(50);not null"`
	OnDate        *time.Time `gorm:"type:datetime2"`
	DeliveryDate  *time.Time `gorm:"type:datetime2"`
	SOCustomerID  string     `gorm:"type:nvarchar(50)"`
	CustomerName  string     `gorm:"type:nvarchar(255)"`
	Status        string     `gorm:"type:nvarchar(50)"`
	TerritoryCode string     `gorm:"type:nvarchar(50)"`
	TotalAmount   float64    `gorm:"type:decimal(18,2)"`
	TotalVat      float64    `gorm:"type:decimal(18,2)"`
	Remark        string     `gorm:"type:nvarchar(max)"`
	SourceFile    string     `gorm:"type:nvarchar(255);not null"`
	CreatedAt     time.Time  `gorm:"type:datetime2;not null"`
}

// SaleOrderItem เก็บข้อมูลรายการสินค้าในเอกสารการขาย
type SaleOrderItem struct {
	ID            uint      `gorm:"primaryKey"`
	DocNo         string    `gorm:"type:nvarchar(50);not null"`
	ItemID        string    `gorm:"type:nvarchar(50)"`
	ProductCode   string    `gorm:"type:nvarchar(50)"`
	SOProductID   string    `gorm:"type:nvarchar(50)"`
	SKUUnitTypeID string    `gorm:"type:nvarchar(50)"`
	Quantity      float64   `gorm:"type:decimal(18,2)"`
	Price         float64   `gorm:"type:decimal(18,2)"`
	Amount        float64   `gorm:"type:decimal(18,2)"`
	Vat           float64   `gorm:"type:decimal(18,2)"`
	VatRate       float64   `gorm:"type:decimal(18,2)"`
	ItemType      string    `gorm:"type:nvarchar(50)"`
	OrderRank     int       
	RefItemID     string    `gorm:"type:nvarchar(50)"`
	IONumber      string    `gorm:"type:nvarchar(50)"`
	SourceFile    string    `gorm:"type:nvarchar(255);not null"`
	CreatedAt     time.Time `gorm:"type:datetime2;not null"`
}

// SaleOrderSummary เก็บข้อมูลสรุปของเอกสารการขาย
type SaleOrderSummary struct {
	ID          uint      `gorm:"primaryKey"`
	HeaderCount int       
	ItemCount   int       
	SourceFile  string    `gorm:"type:nvarchar(255);not null"`
	CreatedAt   time.Time `gorm:"type:datetime2;not null"`
}