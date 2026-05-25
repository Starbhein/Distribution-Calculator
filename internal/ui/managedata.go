package ui

import (
	"errors"
	"strconv"
)

type errorMessage struct {
	error error
	index int
}
type PoissonTask struct {
	Lambda float64
}

func (pt PoissonTask) FillBuffer(buffer []float64) error {
}

func Parser(parameters []string) ([]float64, errorMessage) {
	var err error
	res := make([]float64, len(parameters))
	for i, v := range parameters {
		res[i], err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errorMessage{error: errors.New("ingrese un numero valido"), index: i}
		}
	}
	return res, errorMessage{}
}
