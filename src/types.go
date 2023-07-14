package epaper_backend

type Config struct {
	ListenPort          int         `json:"listenPort"`
	DBHost              string      `json:"dbHost"`
	DBPort              string      `json:"dbPort"`
	DBUsername          string      `json:"dbUsername"`
	DBPassword          string      `json:"dbPassword"`
	DBName              string      `json:"dbName"`
	PublicBasePath      string      `json:"publicBasePath"`
	EPaperDisplayHeight int         `json:"ePaperDisplayHeight"`
	StaticContentPath   string      `json:"staticContentPath"`
	TemplatesPath       string      `json:"templatesPath"`
	MQTT                MQTTConfig  `json:"mqtt"`
	Admin               AdminConfig `json:"admin"`
}

type MQTTConfig struct {
	BrokerAddress     string `json:"brokerAddress"`
	BrokerTLSHostname string `json:"brokerTlsHostname"`
	BrokerPort        int    `json:"brokerPort"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Prefix            string `json:"prefix"`
	CAFile            string `json:"caFile"`
}

type AdminConfig struct {
	Username           string `json:"username"`
	PasswordHashBase64 string `json:"passwordHash"`
	PasswordHash       []byte
}

type Authorzation struct {
	ID         int
	AuthCode   string
	UseStarted *string
	UserCookie *string
}

type Drawing struct {
	ID            int64  `json:"id"`
	DateCreated   string `json:"dateCreated"`
	Author        string `json:"author"`
	Data          []byte `json:"-"`
	Authorization string `json:"-"`
	Removed       bool   `json:"removed"`
}

type SubmitDrawingRequest struct {
	Artist       string `json:"artist"`
	ImageDataURL string `json:"imageDataUrl"`
}
