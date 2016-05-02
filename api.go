package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	connectionString = "php-crud-api:php-crud-api@unix(/var/run/mysqld/mysqld.sock)/php-crud-api"
	maxConnections   = 256
)

var (
	db *sql.DB
)

var (
	listenAddr = flag.String("listenAddr", ":8000", "Address to listen to")
	child      = flag.Bool("child", false, "is child proc")
)

func requestHandler(w http.ResponseWriter, req *http.Request) {
	var msg []byte
	w.Header().Add("Content-Type", "application/json")

	method := req.Method
	u, _ := url.ParseRequestURI(req.RequestURI)
	request := strings.Split(strings.Trim(u.Path, "/"), "/")

	// load input from request body
	var input map[string]interface{}
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&input)

	// retrieve the table and key from the path
	table := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(request[0], "")
	key := 0
	if len(request) > 1 {
		key, _ = strconv.Atoi(request[1])
	}

	// escape the columns from the input object
	var args []interface{}
	if key > 0 {
		args = make([]interface{}, 0, len(input)+1)
	} else {
		args = make([]interface{}, 0, len(input))
	}
	columns := make([]string, 0, len(input))
	for column, arg := range input {
		name := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(column, "")
		args = append(args, arg)
		columns = append(columns, fmt.Sprintf("`%s`=?", name))
	}
	set := strings.Join(columns, ", ")

	if key > 0 {
		args = append(args, key)
	}

	// create SQL based on HTTP method
	query := ""
	switch method {
	case "GET":
		if key > 0 {
			query = fmt.Sprintf("select * from `%s` where `id`=?", table)
		} else {
			query = fmt.Sprintf("select * from `%s`", table)
		}
		break
	case "PUT":
		query = fmt.Sprintf("update `%s` set %s where `id`=?", table, set)
		break
	case "POST":
		query = fmt.Sprintf("insert into `%s` set %s", table, set)
		break
	case "DELETE":
		query = fmt.Sprintf("delete from `%s` where `id`=?", table)
		break
	}

	if method == "GET" {
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Fatal(err)
		}

		cols, err := rows.Columns()
		if err != nil {
			log.Fatal(err)
		}
		values := make([]interface{}, len(cols))
		for i := range values {
			var value *string
			values[i] = &value
		}

		data := make(map[string]interface{})
		var records []interface{}
		for rows.Next() {
			err := rows.Scan(values...)
			if err != nil {
				log.Fatal(err)
			}
			records = append(records, values)
		}
		if key == 0 {
			data["columns"] = cols
			data["records"] = records
		} else {
			if len(records) > 0 {
				for i, col := range cols {
					data[col] = records[0].([]interface{})[i]
				}
			}
		}
		msg, _ = json.Marshal(data)
	} else if method == "POST" {
		result, err := db.Exec(query, args...)
		if err != nil {
			log.Fatal(err)
		} else {
			lastInsertID, _ := result.LastInsertId()
			msg, _ = json.Marshal(lastInsertID)
		}
	} else {
		result, err := db.Exec(query, args...)
		if err != nil {
			log.Fatal(err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			msg, _ = json.Marshal(rowsAffected)
		}
	}

	w.Write(msg)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		panic(err.Error())
	}

	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)

	// close mysql connection
	defer db.Close()

	http.HandleFunc("/", requestHandler)
	err = http.ListenAndServe(*listenAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
