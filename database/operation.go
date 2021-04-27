package database

import (
	"fmt"

	"log"

	"github.com/ingemar0720/cran_retriver/fetch"
	"github.com/jmoiron/sqlx"
)

const (
	HOST     = "db"
	DATABASE = "postgres"
	USER     = "user"
	PASSWORD = "mysecretpassword"
)

type DB struct {
	*sqlx.DB
}

func New() (DB, error) {
	db, err := sqlx.Connect("postgres", "host=db user=user dbname=postgres password=mysecretpassword sslmode=disable")
	if err != nil {
		return DB{}, fmt.Errorf("fail to connect to db, error: %v", err)
	}
	return DB{db}, nil
}

func (db DB) InsertPackages(pkgs []fetch.Package) {
	tx := db.DB.MustBegin()
	authorIDList := make([]int, len(pkgs))
	maintainerIDList := make([]int, len(pkgs))

	for i, p := range pkgs {
		var maintainer_id int
		rows, err := tx.NamedQuery(`INSERT INTO developers (name, email) VALUES (:name, :email)
		                              ON CONFLICT (name) DO UPDATE SET email=EXCLUDED.email
									  RETURNING id`, p.Maintainer)
		if err != nil {
			log.Printf("insert maintainer into developers table fail, error %v", err)
			break
		}
		if rows.Next() {
			rows.Scan(&maintainer_id)
		}
		rows.Close()
		maintainerIDList[i] = maintainer_id
		var author_id int
		rows, err = tx.NamedQuery(`INSERT INTO developers (name, email) VALUES (:name, :email)
								    ON CONFLICT (name) DO UPDATE SET email=EXCLUDED.email
									RETURNING id`, p.Author)
		if err != nil {
			log.Printf("insert author into developers table fail, error %v", err)
			break
		}
		if rows.Next() {
			rows.Scan(&author_id)
		}
		rows.Close()
		authorIDList[i] = author_id
	}
	tx.Commit()

	tx = db.DB.MustBegin()
	for i, np := range pkgs {
		tx.MustExec(`INSERT INTO packages (name, version, md5sum, date_publication, title, description, author_id, maintainer_id)
		               VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		               ON CONFLICT (name, version) DO NOTHING`,
			np.Name, np.Version, np.MD5sum, np.DatePublication, np.Title, np.Description, authorIDList[i], maintainerIDList[i])
	}
	tx.Commit()
}
