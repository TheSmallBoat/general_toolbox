package metrics

import (
	"math"
	"sync"
)
import "fmt"

// stats.go
// Author:   Gary Boone
// Copyright (c) 2011-2013 Gary Boone <gary.boone@gmail.com>.

// There are three ways to use GoStats as your program accumulates values:
// 1. Incremental or streaming -- include the new values one at a time
// 2. Incremental, in chunks -- include the new values in chunks by passing an array of values
//    Obtain the descriptive stats at any time by calling Mean(), Variance(), etc.
// See stats_test.go for examples of each.

// Data structure to contain accumulating values and moments

type Stats struct {
	mu                                 sync.Mutex
	n, min, max, sum, mean, m2, m3, m4 float64
}

// Accessor Functions

func (d *Stats) Count() int {
	return int(d.n)
}

func (d *Stats) Size() int {
	return int(d.n)
}

func (d *Stats) Min() float64 {
	return d.min
}

func (d *Stats) Max() float64 {
	return d.max
}

func (d *Stats) Sum() float64 {
	return d.sum
}

func (d *Stats) Mean() float64 {
	return d.mean
}

// Reset
func (d *Stats) Reset() {
	defer d.mu.Unlock()
	d.mu.Lock()

	d.n = 0.0
	d.min = 0.0
	d.max = 0.0
	d.sum = 0.0
	d.mean = 0.0
	d.m2 = 0.0
	d.m3 = 0.0
	d.m4 = 0.0
}

// Incremental Functions

// Update the stats with the given value.
func (d *Stats) Update(x float64) {
	defer d.mu.Unlock()
	d.mu.Lock()

	if d.n == 0.0 || x < d.min {
		d.min = x
	}
	if d.n == 0.0 || x > d.max {
		d.max = x
	}
	d.sum += x
	nMinus1 := d.n
	d.n += 1.0
	delta := x - d.mean
	delta_n := delta / d.n
	delta_n2 := delta_n * delta_n
	term1 := delta * delta_n * nMinus1
	d.mean += delta_n
	d.m4 += term1*delta_n2*(d.n*d.n-3*d.n+3.0) + 6*delta_n2*d.m2 - 4*delta_n*d.m3
	d.m3 += term1*delta_n*(d.n-2.0) - 3*delta_n*d.m2
	d.m2 += term1
}

// Update the stats with the given array of values.
func (d *Stats) UpdateArray(data []float64) {
	for _, v := range data {
		d.Update(v)
	}
}

func (d *Stats) PopulationVariance() float64 {
	if d.n == 0 || d.n == 1 {
		return math.NaN()
	}
	return d.m2 / d.n
}

func (d *Stats) SampleVariance() float64 {
	if d.n == 0 || d.n == 1 {
		return math.NaN()
	}
	return d.m2 / (d.n - 1.0)
}

func (d *Stats) PopulationStandardDeviation() float64 {
	if d.n == 0 || d.n == 1 {
		return math.NaN()
	}
	return math.Sqrt(d.PopulationVariance())
}

func (d *Stats) SampleStandardDeviation() float64 {
	if d.n == 0 || d.n == 1 {
		return math.NaN()
	}
	return math.Sqrt(d.SampleVariance())
}

func (d *Stats) PopulationSkew() float64 {
	return math.Sqrt(d.n/(d.m2*d.m2*d.m2)) * d.m3
}

func (d *Stats) SampleSkew() float64 {
	if d.n == 2.0 {
		return math.NaN()
	}
	popSkew := d.PopulationSkew()
	return math.Sqrt(d.n*(d.n-1.0)) / (d.n - 2.0) * popSkew
}

// The kurtosis functions return _excess_ kurtosis, so that the kurtosis of a normal
// distribution = 0.0. Then kurtosis < 0.0 indicates platykurtic (flat) while
// kurtosis > 0.0 indicates leptokurtic (peaked) and near 0 indicates mesokurtic.Update
func (d *Stats) PopulationKurtosis() float64 {
	return (d.n*d.m4)/(d.m2*d.m2) - 3.0
}

func (d *Stats) SampleKurtosis() float64 {
	if d.n == 2.0 || d.n == 3.0 {
		return math.NaN()
	}
	populationKurtosis := d.PopulationKurtosis()
	return (d.n - 1.0) / ((d.n - 2.0) * (d.n - 3.0)) * ((d.n+1.0)*populationKurtosis + 6.0)
}

func (d *Stats) ReportStatsInfo(title string) string {
	report := fmt.Sprintf("%s ...... \n", title)
	report += fmt.Sprintf("Size[%d], Min[%.2f], Mean[%.2f], Max[%.2f] \n", d.Size(), d.Min(), d.Mean(), d.Max())
	report += fmt.Sprintf("PopulationVariance[%.2f], SampleVariance[%.2f], PopulationStandardDeviation[%.2f], SampleStandardDeviation[%.2f] \n", d.PopulationVariance(), d.SampleVariance(), d.PopulationStandardDeviation(), d.SampleStandardDeviation())
	report += fmt.Sprintf("PopulationSkew[%.2f], SampleSkew[%.2f], PopulationKurtosis[%.2f], SampleKurtosis[%.2f] \n", d.PopulationSkew(), d.SampleSkew(), d.PopulationKurtosis(), d.SampleKurtosis())
	return report
}
func (d *Stats) ReportStatsInfoWithJson(title string) string {
	pv, sv, psd, ssd, ps, ss, pk, sk := d.PopulationVariance(), d.SampleVariance(), d.PopulationStandardDeviation(), d.SampleStandardDeviation(), d.PopulationSkew(), d.SampleSkew(), d.PopulationKurtosis(), d.SampleKurtosis()

	report := fmt.Sprintf("\"%s\":{", title)
	report += fmt.Sprintf("\"Size\":%d,\"Min\":%.6f,\"Mean\":%.6f,\"Max\":%.6f", d.Size(), d.Min(), d.Mean(), d.Max())
	if pv != math.NaN() {
		report += fmt.Sprintf(",\"PopulationVariance\":%.6f", pv)
	}
	if sv != math.NaN() {
		report += fmt.Sprintf(",\"SampleVariance\":%.6f", sv)
	}
	if psd != math.NaN() {
		report += fmt.Sprintf(",\"PopulationStandardDeviation\":%.6f", psd)
	}
	if ssd != math.NaN() {
		report += fmt.Sprintf(",\"SampleStandardDeviation\":%.6f", ssd)
	}
	if ps != math.NaN() {
		report += fmt.Sprintf(",\"PopulationSkew\":%.6f", ps)
	}
	if ss != math.NaN() {
		report += fmt.Sprintf(",\"SampleSkew\":%.6f", ss)
	}
	if pk != math.NaN() {
		report += fmt.Sprintf(",\"PopulationKurtosis\":%.6f", pk)
	}
	if sk != math.NaN() {
		report += fmt.Sprintf(",\"SampleKurtosis\":%.6f", sk)
	}

	report += "}"
	return report
}
