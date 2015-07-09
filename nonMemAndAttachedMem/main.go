package main

import (
	"database/sql"
	"fmt"
	sqlx "github.com/jmoiron/sqlx"
	"log"
	"time"
	//Driver for sqlite
	sqlite "github.com/mattn/go-sqlite3"
)

// In this test, we create a disk-based database and a memory only one (using
// the ATTACH sql command)
//
// Also, I tested connection hooks (though we have no reason to use them at
// present)
//
func main() {

	sql3Conn := []*sqlite.SQLiteConn{}
	sql.Register("sqlite3ConnectionCatchingDriver",
		&sqlite.SQLiteDriver{
			ConnectHook: func(newConn *sqlite.SQLiteConn) error {
				sql3Conn = append(sql3Conn, newConn)
				return nil
			},
		},
	)

	db, err := sqlx.Connect("sqlite3ConnectionCatchingDriver", "myDb.sqlite?cache=shared")
	if err != nil {
		log.Fatal("Connect: ", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS main.person (
            id INTEGER PRIMARY KEY,
            name TEXT
        );

        ATTACH DATABASE 'file::memory:?cache=shared' AS mem;

        CREATE TABLE IF NOT EXISTS mem.sessionActivity (
            id INTEGER PRIMARY KEY,
            personId INTEGER NOT NULL,
            dateTime DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)

	if err != nil {
		log.Fatal("Schema Build: ", err)
	}

	insertPerson, err := db.PrepareNamed(`
        INSERT OR IGNORE INTO main.person (id, name)
        VALUES(:id, :name);
    `)

	if err != nil {
		log.Fatal("Prepare Statement 1: ", err)
	}

	insertSessionActivity, err := db.PrepareNamed(`
	       INSERT OR IGNORE INTO mem.sessionActivity (personId)
	       VALUES(:personId);
	   `)

	if err != nil {
		log.Fatal("Prepare Statement 2: ", err)
	}

	_, err = insertPerson.Exec(map[string]interface{}{
		"id":   1,
		"name": "Jim",
	})

	if err != nil {
		log.Fatal("insert person: ", err)
	}

	_, err = insertSessionActivity.Exec(map[string]interface{}{
		"personId": 1,
	})

	if err != nil {
		log.Fatal("insert session activity 1: ", err)
	}

	time.Sleep(time.Second)

	_, err = insertSessionActivity.Exec(map[string]interface{}{
		"personId": 1,
	})

	if err != nil {
		log.Fatal("insert session activity 2: ", err)
	}

	// This query is a join between a memory-db table and a disk-db table. It works!
	rows, err := db.Query(`
       SELECT
           main.person.name as personName,
           main.person.id as personId,
           mem.sessionActivity.dateTime as dateTime
       FROM mem.sessionActivity
       LEFT OUTER JOIN main.person
           ON mem.sessionActivity.personId = main.person.id
       ORDER BY
           mem.sessionActivity.dateTime ASC
   `)
	if err != nil {
		log.Fatal("Select : ", err)
	}

	// iterate over each row
	for rows.Next() {
		var personName string
		var personID int64
		var dateTime time.Time
		err = rows.Scan(&personName, &personID, &dateTime)
		fmt.Println("Result:", personID, personName, dateTime)
	}

	fmt.Println("Finished")
}
