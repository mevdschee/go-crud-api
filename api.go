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
	msg := ""
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
	columns := make([]string, 0, len(input))
	var values []interface{}
	if key > 0 {
		values = make([]interface{}, 0, len(input)+1)
	} else {
		values = make([]interface{}, 0, len(input))
	}
	set := ""
	i := 0
	for column, value := range input {
		name := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(column, "")
		columns = append(columns, column)
		values = append(values, value)
		if i > 0 {
			set += ", "
		}
		set += fmt.Sprintf("`%s`=?", name)
		i++
	}

	if key > 0 {
		values = append(values, key)
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
		query = fmt.Sprintf("delete `%s` where `id`=?", table)
		break
	}

	if method == "GET" {
		rows, err := db.Query(query, values...)
		if err != nil {
			log.Fatal(err)
		}

		cols, err := rows.Columns()
		if err != nil {
			log.Fatal(err)
		}
		values = make([]interface{}, len(cols))
		for i := range values {
			var value *string
			values[i] = &value
		}

		if key == 0 {
			msg += "["
		}
		first := true
		for rows.Next() {
			if first {
				first = false
			} else {
				msg += ","
			}
			err := rows.Scan(values...)
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(values)
			if err != nil {
				log.Fatal(err)
			}
			msg += string(b)
		}
		if key == 0 {
			msg += "]"
		}
	} else if method == "POST" {
		result, err := db.Exec(query, values...)
		if err != nil {
			log.Fatal(err)
		} else {
			lastInsertID, _ := result.LastInsertId()
			b, _ := json.Marshal(lastInsertID)
			msg += string(b)
		}
	} else {
		result, err := db.Exec(query, values...)
		if err != nil {
			log.Fatal(err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			b, _ := json.Marshal(rowsAffected)
			msg += string(b)
		}
	}

	fmt.Fprint(w, msg)
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
