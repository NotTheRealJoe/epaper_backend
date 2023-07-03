package epaper_backend

type Config struct {
	ListenPort int    `json:"listenPort"`
	DBHost     string `json:"dbHost"`
	DBPort     string `json:"dbPort"`
	DBUsername string `json:"dbUsername"`
	DBPassword string `json:"dbPassword"`
	DBName     string `json:"dbName"`
}
