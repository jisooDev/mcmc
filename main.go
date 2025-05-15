package main

import (
	"log"
	"mcmc/config"
	"mcmc/database"
	"mcmc/process"
	"mcmc/sftp"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	// ตั้งค่า logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("เริ่มต้นการประมวลผล...")

	// โหลดการตั้งค่า
	cfg := config.GetConfig()

	if cfg.Cron.RunOnce {
		// โหมดรันครั้งเดียว: รันงานทันทีแล้วจบการทำงาน
		log.Println("รันในโหมดครั้งเดียว")
		runTask(cfg)
	} else {
		// โหมดรันเป็น cron job
		log.Printf("รันในโหมด cron job ด้วย schedule: %s", cfg.Cron.Schedule)
		
		// สร้าง cron scheduler
		c := cron.New()
		
		// ลงทะเบียน cron job
		_, err := c.AddFunc(cfg.Cron.Schedule, func() {
			runTask(cfg)
		})
		
		if err != nil {
			log.Fatalf("ไม่สามารถตั้งค่า cron job ได้: %v", err)
		}
		
		// เริ่มต้น cron scheduler
		c.Start()
		
		// รันงานทันทีหนึ่งครั้งโดยไม่ต้องรอให้ถึงเวลาตาม schedule
		log.Println("เริ่มรันงานครั้งแรกทันที...")
		go runTask(cfg)
		
		// แสดงเวลาที่จะรันครั้งถัดไป
		entries := c.Entries()
		if len(entries) > 0 {
			nextRun := entries[0].Next
			log.Printf("ครั้งถัดไปจะรันในเวลา: %s", nextRun.Format("2006-01-02 15:04:05"))
		}
		
		log.Println("กำลังรันเป็น background service (กด Ctrl+C เพื่อหยุดการทำงาน)")
		
		// รอไปเรื่อยๆ เพื่อให้โปรแกรมทำงานตลอดไป
		select {}
	}
}

// runTask ทำงานหลักของการดาวน์โหลดและประมวลผลไฟล์
func runTask(cfg *config.Config) {
	startTime := time.Now()
	log.Printf("เริ่มรันงานเวลา %s", startTime.Format("15:04:05"))

	// เชื่อมต่อกับฐานข้อมูล
	log.Println("กำลังเชื่อมต่อกับฐานข้อมูล...")
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		log.Printf("ไม่สามารถเชื่อมต่อกับฐานข้อมูลได้: %v", err)
		return
	}
	
	if err := db.MigrateDB(); err != nil {
		log.Printf("ไม่สามารถทำ migration ได้: %v", err)
		return
	}
	
	log.Println("เชื่อมต่อกับฐานข้อมูลสำเร็จ")

	// เชื่อมต่อกับ SFTP
	log.Println("กำลังเชื่อมต่อกับ SFTP server...")
	sftpClient, err := sftp.NewClient(cfg.SFTP)
	if err != nil {
		log.Printf("ไม่สามารถเชื่อมต่อกับ SFTP server ได้: %v", err)
		return
	}
	defer sftpClient.Close()
	log.Println("เชื่อมต่อกับ SFTP server สำเร็จ")

	// สร้างโฟลเดอร์สำหรับเก็บไฟล์ที่ดาวน์โหลด
	if _, err := os.Stat(cfg.App.DownloadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.App.DownloadDir, 0755); err != nil {
			log.Printf("ไม่สามารถสร้างโฟลเดอร์ %s ได้: %v", cfg.App.DownloadDir, err)
			return
		}
	}

	// ดาวน์โหลดและประมวลผลไฟล์ตาม prefix
	filesProcessed := 0
	for _, prefix := range cfg.App.FileTypes {
		log.Printf("กำลังค้นหาไฟล์ล่าสุดที่ขึ้นต้นด้วย %s...", prefix)

		// ค้นหาไฟล์ล่าสุดตาม prefix
		latestFile, err := sftpClient.FindLatestFileByPrefix(prefix)
		if err != nil {
			log.Printf("ข้อผิดพลาด: %v", err)
			continue
		}

		// ตรวจสอบว่าไฟล์นี้เคยประมวลผลแล้วหรือไม่
		processed, err := db.CheckFileProcessed(latestFile.Name())
		if err != nil {
			log.Printf("ไม่สามารถตรวจสอบประวัติการประมวลผลไฟล์ได้: %v", err)
		} else if processed {
			log.Printf("ไฟล์ %s เคยประมวลผลแล้ว ข้ามไปไฟล์ถัดไป", latestFile.Name())
			continue
		}

		log.Printf("พบไฟล์ล่าสุด: %s", latestFile.Name())

		// สร้าง path ของไฟล์
		remoteFilePath := cfg.SFTP.RemotePath + "/" + latestFile.Name()
		localFilePath := filepath.Join(cfg.App.DownloadDir, latestFile.Name())

		// ดาวน์โหลดไฟล์
		log.Printf("กำลังดาวน์โหลดไฟล์ %s...", latestFile.Name())
		bytesDownloaded, err := sftpClient.DownloadFile(remoteFilePath, localFilePath)
		if err != nil {
			log.Printf("ไม่สามารถดาวน์โหลดไฟล์ได้: %v", err)
			if err := db.LogFileProcessing(latestFile.Name(), "failed", 0, "ไม่สามารถดาวน์โหลดไฟล์ได้: "+err.Error()); err != nil {
				log.Printf("ไม่สามารถบันทึกประวัติการประมวลผลได้: %v", err)
			}
			continue
		}
		log.Printf("ดาวน์โหลดไฟล์ %s สำเร็จ (%d bytes)", latestFile.Name(), bytesDownloaded)

		// ประมวลผลไฟล์
		log.Printf("กำลังประมวลผลไฟล์ %s...", latestFile.Name())
		recordCount, err := process.ProcessFile(db, localFilePath)
		if err != nil {
			log.Printf("ไม่สามารถประมวลผลไฟล์ได้: %v", err)
			if err := db.LogFileProcessing(latestFile.Name(), "failed", 0, "ไม่สามารถประมวลผลไฟล์ได้: "+err.Error()); err != nil {
				log.Printf("ไม่สามารถบันทึกประวัติการประมวลผลได้: %v", err)
			}
		} else {
			log.Printf("ประมวลผลไฟล์ %s สำเร็จ (%d รายการ)", latestFile.Name(), recordCount)
			if err := db.LogFileProcessing(latestFile.Name(), "success", recordCount, ""); err != nil {
				log.Printf("ไม่สามารถบันทึกประวัติการประมวลผลได้: %v", err)
			}
			filesProcessed++
		}

		// ลบไฟล์ที่ดาวน์โหลดมาเมื่อประมวลผลเสร็จแล้ว
		log.Printf("กำลังลบไฟล์ %s...", localFilePath)
		if err := os.Remove(localFilePath); err != nil {
			log.Printf("ไม่สามารถลบไฟล์ได้: %v", err)
		} else {
			log.Printf("ลบไฟล์ %s สำเร็จ", localFilePath)
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("การประมวลผลเสร็จสิ้น ประมวลผลไฟล์ %d รายการ ใช้เวลา %s", filesProcessed, duration)
}