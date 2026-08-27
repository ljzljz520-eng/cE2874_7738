package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"rehab-followup/internal/httpapi"
	"rehab-followup/internal/service"
	"rehab-followup/internal/store"
)

func main() {
	path := flag.String("db", "rehab-followup.db", "path to the local follow-up database")
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()
	s, err := store.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	platform := service.NewPlatform(s, time.Now)
	if os.Getenv("REHAB_DEMO_THERAPIST") != "" {
		_, _ = platform.RegisterTherapist("demo", "Demo Therapist", "Rehabilitation", "demo")
	}
	server := httpapi.NewServer(platform)
	log.Printf("rehab follow-up listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
