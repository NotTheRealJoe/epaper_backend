package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
	"github.com/nottherealjoe/epaper_backend"
)

func setUpHandlers(router *mux.Router, handlerHolder epaper_backend.HandlerHolder) {
	// Pi Handlers
	router.HandleFunc("/disp/auth-qr", handlerHolder.AuthQRHandlerFunc)

	// Web Handlers
	router.HandleFunc("/", handlerHolder.RootHandlerFunc)
	router.HandleFunc("/favicon.ico", handlerHolder.FaviconHandler)
	router.PathPrefix("/static").HandlerFunc(handlerHolder.StaticContentHandlerFunc)
	// Web API handlers
	router.HandleFunc("/api/drawing", handlerHolder.UploadImageHandlerFunc)
}

func main() {
	config := loadConfigFile()

	// connect to mariadb
	mysqlConfig := mysql.Config{
		User:                 config.DBUsername,
		Passwd:               config.DBPassword,
		Net:                  "tcp",
		Addr:                 config.DBHost + ":" + config.DBPort,
		DBName:               config.DBName,
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatal(fmt.Errorf("%s :: %w", "Failed to connect to database.", err))
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	repo := epaper_backend.CreateMysqlRepository(db)
	if !repo.CheckConnection() {
		log.Fatal("Failed to connect to database! Connection step passed, but verify string didn't match.")
	}

	mqttClient := epaper_backend.NewMQTTClient(&repo, &config)

	// start web server
	handler := epaper_backend.NewHandlerHolder(&repo, &config, &mqttClient)
	router := mux.NewRouter()
	setUpHandlers(router, handler)
	listenPort := strconv.Itoa(config.ListenPort)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + listenPort,
	}
	log.Println("Server listening on " + listenPort + "...")
	log.Fatal(server.ListenAndServe())
}

func loadConfigFile() epaper_backend.Config {
	config := epaper_backend.Config{}

	configFileRaw, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(fmt.Errorf(
			"%s :: %w",
			"Reminder: config.json must be in the directory you run the binary from",
			err,
		))
	}

	err = json.Unmarshal(configFileRaw, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}
