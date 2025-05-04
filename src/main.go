package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
)

func main() {
	flagAPIKey := flag.String("api-key", "", "initial API key to start the service")
	flagComment := flag.String("comment", "", "comment for initial API key")
	flagConfigPath := flag.String("config", "", "config file path")

	flag.Parse()

	if *flagConfigPath == "" {
		log.Fatalf("Please provide a config file using the parameter --config.")
	}

	cfg, err := config.LoadConfig(*flagConfigPath)
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

	// add api-key if passed
	if *flagAPIKey != "" {
		_, err := db.SaveAPIKey(*flagAPIKey, *flagComment)
		if err != nil {
			log.Fatalf("Error saving initial API key: %v", err)
		}
		log.Printf("Initial API key was added.")
	}

	if ok, _ := db.AnyAPIKeyExists(); !ok {
		log.Fatalf("No API key exists. Please provide an API key when starting the service by using the parameters --api-key and --comment.")
	}

	_, err = db.SaveFile("test.txt", false, "123")
	if err != nil {
		log.Printf("File NOT stored: %v\n", err)
	}
	log.Printf("Testfile stored!")

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
