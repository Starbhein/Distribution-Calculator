package stats

type EmpiricalStats struct {
	Count    int64
	Avg      float64
	Variance float64
}

func AnalyzeBuffer(buffer []float64) EmpiricalStats {
	var count int64
	var avg float64
	var m2 float64
	for _, x := range buffer {
		count++
		delta := x - avg
		avg += delta / float64(count)
		delta2 := x - avg
		m2 += delta * delta2
	}
	variance := 0.0
	if count > 1 {
		variance = m2 / float64(count-1) // removing imprecision of range [] ,instead using ranges-> ()
	}
	return EmpiricalStats{
		Count:    count,
		Avg:      avg,
		Variance: variance,
	}
}
