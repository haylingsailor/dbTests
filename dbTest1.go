// Package test1 is a simple test
package test1

import (
	"database/sql"
	// blank import because sqlite3 example has it!
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

// Test1 is my test suite object
type Test1 struct {
	dbFile string
	db     *DB
}

func (t *Test1) init(dbFile string) {
	t.dbFile = dbFile
	os.Remove(t.dbFile)
	db, err := sql.Open("sqlite3", t.dbFile)
	t.db = db
}

func (t *Test1) end() {
	db.Close()
}

func (t *Test1) run() {

	sqlStmt := `
        CREATE TABLE IF NOT EXISTS foo (
            id    INTEGER NOT NULL PRIMARY KEY,
            name  TEXT);
        DELETE FROM foo;
        `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
}
