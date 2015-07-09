package main

import (
	_ "database/sql"
	"fmt"
	"log"
	"time"
	//Driver for sqlite
	sqlite "code.google.com/p/go-sqlite/go1/sqlite3"
)

//This test tries out the go-sqlite plug-in which google provides. Unlike
//mattn/sqlite3, this does offer encrption, but we can't use the jmoiron helper
//package...
//
// Note here that for thread-safety, all go routines must separately open up
// the database, but they're also SHARING a single memory-only database. The
// magic is ?cache=shared plus the naming of the memory database ('mem').
// So in each connection, we attach the memory only database
//

// InsertPerson blah
func InsertPerson(conn *sqlite.Conn, id int, name string) (err error) {
	args := sqlite.NamedArgs{
		"$id":   id,
		"$name": name,
	}

	err = conn.Exec(
		`INSERT OR IGNORE INTO main.person (id, name)
            VALUES($id, $name);
        `, args)
	return
}

// InsertSessionActivity blah
func InsertSessionActivity(conn *sqlite.Conn, personID int) (err error) {
	args := sqlite.NamedArgs{
		"$personId": personID,
	}

	err = conn.Exec(
		`INSERT OR IGNORE INTO mem.sessionActivity (personId)
           VALUES($personId);
        `, args)
	return
}

// OpenDb blah
func OpenDb() (conn *sqlite.Conn, err error) {
	conn, err = sqlite.Open("file:myDb.sqlite?cache=shared")
	if err != nil {
		log.Fatal("Connect: ", err)
	}

	err = conn.Exec(`ATTACH DATABASE 'file::memory:?cache=shared' AS mem;`)
	return
}

// Writer Creates the schema in both mem and disk db and writes into both
func Writer() {
	db, err := OpenDb()
	if err != nil {
		log.Fatal("Connect: ", err)
	}

	err = db.Exec(`
        CREATE TABLE IF NOT EXISTS main.person (
            id INTEGER PRIMARY KEY,
            name TEXT
        );

        CREATE TABLE IF NOT EXISTS mem.sessionActivity (
            id INTEGER PRIMARY KEY,
            personId INTEGER NOT NULL,
            dateTime DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)

	if err != nil {
		log.Fatal("Schema Build: ", err)
	}

	err = InsertPerson(db, 1, "Andy")
	if err != nil {
		log.Fatal("insert person: ", err)
	}

	err = InsertSessionActivity(db, 1)
	if err != nil {
		log.Fatal("insert session activity 1: ", err)
	}

	time.Sleep(time.Second)

	err = InsertSessionActivity(db, 1)
	if err != nil {
		log.Fatal("insert session activity 2: ", err)
	}
}

//Reader reads data using a join of tables from both mem and disk database to
//show that a) memory/disk db joins work, and b) that the mem-only db has data
//in it already (which means it IS being shared, rather than being a new memory
//only db)
func Reader() {
	db, err := OpenDb()
	if err != nil {
		log.Fatal("Connect: ", err)
	}

	query := `
       SELECT
           main.person.name as personName,
           main.person.id as personId,
           mem.sessionActivity.dateTime as dateTime
       FROM mem.sessionActivity
       LEFT OUTER JOIN main.person
           ON mem.sessionActivity.personId = main.person.id
       ORDER BY
           mem.sessionActivity.dateTime ASC
   `

	row := make(sqlite.RowMap)

	for s, err := db.Query(query); err == nil; err = s.Next() {
		s.Scan(row)
		fmt.Println(row) // Prints "1 map[a:1 b:demo c:<nil>]"
	}

}

func main() {
	Writer()
	Reader()
	fmt.Println("Finished")
}
