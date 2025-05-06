package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

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

	/******************************
	 * config
	 ******************************/
	cfg, err := config.LoadConfig(*flagConfigPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err = cfg.Validate(); err != nil {
		log.Fatalf("Config validation error: %v", err)
	}

	log.Printf("Config loaded\n")

	/******************************
	 * db
	 ******************************/
	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Error loading sqlite: %v", err)
	}

	/******************************
	 * api key + home dir
	 ******************************/
	apiKeyService := store.NewAPIKeyService(db)
	fileService := store.NewFileService(cfg, db)

	// add api-key if provided
	if *flagAPIKey != "" {
		key, err := apiKeyService.AddAPIKey(*flagAPIKey, *flagComment)
		if err != nil {
			log.Fatalf("Error saving initial API key: %v", err)
		}
		log.Printf("Initial API key was added.")

		r, err := fileService.GetOrCreateHomeDir(key.HashedKey)
		if err != nil {
			log.Fatalf("Error creating home dir: %v", err)
		}

		log.Printf("Created home dir: %s\n", filepath.Join(cfg.UploadPath, r.Name))
	}

	if ok, _ := apiKeyService.AnyAPIKeyExists(); !ok {
		log.Fatalf("No API key exists. Please provide an API key when starting the service by using the parameters --api-key and --comment.")
	}

	/******************************
	 * start service
	 ******************************/
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
