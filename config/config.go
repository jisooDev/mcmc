package config

import "time"

type Config struct {
	Database DatabaseConfig
	SFTP     SFTPConfig
	App      AppConfig
	Cron     CronConfig
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type SFTPConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	RemotePath string
}

type AppConfig struct {
	DownloadDir string
	FileTypes   []string
}

type CronConfig struct {
	Schedule      string        // Cron expression สำหรับตั้งเวลาทำงาน (เช่น "0 */4 * * *" = ทุก 4 ชั่วโมง)
	RunOnce       bool          // true = รันครั้งเดียวแล้วจบ, false = รันเป็น cron
	RetryInterval time.Duration // เวลาที่จะลองใหม่ถ้าการเชื่อมต่อล้มเหลว
}

func GetConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Driver:   "sqlserver",
			Host:     "SLEAPOTCDEVST01.thaibev.com",
			Port:     "1433",
			User:     "mcmc_user",
			Password: "mcmc$3rM1ddleware",
			DBName:   "MCMC_Middleware",
		},
		SFTP: SFTPConfig{
			Host:       "10.7.57.119",
			Port:       "22",
			User:       "mcmcvendoruser",
			Password:   "KPHaE8dmZA0baWzU",
			RemotePath: "VSMSUAT/back",
		},
		App: AppConfig{
			DownloadDir: "./downloaded_files",
			FileTypes: []string{
				"saleorder_summary_",
				"saleorder_item_",
				"saleorder_header_",
			},
		},
		Cron: CronConfig{
			Schedule:      "*/5 * * * *", 
			RunOnce:       false,          
			RetryInterval: 5 * time.Minute, 
		},
	}
}