package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
)

func main() {
	cfg, err := config.LoadConfig("../data/config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err = cfg.Validate(); err != nil {
		log.Fatalf("Config validation error: %v", err)
	}

	log.Printf("Config loaded. Port: %d, UploadPath: %s\n, MaxFileSizeInMB: %d\n", cfg.Port, cfg.UploadPath, cfg.MaxFileSizeInMB)

	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Error loading sqlite: %v", err)
	}

	_, err = db.SaveFile("test.txt", false, "abc")
	if err != nil {
		fmt.Printf("File NOT stored: %v\n", err)
	}

	if err := startServer(cfg); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func startServer(cfg *config.Config) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Starting server on %s ...\n", addr)

	handler := httpapi.NewHTTPHandler(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", handler.UploadHandler)

	return http.ListenAndServe(addr, mux)
}
