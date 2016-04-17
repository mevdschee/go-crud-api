package main

import (
	"encoding/json"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"bufio"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	method := req.Method
	u, _ := url.ParseRequestURI(req.RequestURI)
	request := strings.Split(strings.Trim(u.Path, "/"), "/")

	// load input from request body
	var input map[string]interface{}
	r := bufio.NewReader(req.Body);
	buf, _ := r.ReadBytes(0)
	json.Unmarshal(buf, &input)

	db, err := sql.Open("mysql", "php-crud-api:php-crud-api@/php-crud-api")
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
	values := make([]interface{}, 0, len(input))
	set := ""
	i:=0
	for column := range input {
		name := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(column, "")
		columns = append(columns, name)
		values = append(values, input[column])
		if i>0 {
			set += ", "
		}
		set+=fmt.Sprintf("`%s`=@%d",name,i)
		i++;
 	}

	// create SQL based on HTTP method
	sql:="";
	switch method {
	case "GET":
		if key > 0 {
			sql = fmt.Sprintf("select * from `%s` where `id`=@pk", table)
		} else {
			sql = fmt.Sprintf("select * from `%s`", table)
		}
		break
	case "PUT":
		sql = fmt.Sprintf("update `%s` set %s where `id`=@pk", table, set)
		break
	case "POST":
		sql = fmt.Sprintf("insert into `%s` set %s; select last_insert_id()", table, set)
		break
	case "DELETE":
		sql = fmt.Sprintf("delete `%s` where `id`=@pk", table)
		break
	}

	// close mysql connection
	defer db.Close()

	fmt.Fprint(w, sql)
}

func main() {
	http.HandleFunc("/", Handler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
