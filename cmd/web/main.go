package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"github.com/bojkrstic/internal/db"
	"github.com/bojkrstic/internal/handlers"
	tmplpkg "github.com/bojkrstic/internal/templates"
)

func main() {
	dsn := loadDSN()

	conn, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer conn.Close()

	// load templates
	tmpl, err := tmplpkg.Load("templates/*.html")
	if err != nil {
		log.Fatalf("load templates: %v", err)
	}

	r := mux.NewRouter()
	r.Handle("/", handlers.NewHomeHandler(conn, tmpl))
	r.Handle("/mnp-gate", handlers.NewMnpGatePageHandler(conn, tmpl))
	r.Handle("/mnp-gates", handlers.NewMnpGatesListHandler(conn, tmpl))
	fmt.Println("Listening on :8086")
	if err := http.ListenAndServe(":8086", r); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func loadDSN() string {
	// First let explicit environment configuration win.
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}

	// Fallback to a local .env file to simplify developer setup.
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal("DB_DSN environment variable is not set and .env file could not be opened")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "DB_DSN=") {
			value := strings.TrimPrefix(line, "DB_DSN=")
			value = strings.Trim(value, "\"'")
			if value != "" {
				return value
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("reading .env file: %v", err)
	}
	log.Fatal("DB_DSN is not configured; set environment variable or add DB_DSN to .env")
	return ""
}
