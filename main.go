package main

import (
	"fmt"
	"net/http"

	cronjobs "github.com/Arinji2/sense-backend/cron-jobs"
	"github.com/Arinji2/sense-backend/pocketbase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Sense Backend: Request Received")
		w.Write([]byte("Sense Backend: Request Received"))
		render.Status(r, 200)

		token := pocketbase.PocketbaseAdminLogin()
		fmt.Println(token)

		fmt.Println(cronjobs.GetLevel("real_words"))

	})
	http.ListenAndServe(":3000", r)
}
