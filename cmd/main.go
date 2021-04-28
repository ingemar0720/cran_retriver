package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ingemar0720/cran_retriver/database"
	"github.com/ingemar0720/cran_retriver/fetch"
	_ "github.com/lib/pq"
)

const (
	HOST     = "db"
	DATABASE = "postgres"
	USER     = "user"
	PASSWORD = "mysecretpassword"
)

func main() {
	numbefOfPkgs, err := strconv.Atoi(os.Getenv("numbefOfPkgs"))
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully created connection to database")
	fetchService := fetch.NewFetchService("https://cran.r-project.org/src/contrib/", numbefOfPkgs)
	pkgs := fetchService.FetchPkgList()
	db.InsertPackages(pkgs)
	fmt.Println("seed all packages information into DB")

	r := chi.NewRouter()
	r.Get("/packages", func(w http.ResponseWriter, r *http.Request) {
		packageName := r.URL.Query().Get("name")
		if packageName != "" {
			foundPkgs, err := db.QueryPackages(packageName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			buf, err := json.Marshal(foundPkgs)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(buf)
		}
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!!")
	})
	log.Fatal(http.ListenAndServe(":5000", r))
}
