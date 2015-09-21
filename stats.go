package main

import (
	"math"
	"strconv"
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

type QueryStat struct {
	TblStats TableStats
	LastQry  string
}

func newQueryStat() *QueryStat {
	return &QueryStat{
		TblStats: make(TableStats),
	}
}

type QueryStats map[string]*QueryStat
