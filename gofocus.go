package main

import (
	"fmt"
	"net/http"
    "database/sql"
    "encoding/json"

    _ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Id			string
	Name		string
	Parent		string
	Completed	string
	Due			string
}

func idHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/id/"):]

	t, err := loadId(id)
	if ( err != nil ) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	if err := enc.Encode(&t); err != nil {
		fmt.Fprintf(w, "%s", err)
	}
}

func loadId(id string) (*Task, error) {
    database, _ := sql.Open("sqlite3", "file:./OmniFocusDatabase2?mode=ro")
    // database, _ := sql.Open("sqlite3", "./OmniFocusDatabase2")

    rows, err := database.Query("SELECT persistentIdentifier, name, parent, dateCompleted, dateDue FROM task where persistentIdentifier = ?", id)
    if err != nil {
		return nil, err
    }

    var persistentIdentifier string
    var name string
    var parent string
    var dateCompleted string
    var dateDue string

    rows.Next();
	rows.Scan(&persistentIdentifier, &name, &parent, &dateCompleted, &dateDue)
	t := &Task{Id: persistentIdentifier, Name: name, Parent: parent, Completed: dateCompleted, Due: dateDue}
	// fmt.Fprintf(w, "ID:'%s', NAME:'%s'\n", t.Id, t.Name)

	return t, nil
}

func main() {
    http.HandleFunc("/id/", idHandler)
    http.ListenAndServe(":8080", nil)
}
