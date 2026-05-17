package sim

type EmpiricalSimulator interface {
	FillBuffer(buffer []float64) error
}
