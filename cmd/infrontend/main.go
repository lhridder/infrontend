package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"infrontend"
	"log"
	"net/http"
)

func main() {
	err := infrontend.LoadGlobalConfig()
	if err != nil {
		log.Printf("Failed to load config: %s", err)
		return
	}

	err = infrontend.ConnectRedis()
	if err != nil {
		log.Printf("Failed connecting to redis: %s", err)
		return
	}

	r := chi.NewRouter()
	if infrontend.GlobalConfig.Debug {
		r.Use(middleware.Logger)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/client", http.StatusFound)
	})

	r.Get("/login", infrontend.GetLogin)
	r.Post("/login", infrontend.PostLogin)

	r.Get("/register", infrontend.GetRegister)
	r.Post("/register", infrontend.PostRegister)

	r.Route("/client", func(r chi.Router) {
		r.Use(infrontend.Auth())
		r.Get("/", infrontend.GetClientHome)
		r.Post("/logout", infrontend.PostLogout)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(infrontend.Auth())
		r.Use(infrontend.Admin())
		r.Get("/", infrontend.GetAdminHome)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(infrontend.Auth())
	})

	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	listen := infrontend.GlobalConfig.Listen
	log.Printf("Starting web listener on %s", listen)
	panic(http.ListenAndServe(listen, r))
}
