package main

import (
	"fmt"
	"log"
	"net/http"

	"evidence-provider/internal/service/tlscertificate"
	transport "evidence-provider/internal/transport/http"
)

func main() {
	// Generate TLS certificate at startup.
	if err := tlscertificate.InitCertificate(); err != nil {
		log.Printf("TLS certificate generation failed: %v", err)
	} else {
		log.Printf("TLS certificate fingerprint: %s...", tlscertificate.GetFingerprint()[:32])
	}

	router := transport.NewRouter()
	handler := transport.WithCORS(router)

	go func() {
		log.Println("📊 pprof at :6060/debug/pprof/")
		_ = http.ListenAndServe(":6060", nil)
	}()

	port := 8080
	log.Printf("Evidence Provider running on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
