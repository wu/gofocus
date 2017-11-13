package main

import (
	"fmt"
	"net/http"
    "database/sql"
    "encoding/json"
	"os/exec"
	"strings"

    _ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Id			string
	Name		string
	Parent		string
	Completed	string
	Due			string
}

var db = openDB()

var dbfile = "/Users/wu/Library/Containers/com.omnigroup.OmniFocus2.MacAppStore/Data/Library/Caches/com.omnigroup.OmniFocus2.MacAppStore/OmniFocusDatabase2"

func openDB() (*sql.DB) {
	fmt.Println("Opening database connection...")
	mydb, _ := sql.Open("sqlite3", "file:" + dbfile + "?mode=ro")
	return mydb
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	parent := r.FormValue("parent")
	deferDate := r.FormValue("defer")
	dueDate := r.FormValue("due")

	if parent == "" {
		http.Error(w, "ERROR: parent not specified", http.StatusInternalServerError)
		return
	}

	cmd := "/Users/wu/bin/of-create";
	args := []string{"--" + name, "::" + parent, "#" + deferDate, "#" + dueDate}

	fmt.Println(args)

	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return;
	}
	fmt.Println(string(output))

	strings := strings.Split(string(output), " ")

	if strings[1] == "id" {
		http.Redirect(w, r, "/id/"+strings[2], http.StatusFound)
	} else if strings[2] == "id" {
		http.Redirect(w, r, "/id/"+strings[3], http.StatusFound)
	}

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
    rows, err := db.Query("SELECT persistentIdentifier, name, parent, dateCompleted, dateDue FROM task where persistentIdentifier = ?", id)
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

	return t, nil
}

func main() {
    http.HandleFunc("/id/",     idHandler)
    http.HandleFunc("/create", createHandler)

    http.ListenAndServe(":8080", nil)
}
