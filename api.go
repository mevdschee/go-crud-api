package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	method := req.Method
	u, _ := url.ParseRequestURI(req.RequestURI)
	request := strings.Split(strings.Trim(u.Path, "/"), "/")

	fmt.Fprintf(w, "<p>%s</p>", method)
	fmt.Fprintf(w, "<p>%s</p>", request)

	db, err := sql.Open("mysql", "php-crud-api:php-crud-api@/php-crud-api")
	if err != nil {
		panic(err.Error())
	}

	// retrieve the table and key from the path
	table := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(request[0], "")
	key := 0
	if len(request) > 1 {
		strconv.ParseInt(request[1], 10, 64)
	}

	/*
	   // escape the columns from the input object
	   string[] columns = input!=null ? input.Keys.Select(i => Regex.Replace(i.ToString(), "[^a-z0-9_]+", "")).ToArray() : null;

	   // build the SET part of the SQL command
	   string set = input != null ? String.Join (", ", columns.Select (i => "[" + i + "]=@_" + i).ToArray ()) : "";
	*/
	// create SQL based on HTTP method
	//	string sql := null;
	switch method {
	case "GET":
		if key > 0 {
			sql := fmt.Sprintf("select * from `{0}` where `id`=@pk", table)
		} else {
			sql := fmt.Sprintf("select * from `{0}`", table)
		}
		break
	case "PUT":
		sql := fmt.Sprintf("update `{0}` set {1} where `id`=@pk", table, set)
		break
	case "POST":
		sql := fmt.Sprintf("insert into `{0}` set {1}; select last_insert_id()", table, set)
		break
	case "DELETE":
		sql := fmt.Sprintf("delete `{0}` where `id`=@pk", table)
		break
	}

	// close mysql connection
	defer db.Close()

	fmt.Fprint(w, "<p>Hello World</p>")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{table}/{key}", Handler)
	r.HandleFunc("/{table}", Handler)
	r.HandleFunc("/", Handler)
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
