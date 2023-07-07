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
	"github.com/nottherealjoe/epaper_backend/repository"
)

func setUpHandlers(router *mux.Router, handler epaper_backend.HandlerHolder) {
	router.HandleFunc("/", handler.RootHandlerFunc)
	router.HandleFunc("/api/disp/auth-qr", handler.AuthQRHandlerFunc)

	router.PathPrefix("/static").HandlerFunc(handler.StaticContentHandlerFunc)
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
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.CreateMysqlRepository(db)
	if !repo.CheckConnection() {
		log.Fatal("Failed to connect to database!")
	}

	// start web server
	handler := epaper_backend.NewHandler(&repo, &config)
	router := mux.NewRouter()
	setUpHandlers(router, handler)
	listenPort := strconv.Itoa(config.ListenPort)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + listenPort,
	}
	println("Server listening on " + listenPort + "...")
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
