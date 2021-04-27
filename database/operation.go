package database

import (
	"database/sql"
	"fmt"
	"time"

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

type PackageModel struct {
	ID              int            `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	Version         string         `json:"version" db:"version"`
	MD5sum          string         `json:"md5sum" db:"md5sum"`
	DatePublication sql.NullTime   `json:"date_publication" db:"date_publication"`
	Title           sql.NullString `json:"title" db:"title"`
	Description     sql.NullString `json:"description" db:"description"`
	AuthorID        int            `json:"author_id" db:"author_id"`
	MaintainerID    int            `json:"maintainer_id" db:"maintainer_id"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
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

func (db DB) QueryPackages(name string) ([]PackageModel, error) {
	packages := []PackageModel{}
	fmt.Println("prepare to search package based on ", name)
	err := db.DB.Select(&packages, `SELECT * FROM packages WHERE name LIKE $1`, "%"+name+"%")
	if err != nil {
		fmt.Println(err)
		return []PackageModel{}, fmt.Errorf("search name from DB fail, error %v", err)
	}
	fmt.Printf("find %v of records has similar name as %v\n", len(packages), name)
	return packages, nil
}
