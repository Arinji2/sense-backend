package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	cronjobs "github.com/Arinji2/sense-backend/cron-jobs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

func SkipLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		middleware.Logger(next).ServeHTTP(w, r)
	})
}

func main() {
	r := chi.NewRouter()
	r.Use(SkipLoggingMiddleware)

	err := godotenv.Load()
	if err != nil {
		isProduction := os.Getenv("ENVIRONMENT") == "PRODUCTION"
		if !isProduction {
			log.Fatal("Error loading .env file")
		} else {
			fmt.Println("Using Production Environment")
		}
	} else {
		fmt.Println("Using Development Environment")
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query()["key"]
		if len(key) != 0 {
			if key[0] == os.Getenv("ACCESS_KEY") {
				fmt.Println("RUNNING TASKS")
				cronjobs.InsertWords()
				cronjobs.ResetWords()
			}
		}
		fmt.Println("Sense Backend: Request Received")
		w.Write([]byte("Sense Backend: Request Received"))
		render.Status(r, 200)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("Sense Backend: Health Check"))
		render.Status(r, 200)
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
