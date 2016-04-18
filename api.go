package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func RequestHandler(w http.ResponseWriter, req *http.Request) {
	msg := ""
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	method := req.Method
	u, _ := url.ParseRequestURI(req.RequestURI)
	request := strings.Split(strings.Trim(u.Path, "/"), "/")

	// load input from request body
	var input map[string]interface{}
	r := bufio.NewReader(req.Body)
	buf, _ := r.ReadBytes(0)
	json.Unmarshal(buf, &input)

	db, err := sql.Open("mysql", "php-crud-api:php-crud-api@unix(/var/run/mysqld/mysqld.sock)/php-crud-api")
	if err != nil {
		panic(err.Error())
	}

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
	for column := range input {
		name := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(column, "")
		columns[i] = name
		values[i] = input[column]
		if i > 0 {
			set += ", "
		}
		set += fmt.Sprintf("`%s`=@%d", name, i)
		i++
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
		query = fmt.Sprintf("insert into `%s` set %s; select last_insert_id()", table, set)
		break
	case "DELETE":
		query = fmt.Sprintf("delete `%s` where `id`=?", table)
		break
	}

	if key > 0 {
		values = append(values, key)
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
		values := make([]interface{}, len(cols))
		for i, _ := range values {
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
	} else {
		result, err := db.Exec(query, values...)
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
		}
		msg += string(b)
	}

	// close mysql connection
	defer db.Close()

	fmt.Fprint(w, msg)
}

func main() {
	http.HandleFunc("/", RequestHandler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
