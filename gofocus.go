package main

import (
	"fmt"
	"net/http"
    "database/sql"
    "encoding/json"
	"os/exec"
	"strings"
	"time"

    _ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Id			string
	Name		string
	Parent		string
	Start		time.Time
	Due			time.Time
	Completed	bool
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
	dbfile := "/Users/wu/Library/Containers/com.omnigroup.OmniFocus2.MacAppStore/Data/Library/Caches/com.omnigroup.OmniFocus2.MacAppStore/OmniFocusDatabase2"
	db, _ := sql.Open("sqlite3", "file:" + dbfile + "?mode=ro")

	rows, err := db.Query("SELECT persistentIdentifier, name, parent, dateToStart + 978307200, dateDue + 978307200, CAST(dateCompleted AS INT) + 978307200 FROM task where persistentIdentifier = ?", id)
    if err != nil {
		return nil, err
    }

    var persistentIdentifier string
    var name string
    var parent string
    var dateToStart int64
    var dateDue int64
    var dateCompleted int64

    rows.Next();
	rows.Scan(&persistentIdentifier, &name, &parent, &dateToStart, &dateDue, &dateCompleted)

	// dateToStart = dateToStart + 978307200;
	dateToStartReal := time.Unix(dateToStart, 0)
	fmt.Println("%s", dateToStartReal);

	// dateDue = dateDue + 978307200;
	dateDueReal := time.Unix(dateDue, 0)
	fmt.Println("%s", dateDueReal);

	// dateCompletedReal := time.Unix(dateCompleted, 0)
	fmt.Println("completed: ", dateCompleted);

	var completeFlag bool
	if dateCompleted == 0 {
		completeFlag = false
	} else {
		completeFlag = true
	}

	t := &Task{Id: persistentIdentifier, Name: name, Parent: parent, Start: dateToStartReal, Due: dateDueReal, Completed: completeFlag}


	rows.Close()

	db.Close()

	return t, nil
}

func main() {
    http.HandleFunc("/id/",     idHandler)
    http.HandleFunc("/create", createHandler)

    http.ListenAndServe(":8080", nil)
}
