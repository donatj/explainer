package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	dsn string

	port      = flag.Uint("port", 8080, "The network port")
	sleepTime = flag.Uint("sleep", 500, "Time to sleep in ms")
)

var (
	httpUser = flag.String("basic-http-user", "", "Username for basic auth protected pages")
	httpPass = flag.String("basic-http-pass", "", "Password for basic auth protected pages")
)

func init() {
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

var (
	data = make(QueryStats)
)

func main() {
	log.Println("starting...")

	db, err := getDb(dsn)
	if err != nil {
		log.Fatal(err)
	}

	go updateLoop(db)

	http.HandleFunc("/queryData/", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(data)
	}, *httpUser, *httpPass))

	err = http.ListenAndServe(":"+strconv.Itoa(int(*port)), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func updateLoop(db *sql.DB) {

	for {
		result := getActiveQueries(db)

		for _, y := range result {
			//			log.Println(y.Query.c14n(), y.Id, y.Time)
			exp, err := y.Query.explain(db)
			if err != nil {
				log.Println(err)
				continue
			}

			c := y.Query.csha1()

			if _, ok := data[c]; !ok {
				data[c] = newQueryStat()
			}

			data[c].LastSeenTimestamp = time.Now().Unix()
			data[c].LastQry = string(y.Query)
			data[c].C14nQry = y.Query.c14n()

			totalRows := 0
			for _, e := range exp {
				if _, ok := data[c].TblStats[e.Table]; !ok {
					data[c].TblStats[e.Table] = newTableMinMaxAvgHisto()
				}

				data[c].TblStats[e.Table].AddValue(e.Rows)
				data[c].TblStats[e.Table].AddIndex(e.Key)

				totalRows += e.Rows
			}

			data[c].PowerHisto.AddValue(totalRows)
			data[c].QueryPowerSamples.AddSample(totalRows, data[c].LastQry)
		}

		if len(result) > 0 {
			f, err := os.Create("explainer-log.json")
			if err != nil {
				panic(err)
			}

			enc := json.NewEncoder(f)

			err = enc.Encode(data)
			if err != nil {
				panic(err)
			}

			f.Close()
			fmt.Println(".", len(result), len(data))
		}

		time.Sleep(time.Duration(*sleepTime) * time.Millisecond)
	}
}
