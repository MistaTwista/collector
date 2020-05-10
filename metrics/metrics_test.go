package metrics_test

import (
	"testing"

	"collector/metrics"
)

type counter struct {
	Val float64
}

func (t *counter) Add(n float64) {
	t.Val = n
}

func TestCounter(t *testing.T) {
	cases := []struct {
		Seq []float64
		Res float64
	}{
		{[]float64{10,5,10,15,10,5,5}, 40},
		{[]float64{3.2,3.3,3.5,0.2,0.2,0.2,3}, 6.5},
		{[]float64{0,1,4,4,2,6,4}, 14},
	}

	for _, c := range cases {
		ta := &counter{}
		cm := metrics.NewCounter(ta)
		for _, n := range c.Seq {
			cm.Set(n)
		}
		if cm.Value() != c.Res {
			t.Errorf("got %f, want %f", cm.Value(), c.Res)
		}
	}
}
