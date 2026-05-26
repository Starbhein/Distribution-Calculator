package export

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ExportCSV writes simulated data to a CSV file.
func ExportCSV(data []float64, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"index", "value"}); err != nil {
		return fmt.Errorf("error escribiendo headers: %w", err)
	}

	for i, v := range data {
		if err := w.Write([]string{fmt.Sprintf("%d", i), fmt.Sprintf("%f", v)}); err != nil {
			return fmt.Errorf("error escribiendo fila %d: %w", i, err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("error finalizando CSV: %w", err)
	}
	return nil
}
