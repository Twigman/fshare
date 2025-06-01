package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/twigman/fshare/src/config"
	"github.com/twigman/fshare/src/httpapi"
	"github.com/twigman/fshare/src/store"
	"github.com/twigman/fshare/src/utils"
)

func main() {
	flagAPIKey := flag.String("api-key", "", "initial API key to start the service")
	flagComment := flag.String("comment", "", "comment for initial API key")
	flagHighlyTrusted := flag.Bool("highly-trusted", false, "more privileges for the key user")
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
	db, err := store.NewDB(cfg.DataPath)
	if err != nil {
		log.Fatalf("Error loading sqlite: %v", err)
	}

	/******************************
	 * api key + home dir
	 ******************************/
	as := store.NewAPIKeyService(db)
	rs := store.NewResourceService(cfg, db)

	// create upload folder
	err = rs.CreateUploadDir()
	if err != nil {
		log.Fatalf("Error creating upload folder: %v", err)
	}

	// add api-key if provided
	if *flagAPIKey != "" {
		key, err := as.AddAPIKey(*flagAPIKey, *flagComment, *flagHighlyTrusted, nil)
		if err != nil {
			log.Fatalf("Error saving initial API key: %v", err)
		}
		log.Printf("Initial API key was added.")

		r, err := rs.GetOrCreateHomeDir(key.HashedKey)
		if err != nil {
			log.Fatalf("Error creating home dir: %v", err)
		}

		log.Printf("Created home dir: %s\n", filepath.Join(cfg.UploadPath, r.Name))
	}

	if !as.AnyAPIKeyExists() {
		// generate first API key
		log.Printf("No API key exists and no one was provided as a parameter.")
		key, err := utils.GenerateSecret(32)
		if err != nil {
			log.Fatalf("Could not generate key: %v", err)
		}

		_, err = as.AddAPIKey(key, "initial key", true, nil)
		if err != nil {
			log.Fatalf("Could not register key")
		}

		initPath, err := config.CreateInitDataEnv(cfg.DataPath, key)
		if err != nil {
			log.Fatalf("Could not save initial key: %v", err)
		}

		log.Printf("Initial API key generated and saved to: %s\n", initPath)
	}

	/******************************
	 * start service
	 ******************************/
	if err := startServer(cfg, as, rs); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func startServer(cfg *config.Config, as *store.APIKeyService, rs *store.ResourceService) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting server on %s ...\n", addr)

	restService := httpapi.NewRESTService(cfg, as, rs)

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", restService.UploadHandler)
	mux.HandleFunc("/r/", restService.ResourceHandler)
	mux.HandleFunc("/delete/", restService.DeleteHandler)
	mux.HandleFunc("/raw/", restService.RawResourceHandler)
	mux.HandleFunc("/apikey", restService.CreateAPIKeyHandler)

	// start cleanup worker for autodelete
	stopCh := make(chan struct{})
	go rs.StartCleanupWorker(time.Duration(cfg.AutoDeleteIntervalInSec)*time.Second, stopCh)

	// handle graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill)
	go func() {
		<-signalCh
		log.Println("Shutdown signal received, stopping cleanup worker...")
		close(stopCh)
		os.Exit(0)
	}()

	return http.ListenAndServe(addr, mux)
}
