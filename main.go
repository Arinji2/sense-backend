package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	cronjobs "github.com/Arinji2/sense-backend/cron-jobs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Sense Backend: Request Received")
		w.Write([]byte("Sense Backend: Request Received"))
		render.Status(r, 200)
	})

	r.Get("/env", func(w http.ResponseWriter, r *http.Request) {
		keys := os.Environ()

		for _, key := range keys {
			val := strings.Split(key, "=")

			fmt.Println(val[0] + ": " + val[1])
		}

		render.Status(r, 404)
	})

	go startCronJob()

	http.ListenAndServe(":8080", r)
}

func startCronJob() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {

		fmt.Println("Running hourly cron job")
		cronjobs.InsertWords()
		cronjobs.ResetWords()

	}
}
