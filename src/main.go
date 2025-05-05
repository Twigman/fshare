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

	log.Printf("Config loaded\n")

	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Error loading sqlite: %v", err)
	}

	apiKeyService := store.NewAPIKeyService(db)

	// add api-key if provided
	if *flagAPIKey != "" {
		_, err := apiKeyService.AddAPIKey(*flagAPIKey, *flagComment)
		if err != nil {
			log.Fatalf("Error saving initial API key: %v", err)
		}
		log.Printf("Initial API key was added.")
	}

	if ok, _ := db.AnyAPIKeyExists(); !ok {
		log.Fatalf("No API key exists. Please provide an API key when starting the service by using the parameters --api-key and --comment.")
	}

	fileService := store.NewFileService(cfg.UploadPath, db)

	if err := startServer(cfg, apiKeyService, fileService); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func startServer(cfg *config.Config, apiKeyService *store.APIKeyService, fileService *store.FileService) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting server on %s ...\n", addr)

	restService := httpapi.NewRESTService(cfg, apiKeyService, fileService)

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", restService.UploadHandler)

	return http.ListenAndServe(addr, mux)
}
