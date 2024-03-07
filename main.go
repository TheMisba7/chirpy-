package main

import (
	"chirpy/storage"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

var newStorage = storage.NewStorage("database.json")

type apiConfig struct {
	fileServerHits int
	error          string
}

func newChirp() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		chirp := storage.NewChirp()
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(chirp)
		if err != nil {
			panic(err)
		}
		add := newStorage.Add(*chirp)
		bytes, err := json.Marshal(add)
		if err != nil {
			panic(err)
		}
		writer.Write(bytes)
	}
}
func getById() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		param := chi.URLParam(request, "chirpId")
		intV, err := strconv.Atoi(param)
		if err != nil {
			panic(err)
		}
		chirp, ok := newStorage.GetById(intV)
		if ok {
			data, _ := json.Marshal(chirp)
			writer.Write(data)
		} else {
			data, _ := json.Marshal(apiConfig{error: "No found"})
			writer.Write(data)
		}
	}
}
func (apiConfig *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiConfig.fileServerHits++
		next.ServeHTTP(w, r)
	})
}

func (apiConfig *apiConfig) metricHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		hits := fmt.Sprintf("Hits: %v", apiConfig.fileServerHits)
		writer.Write([]byte(hits))
	}
}
func (apiConfig *apiConfig) resetHandler() http.HandlerFunc {
	apiConfig.fileServerHits = 0
	return func(writer http.ResponseWriter, request *http.Request) {
		hits := fmt.Sprintf("Hits: %v	", apiConfig.fileServerHits)
		writer.Write([]byte(hits))
	}
}
func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func health() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Write([]byte("OK"))
	}
}

func main() {

	conf := &apiConfig{}
	chiRouter := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Get("/metrics", conf.metricHandler())
	apiRouter.Get("/reset", conf.resetHandler())
	apiRouter.Post("/chirps", newChirp())
	apiRouter.Get("/chirps/{chirpId}", getById())
	apiRouter.Handle("/healthz", health())
	chiRouter.Handle("/app", conf.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	chiRouter.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	chiRouter.Handle("/app/*", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))

	chiRouter.Mount("/api", apiRouter)
	http.ListenAndServe(":8888", chiRouter)

}
