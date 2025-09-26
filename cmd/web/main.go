package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bojkrstic/internal/db"
	"github.com/bojkrstic/internal/handlers"
	tmplpkg "github.com/bojkrstic/internal/templates"
)

func main() {
	// initialize DB
	// dsn := "root:root@tcp(127.0.0.1:3308)/sys_core?parseTime=true"
	dsn := "root:root@tcp(127.0.0.1:3308)/bulk_gate?parseTime=true"
	// dsn := "dev-stg-krle:eFQr_dCf2=kXi3_0@tcp(192.168.209.103:6033)/bulk_gate?parseTime=true"  //stg-stage
	// dsn := "dev-krle:j4tvgmXf($H-@tcp(192.168.209.103:6033)/bulk_gate?parseTime=true" //hag

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
	http.ListenAndServe(":8086", r)
}
