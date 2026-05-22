package distributions

import (
	"fmt"
	"testing"
)

func TestNormal(t *testing.T) {
	t.Run("testing a normal distribution with miu=0, sigma=1", func(t *testing.T) {
		normal, err := NewNormal(0, 1)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(normal.PDF(0))
		fmt.Println(normal.CDF(0))
	})
}
