package main

import (
	"database/sql"
	"flag"
	//	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	//	"strconv"
	"regexp"
	"strconv"
	"time"
)

var (
	dsn       string
	sleepTime *uint
)

type dsnStat struct {
}

func init() {
	sleepTime = flag.Uint("sleep", 500, "Time to sleep in ms")

	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("First argument must be a valid DSN in the format of [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]\n\texample: username:password@tcp(10.10.10.10:3306)/databaseName")
	}

	dsn = flag.Arg(0)
}

func getDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	log.Println("starting...")
	sleepDuration := time.Duration(*sleepTime) * time.Millisecond

	db, err := getDb(dsn)
	if err != nil {
		log.Fatal(dsn)
	}

	for {
		result := getActiveQueries(db)
		//		log.Printf("%#v", result)
		//		log.Print(len(result))

		for _, y := range result {
			log.Println(y.Query.c14n(), y.Id, y.Time)
		}

		time.Sleep(sleepDuration)
	}
}

func colPos(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func getColByName(name string, cols []string, vals []interface{}) *string {
	if cmdi := colPos(cols, name); cmdi != -1 {
		if bytes, ok := vals[cmdi].(*sql.RawBytes); ok {
			str := string(*bytes)
			return &str
		} else {
			panic("not raw bytes")
		}
	}

	return nil
}

type selectQuery string

type procEntry struct {
	Id    int64
	Time  int
	Query selectQuery
}

func (qry *selectQuery) c14n() string {
	out := string(*qry)

	out = regexp.MustCompile(`"(?:\\"|""|[^"])+"|'(?:\\'|''|[^'])+'`).ReplaceAllString(out, "[[string]]")

	// @todo negative numbers present interesting problems

	lastOut := out
	for { //solves a problem with sets like 10,20,30 when there are no lookaround options as in go
		out = regexp.MustCompile(`(?m)(^|\s|,|\()\d+\.\d+($|\s|,|\))`).ReplaceAllString(out, `$1[[float]]$2`)
		if out == lastOut {
			break
		}
		lastOut = out
	}

	lastOut = out
	for {
		out = regexp.MustCompile(`(?m)(^|\s|,|\()\d+($|\s|,|\))`).ReplaceAllString(out, `$1[[int]]$2`)
		if out == lastOut {
			break
		}
		lastOut = out
	}

	out = regexp.MustCompile(`\((?:\s*\[\[([a-z]+)\]\]\s*,?\s*)+\)`).ReplaceAllString(out, `[[$1-list]]`)

	return out
}

func getActiveQueries(db *sql.DB) []procEntry {
	rows, err := db.Query("SHOW FULL PROCESSLIST")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	output := make([]procEntry, 0)

	isSelect := regexp.MustCompile("(?i)^\\s*select\\s")

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for i, _ := range cols {
			vals[i] = new(sql.RawBytes)
		}
		err = rows.Scan(vals...)
		if err != nil {
			panic(err)
		}

		id := getColByName("Id", cols, vals)
		timez := getColByName("Time", cols, vals)
		cmd := getColByName("Command", cols, vals)
		info := getColByName("Info", cols, vals)

		if *cmd != "Query" || !isSelect.MatchString(*info) {
			continue
		}

		idInt, err := strconv.ParseInt(*id, 10, 64)
		if err != nil {
			panic(err)
		}

		timeInt, err := strconv.Atoi(*timez)
		if err != nil {
			panic(err)
		}

		output = append(output, procEntry{
			Id:    idInt,
			Time:  timeInt,
			Query: selectQuery(*info),
		})
	}

	return output
	//	return sessions
}
