// Package server provides the frontend service to handle all the requests coming into the world-map application.
package server

import (
	"copilots-world-map/countrydb"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Server represents a world map server.
type Server struct {
	db countrydb.DB

	router *mux.Router
	indexTpl *template.Template
}

// New returns a new world map server that queries against the provided country db.
func New(db countrydb.DB) (*Server, error) {
	indexTpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("server: parse index.html: %v", err)
	}

	s := &Server{
		db: db,
		router: mux.NewRouter(),
		indexTpl: indexTpl,
	}
	s.routes()
	return s, nil
}

// ServeHTTP delegates to the mux router.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}


func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleIndex())
	s.router.HandleFunc("/visits", s.handleVisits()).Methods("GET")
	s.router.HandleFunc("/uniquevisits", s.handleUniqueVisits()).Methods("GET")
	s.router.HandleFunc("/visits/{country}", s.handleVisitCountry()).Methods("POST")
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count, err := s.db.UniqueTotal()
		if err != nil {
			log.Printf("ERROR: unique total: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = s.indexTpl.Execute(w, struct {
			TotalCountries int
		} {
			TotalCountries: count,
		})
		if err != nil {
			log.Printf("ERROR: execute index template: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleVisits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		countries, err := s.db.Results()
		if err != nil {
			log.Printf("ERROR: countries results: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dat, err := json.Marshal(countries)
		if err != nil {
			log.Printf("ERROR: marshal json countries: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
	}
}

func (s *Server) handleUniqueVisits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count, err := s.db.UniqueTotal()
		if err != nil {
			log.Printf("ERROR: unique total: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dat, err := json.Marshal(struct{
			Count int
		} {
			Count: count,
		})
		if err != nil {
			log.Printf("ERROR: marshal json unique count: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
	}
}

func (s *Server) handleVisitCountry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		country := mux.Vars(r)["country"]
		visit, err := s.db.Save(country)
		if err != nil {
			log.Printf("ERROR: save country '%s': %v\n", country, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("New visit to %s with visit count %d\n", country, visit)
		dat, err := json.Marshal(countrydb.Country{
			Country: country,
			Visit: visit,
		})
		if err != nil {
			log.Printf("ERROR: marshal json countries: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
	}
}