package main

import (
	"math"
	"strconv"
	"time"
)

type MinMaxAvg struct {
	Min    int
	Max    int
	Last   int
	AvgSum int
	Count  int
}

func newMinMaxAvg() *MinMaxAvg {
	return &MinMaxAvg{
		Min: math.MaxInt32,
	}
}

func (d *MinMaxAvg) AddValue(val int) {
	if val > d.Max {
		d.Max = val
	}

	if val < d.Min {
		d.Min = val
	}

	d.Last = val
	d.AvgSum += val
	d.Count += 1
}

type Histo map[string]int

func (d *Histo) AddValue(val int) {
	(*d)[strconv.Itoa(val)] += 1
}

type Log10Histo map[string]int

func (d *Log10Histo) AddValue(val int) {
	index := strconv.Itoa(int(math.Log10(float64(val))))
	(*d)[index] += 1
}

type MinMaxAvgHisto struct {
	Histo
	MinMaxAvg
}

func newMinMaxAvgHisto() *MinMaxAvgHisto {
	return &MinMaxAvgHisto{
		Histo:     make(Histo),
		MinMaxAvg: *newMinMaxAvg(),
	}
}

func (d *MinMaxAvgHisto) AddValue(val int) {
	d.Histo.AddValue(val)
	d.MinMaxAvg.AddValue(val)
}

type TableMinMaxAvgHisto struct {
	IndexHisto Histo
	MinMaxAvgHisto
}

func newTableMinMaxAvgHisto() *TableMinMaxAvgHisto {
	return &TableMinMaxAvgHisto{
		IndexHisto:     make(Histo),
		MinMaxAvgHisto: *newMinMaxAvgHisto(),
	}
}

func (d *TableMinMaxAvgHisto) AddIndex(val string) {
	d.IndexHisto[val] += 1
}

type TableStats map[string]*TableMinMaxAvgHisto

type QueryLog10Log map[string]string

func (d *QueryLog10Log) AddSample(val int, query string) {
	index := strconv.Itoa(int(math.Log10(float64(val))))
	(*d)[index] = query
}

type QueryStat struct {
	TblStats          TableStats
	LastSeenTimestamp int64
	LastQry           string
	C14nQry           string

	PowerSumHisto        Log10Histo
	QuerySumPowerSamples QueryLog10Log

	PowerProductHisto        Log10Histo
	QueryProductPowerSamples QueryLog10Log
}

func newQueryStat() *QueryStat {
	return &QueryStat{
		TblStats: make(TableStats),

		PowerSumHisto:        make(Log10Histo),
		QuerySumPowerSamples: make(QueryLog10Log),

		PowerProductHisto:        make(Log10Histo),
		QueryProductPowerSamples: make(QueryLog10Log),
	}
}

type QueryStats map[string]*QueryStat

func (qs *QueryStats) AddQueryStat(qry selectQuery, exp []explainEntry) {
	q := (*qs)
	c := qry.csha1()

	if _, ok := q[c]; !ok {
		q[c] = newQueryStat()
	}

	q[c].LastSeenTimestamp = time.Now().Unix()
	q[c].LastQry = string(qry)
	q[c].C14nQry = qry.c14n()

	rowSum := 0
	rowProduct := 1
	for _, e := range exp {
		if _, ok := data[c].TblStats[e.Table]; !ok {
			q[c].TblStats[e.Table] = newTableMinMaxAvgHisto()
		}

		q[c].TblStats[e.Table].AddValue(e.Rows)
		q[c].TblStats[e.Table].AddIndex(e.Key)

		rowSum += e.Rows
		rowProduct *= e.Rows
	}

	q[c].PowerSumHisto.AddValue(rowSum)
	q[c].QuerySumPowerSamples.AddSample(rowSum, q[c].LastQry)

	q[c].PowerProductHisto.AddValue(rowProduct)
	q[c].QueryProductPowerSamples.AddSample(rowProduct, q[c].LastQry)
}
