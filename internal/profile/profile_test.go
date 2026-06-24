package profile

import "testing"

func TestCompactWindow(t *testing.T) {
	cases := []struct {
		ctx  int
		pct  float64
		want int
	}{
		{131072, 80, 104858},
		{262144, 80, 209715},
		{131072, 100, 131072},
		{0, 80, 0},
		{131072, 0, 0},
		{1000, -5, 0},
	}
	for _, c := range cases {
		if got := CompactWindow(c.ctx, c.pct); got != c.want {
			t.Errorf("CompactWindow(%d,%v)=%d want %d", c.ctx, c.pct, got, c.want)
		}
	}
}
