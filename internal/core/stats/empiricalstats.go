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

type WelfordAccumulator struct {
	Count float64
	Avg   float64
	M2    float64
}

func (w *WelfordAccumulator) Update(value float64) {
	w.Count++
	delta := value - w.Avg
	w.Avg += delta / w.Count
	delta2 := value - w.Avg
	w.M2 += delta * delta2
}

func MergeWelford(a, b WelfordAccumulator) WelfordAccumulator {
	if a.Count == 0 {
		return b
	}
	if b.Count == 0 {
		return a
	}
	combined := WelfordAccumulator{}
	combined.Count = a.Count + b.Count

	delta := b.Avg - a.Avg
	combined.Avg = a.Avg + delta*(b.Count/combined.Count)
	combined.M2 = a.M2 + b.M2 + (delta*delta)*(a.Count*b.Count/combined.Count)

	return combined
}
