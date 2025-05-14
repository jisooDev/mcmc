package config

type Config struct {
	Database DatabaseConfig
	SFTP     SFTPConfig
	App      AppConfig
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
	}
}