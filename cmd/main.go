package main

import (
	"fmt"
	"log"
	"net/http"

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
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully created connection to database")

	fetchService := fetch.NewFetchService("https://cran.r-project.org/src/contrib/", 50)
	pkgs := fetchService.FetchPkgList()
	db.InsertPackages(pkgs)
	fmt.Println("seed all packages information into DB")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!!")
	})
	log.Fatal(http.ListenAndServe(":5000", nil))
}
