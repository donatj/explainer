package main

type MinMaxAvg struct {
	Min    int
	Max    int
	Last   int
	AvgSum int
	Count  int
}

func newMinMaxAvg() *MinMaxAvg {
	return &MinMaxAvg{
		Min: 2147483648,
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

type TableStats map[string]*MinMaxAvg

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
