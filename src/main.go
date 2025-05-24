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
	db, err := store.NewDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Error loading sqlite: %v", err)
	}

	/******************************
	 * api key + home dir
	 ******************************/
	as := store.NewAPIKeyService(db)
	rs := store.NewResourceService(cfg, db)

	// add api-key if provided
	if *flagAPIKey != "" {
		key, err := as.AddAPIKey(*flagAPIKey, *flagComment, *flagHighlyTrusted)
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

	if ok, _ := as.AnyAPIKeyExists(); !ok {
		log.Fatalf("No API key exists. Please provide an API key when starting the service by using the parameters --api-key and --comment.")
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

	return http.ListenAndServe(addr, mux)
}
