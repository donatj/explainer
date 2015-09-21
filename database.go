package main

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"regexp"
	"strconv"
)

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
		out = regexp.MustCompile(`(?m)(^|\s|,|\()\-?\d+($|\s|,|\))`).ReplaceAllString(out, `$1[[int]]$2`)
		if out == lastOut {
			break
		}
		lastOut = out
	}

	out = regexp.MustCompile(`\((?:\s*\[\[([a-z]+)\]\]\s*,?\s*)+\)`).ReplaceAllString(out, `[[$1-list]]`)

	return out
}

func (qry *selectQuery) csha1() string {
	h := sha1.New()
	io.WriteString(h, qry.c14n())
	return fmt.Sprintf("%x", h.Sum(nil))
}

type explainEntry struct {
	Table string
	Key   string
	Rows  int
}

func (qry *selectQuery) explain(db *sql.DB) ([]explainEntry, error) {
	output := make([]explainEntry, 0)

	rows, err := db.Query("EXPLAIN " + string(*qry))
	if err != nil {
		return output, fmt.Errorf("Explain Error, %s", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for i, _ := range cols {
			vals[i] = new(sql.RawBytes)
		}
		err = rows.Scan(vals...)
		if err != nil {
			panic(err)
		}

		tbl := getColByName("table", cols, vals)
		rows := getColByName("rows", cols, vals)
		key := getColByName("key", cols, vals)

		rowInt, err := strconv.Atoi(*rows)
		if err != nil {
			rowInt = 0
		}

		output = append(output, explainEntry{
			Table: *tbl,
			Rows:  rowInt,
			Key:   *key,
		})
	}

	return output, nil
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
			idInt = int64(0)
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
}
