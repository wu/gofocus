package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type task struct {
	ID        string
	Name      string
	Parent    string
	Start     time.Time
	Due       time.Time
	Completed bool
}

var validID = regexp.MustCompile("^/(id|done|undone)/([a-zA-Z0-9_.]+)$")
var validQuery = regexp.MustCompile("^/query/([a-zA-Z0-9_%.]+)$")
var validName = regexp.MustCompile("^([a-zA-Z0-9_. -]+)$")
var validDate = regexp.MustCompile("^([a-zA-Z0-9_. -]+)$")

var queryBase = "SELECT persistentIdentifier, name, parent, dateToStart + 978307200, dateDue + 978307200, CAST(dateCompleted AS INT) + 978307200 FROM task "

func parseRequestID(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validID.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid ID")
	}
	return m[2], nil
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	m1 := validName.FindStringSubmatch(r.FormValue("name"))
	if m1 == nil {
		http.Error(w, "Invalid Name", http.StatusInternalServerError)
		return
	}
	name := m1[1]

	m2 := validName.FindStringSubmatch(r.FormValue("parent"))
	if m2 == nil {
		http.Error(w, "Invalid Parent", http.StatusInternalServerError)
		return
	}
	parent := m2[1]
	if parent == "" {
		http.Error(w, "ERROR: parent not specified", http.StatusInternalServerError)
		return
	}

	deferDate := ""
	defered := r.FormValue("defer")
	if defered != "" {
		m3 := validDate.FindStringSubmatch(defered)
		if m3 == nil {
			http.Error(w, "Invalid Defer Date", http.StatusInternalServerError)
			return
		}
		deferDate = m3[1]

	}

	dueDate := ""
	due := r.FormValue("due")
	if due != "" {
		m4 := validDate.FindStringSubmatch(due)
		if m4 == nil {
			http.Error(w, "Invalid Due Date", http.StatusInternalServerError)
			return
		}
		dueDate = m4[1]
	}

	tasktxt := "--" + name + " ::" + parent + " #" + deferDate + " #" + dueDate

	cmd := "/usr/bin/osascript"
	args := []string{"-e", `tell application "OmniFocus"`,
		"-e", `parse tasks into default document with transport text "` + tasktxt + `"`,
		"-e", "end tell"}

	fmt.Println(cmd, args)

	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	id, err := parseRequestID(w, r)
	if err != nil {
		return
	}

	t, err := loadID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	if err := enc.Encode(&t); err != nil {
		fmt.Fprintf(w, "%s", err)
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	m := validQuery.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.Error(w, "Invalid Query", http.StatusInternalServerError)
		return
	}
	query := m[1]

	db, dberr := openDB()
	if dberr != nil {
		http.Error(w, dberr.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(queryBase+"where name like ? LIMIT 50", query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var persistentIdentifier string
	var name string
	var parent string
	var dateToStart int64
	var dateDue int64
	var dateCompleted int64

	tasks := make([]*task, 0)

	idx := -1
	for rows.Next() {
		rows.Scan(&persistentIdentifier, &name, &parent, &dateToStart, &dateDue, &dateCompleted)

		dateToStartReal := time.Unix(dateToStart, 0)
		dateDueReal := time.Unix(dateDue, 0)

		var completeFlag bool
		if dateCompleted == 0 {
			completeFlag = false
		} else {
			completeFlag = true
		}

		t := &task{ID: persistentIdentifier, Name: name, Parent: parent, Start: dateToStartReal, Due: dateDueReal, Completed: completeFlag}

		idx++
		//tasks[idx] = t
		tasks = append(tasks, t)
	}

	rows.Close()
	db.Close()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	if err := enc.Encode(&tasks); err != nil {
		fmt.Fprintf(w, "%s", err)
	}
}

func doneHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseRequestID(w, r)
	if err != nil {
		return
	}

	cmd := "/usr/bin/osascript"
	args := []string{"-e", `tell application "OmniFocus"`,
		"-e", "tell default document",
		"-e", `set selectedTask to first flattened task whose (id = "` + id + `")`,
		"-e", "set completed of selectedTask to true",
		"-e", "end tell",
		"-e", "end tell"}

	fmt.Println(cmd, args)

	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(output))

	http.Redirect(w, r, "/id/"+id, http.StatusFound)
}

func undoneHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseRequestID(w, r)
	if err != nil {
		return
	}

	cmd := "/usr/bin/osascript"
	args := []string{"-e", `tell application "OmniFocus"`,
		"-e", "tell default document",
		"-e", `set selectedTask to first flattened task whose (id = "` + id + `")`,
		"-e", "set completed of selectedTask to false",
		"-e", "end tell",
		"-e", "end tell"}

	fmt.Println(cmd, args)

	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(output))

	http.Redirect(w, r, "/id/"+id, http.StatusFound)
}

func openDB() (*sql.DB, error) {
	var dbfile string
	if os.Getenv("DBFILE") == "" {
		dbfile = os.Getenv("HOME") + "/Library/Containers/com.omnigroup.OmniFocus2.MacAppStore/Data/Library/Caches/com.omnigroup.OmniFocus2.MacAppStore/OmniFocusDatabase2"
	} else {
		dbfile = os.Getenv("DBFILE")
	}
	db, _ := sql.Open("sqlite3", "file:"+dbfile+"?mode=ro")
	return db, nil
}

func loadID(id string) (*task, error) {
	db, dberr := openDB()
	if dberr != nil {
		return nil, dberr
	}

	rows, err := db.Query(queryBase+"where persistentIdentifier = ? LIMIT 1", id)
	if err != nil {
		return nil, err
	}

	var persistentIdentifier string
	var name string
	var parent string
	var dateToStart int64
	var dateDue int64
	var dateCompleted int64

	rows.Next()
	rows.Scan(&persistentIdentifier, &name, &parent, &dateToStart, &dateDue, &dateCompleted)

	dateToStartReal := time.Unix(dateToStart, 0)
	dateDueReal := time.Unix(dateDue, 0)

	var completeFlag bool
	if dateCompleted == 0 {
		completeFlag = false
	} else {
		completeFlag = true
	}

	t := &task{ID: persistentIdentifier, Name: name, Parent: parent, Start: dateToStartReal, Due: dateDueReal, Completed: completeFlag}

	rows.Close()
	db.Close()

	return t, nil
}

func main() {
	http.HandleFunc("/id/", idHandler)
	http.HandleFunc("/done/", doneHandler)
	http.HandleFunc("/undone/", undoneHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/query/", queryHandler)

	http.ListenAndServe(":8080", nil)
}
