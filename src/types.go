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
	UserCookie *string
}
