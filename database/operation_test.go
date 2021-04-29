package database

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ingemar0720/cran_retriver/fetch"
	"github.com/jmoiron/sqlx"
)

var pkgs = []fetch.Package{
	{
		Name:            "name1",
		Version:         "version1",
		MD5sum:          "md5sum1",
		DatePublication: "2000-01-01 00:00:00 UTC",
		Title:           "Title1",
		Description:     "Description1",
		Author:          fetch.Developer{Name: "Author1", Email: "Email1"},
		Maintainer:      fetch.Developer{Name: "Maintainer1", Email: "Email2"},
	},
}

type Any struct{}

func (a Any) Match(v driver.Value) bool {
	return true
}

func TestDB_InsertPackages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO developers (.+) RETURNING id").WithArgs(pkgs[0].Maintainer.Name, pkgs[0].Maintainer.Email).WillReturnRows((sqlmock.NewRows([]string{"id"}).AddRow(1)))
	mock.ExpectQuery("INSERT INTO developers (.+) RETURNING id").WithArgs(pkgs[0].Author.Name, pkgs[0].Author.Email).WillReturnRows((sqlmock.NewRows([]string{"id"}).AddRow(2)))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO packages").WithArgs(pkgs[0].Name, pkgs[0].Version, pkgs[0].MD5sum, Any{}, pkgs[0].Title, pkgs[0].Description, 2, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	database := DB{sqlxDB}
	database.InsertPackages(pkgs)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_strToNullString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  sql.NullString
	}{
		{
			name:  "happy path",
			input: "valid_input",
			want:  sql.NullString{String: "valid_input", Valid: true},
		},
		{
			name:  "fail path",
			input: "",
			want:  sql.NullString{Valid: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strToNullString(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strToNullString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_strToNullTime(t *testing.T) {
	fixtureTime := "2000-01-01 00:00:00 UTC"
	parsedTime, _ := time.Parse("2006-01-02 15:04:05 UTC", fixtureTime)
	type args struct {
		t string
	}
	tests := []struct {
		name  string
		input string
		want  sql.NullTime
	}{
		{
			name:  "happy path",
			input: fixtureTime,
			want:  sql.NullTime{Time: parsedTime, Valid: true},
		},
		{
			name:  "fail path",
			input: "",
			want:  sql.NullTime{Valid: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strToNullTime(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strToNullTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
