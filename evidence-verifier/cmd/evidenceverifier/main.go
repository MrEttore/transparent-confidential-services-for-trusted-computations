package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* handlers :contentReference[oaicite:6]{index=6}

	transport "evidence-verifier/internal/transport/http"
)

func main() {
	router := transport.NewRouter()
	handler := transport.WithCORS(router)

	go func() {
		log.Println("📊 pprof at :6060/debug/pprof/")
		_ = http.ListenAndServe(":6060", nil)
	}()

	port := 8081
	log.Printf("Evidence Verifier running on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
