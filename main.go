package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

var (
	dsn       string
	sleepTime *uint
)

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

	data := make(QueryStats)

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

			data[c].LastQry = string(y.Query)

			for _, e := range exp {
				if _, ok := data[c].TblStats[e.Table]; !ok {
					data[c].TblStats[e.Table] = newMinMaxAvg()
				}

				data[c].TblStats[e.Table].AddValue(e.Rows)
			}
		}

		if len(result) > 0 {
			f, err := os.Create("explainer-log.json")
			if err != nil {
				panic(err)
			}

			enc := json.NewEncoder(f)

			enc.Encode(data)

			f.Close()

		}

		time.Sleep(sleepDuration)
	}
}
