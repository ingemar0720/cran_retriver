package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

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
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", HOST, USER, PASSWORD, DATABASE),
	)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Successfully created connection to database")

	fetchService := fetch.NewFetchService("https://cran.r-project.org/src/contrib/", 50)
	pkgs := fetchService.FetchPkgList()
	fmt.Println(len(pkgs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!!")
	})
	log.Fatal(http.ListenAndServe(":5000", nil))
}
