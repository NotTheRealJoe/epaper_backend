package epaper_backend

type Config struct {
	ListenPort          int        `json:"listenPort"`
	DBHost              string     `json:"dbHost"`
	DBPort              string     `json:"dbPort"`
	DBUsername          string     `json:"dbUsername"`
	DBPassword          string     `json:"dbPassword"`
	DBName              string     `json:"dbName"`
	PublicBasePath      string     `json:"publicBasePath"`
	EPaperDisplayHeight int        `json:"ePaperDisplayHeight"`
	StaticContentPath   string     `json:"staticContentPath"`
	TemplatesPath       string     `json:"templatesPath"`
	MQTT                MQTTConfig `json:"mqtt"`
}

type MQTTConfig struct {
	BrokerAddress string `json:"brokerAddress"`
	BrokerPort    int    `json:"brokerPort"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Prefix        string `json:"prefix"`
	CAFile        string `json:"caFile"`
}

type Authorzation struct {
	ID         int
	AuthCode   string
	UseStarted *string
	UserCookie *string
}

type Drawing struct {
	ID            int64
	DateCreated   string
	Author        string
	Data          []byte
	Authorization string
}
