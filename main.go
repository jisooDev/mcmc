package main

import (
	"log"
	"mcmc/config"
	"mcmc/database"
	"mcmc/process"
	"mcmc/sftp"
	"os"
	"path/filepath"
)

func main() {
	// ตั้งค่า logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("เริ่มต้นการประมวลผล...")

	// โหลดการตั้งค่า
	cfg := config.GetConfig()

	// เชื่อมต่อกับฐานข้อมูล
	log.Println("กำลังเชื่อมต่อกับฐานข้อมูล...")
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("ไม่สามารถเชื่อมต่อกับฐานข้อมูลได้: %v", err)
	}
	defer db.Close()
	log.Println("เชื่อมต่อกับฐานข้อมูลสำเร็จ")

	// เชื่อมต่อกับ SFTP
	log.Println("กำลังเชื่อมต่อกับ SFTP server...")
	sftpClient, err := sftp.NewClient(cfg.SFTP)
	if err != nil {
		log.Fatalf("ไม่สามารถเชื่อมต่อกับ SFTP server ได้: %v", err)
	}
	defer sftpClient.Close()
	log.Println("เชื่อมต่อกับ SFTP server สำเร็จ")

	// สร้างโฟลเดอร์สำหรับเก็บไฟล์ที่ดาวน์โหลด
	if _, err := os.Stat(cfg.App.DownloadDir); os.IsNotExist(err) {
		os.MkdirAll(cfg.App.DownloadDir, 0755)
	}

	// ดาวน์โหลดและประมวลผลไฟล์แต่ละประเภท
	for _, prefix := range cfg.App.FileTypes {
		log.Printf("กำลังค้นหาไฟล์ล่าสุดที่ขึ้นต้นด้วย %s...", prefix)

		// หาไฟล์ล่าสุดตาม prefix
		latestFile, err := sftpClient.FindLatestFileByPrefix(prefix)
		if err != nil {
			log.Printf("ข้อผิดพลาด: %v", err)
			continue
		}

		log.Printf("พบไฟล์ล่าสุด: %s", latestFile.Name())

		// กำหนดพาธของไฟล์
		remoteFilePath := cfg.SFTP.RemotePath + "/" + latestFile.Name()
		localFilePath := filepath.Join(cfg.App.DownloadDir, latestFile.Name())

		// ดาวน์โหลดไฟล์
		log.Printf("กำลังดาวน์โหลดไฟล์ %s...", latestFile.Name())
		bytesDownloaded, err := sftpClient.DownloadFile(remoteFilePath, localFilePath)
		if err != nil {
			log.Printf("ไม่สามารถดาวน์โหลดไฟล์ได้: %v", err)
			db.LogFileProcessing(latestFile.Name(), "failed", 0, "ไม่สามารถดาวน์โหลดไฟล์ได้: "+err.Error())
			continue
		}
		log.Printf("ดาวน์โหลดไฟล์ %s สำเร็จ (%d bytes)", latestFile.Name(), bytesDownloaded)

		// ประมวลผลไฟล์
		log.Printf("กำลังประมวลผลไฟล์ %s...", latestFile.Name())
		recordCount, err := process.ProcessFile(db, localFilePath)
		if err != nil {
			log.Printf("ไม่สามารถประมวลผลไฟล์ได้: %v", err)
			db.LogFileProcessing(latestFile.Name(), "failed", 0, "ไม่สามารถประมวลผลไฟล์ได้: "+err.Error())
		} else {
			log.Printf("ประมวลผลไฟล์ %s สำเร็จ (%d รายการ)", latestFile.Name(), recordCount)
			db.LogFileProcessing(latestFile.Name(), "success", recordCount, "")
		}

		// ลบไฟล์ที่ดาวน์โหลดมาเมื่อประมวลผลเสร็จแล้ว
		log.Printf("กำลังลบไฟล์ %s...", localFilePath)
		if err := os.Remove(localFilePath); err != nil {
			log.Printf("ไม่สามารถลบไฟล์ได้: %v", err)
		} else {
			log.Printf("ลบไฟล์ %s สำเร็จ", localFilePath)
		}
	}

	log.Println("การประมวลผลเสร็จสิ้น")
}