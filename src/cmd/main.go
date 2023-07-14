package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	router.HandleFunc("/admin", handlerHolder.AdminPanelHandlerFunc)
	// Web API handlers
	router.HandleFunc("/api/drawing", handlerHolder.UploadImageHandlerFunc)
	router.HandleFunc("/admin/api/drawings", handlerHolder.AdminGetDrawingsHandlerFunc)
	router.PathPrefix("/admin/api/drawing").HandlerFunc(handlerHolder.AdminGetDrawingDataHandlerFunc)
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
	waitForDBToWork(db)

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

func waitForDBToWork(db *sql.DB) {
	err := db.Ping()
	if err == nil {
		return
	}
	for tries := 0; tries < 30; tries++ {
		if !strings.Contains(err.Error(), "connection refused") {
			log.Fatal(fmt.Errorf("%s :: %w", "database ping failed", err))
		}

		log.Print("Failed to ping database, waiting one second")
		time.Sleep(time.Second)

		if err = db.Ping(); err == nil {
			return
		}
	}
	log.Fatal("did not ping database in time")
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

	config.Admin.PasswordHash, err = base64.StdEncoding.DecodeString(config.Admin.PasswordHashBase64)
	if err != nil {
		log.Fatal(err)
	}

	return config
}
