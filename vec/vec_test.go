package vec

import (
	"fmt"
	"math"
	"testing"
)

func TestSimVectors(t *testing.T) {

	a := NewSimVec(3, 4, 0)
	b := NewSimVec(0, 3, 4)

	if NotApprox(a.Length(), 5) {
		t.Errorf("SimVec length failed")
	}
	if NotApprox(b.LengthSq(), 25) {
		t.Errorf("SimVec length squared failed")
	}

	c := NewSimVec(1, 0, 0)
	d := NewSimVec(math.Sin(math.Pi/4), math.Cos(math.Pi/4), 0)
	if NotApprox(c.Dot(d), math.Sin(math.Pi/4)) {
		t.Errorf("SimVec dot failed")
	}

	e := NewSimVec(0, 1, 0)
	f := c.Cross(e)
	if NotApprox(f.X(), 0) || NotApprox(f.Y(), 0) || NotApprox(f.Z(), 1) {
		t.Errorf("SimVec cross failed")
	}

	g := a.Normalized()
	if NotApprox(g.Length(), 1) {
		t.Errorf("SimVec Normalized failed")
	}

	h := g.Scale(23.1)
	if NotApprox(h.Length(), 23.1) {
		t.Errorf("SimVec Scale failed")
	}
}

func TestCPUVectors(t *testing.T) {

	a := NewCPUVec(3, 4, 0)
	b := NewCPUVec(0, 3, 4)

	if NotApprox(a.Length(), 5) {
		t.Errorf("CPUVec length failed")
	}
	if NotApprox(b.LengthSq(), 25) {
		t.Errorf("CPUVec length squared failed")
	}

	c := NewCPUVec(1, 0, 0)
	d := NewCPUVec(math.Sin(math.Pi/4), math.Cos(math.Pi/4), 0)
	if NotApprox(c.Dot(&d), math.Sin(math.Pi/4)) {
		t.Errorf("CPUVec dot failed")
	}

	e := NewCPUVec(0, 1, 0)
	f := c.Cross(e)
	if NotApprox(f.X(), 0) || NotApprox(f.Y(), 0) || NotApprox(f.Z(), 1) {
		t.Errorf("CPUVec cross failed")
	}

	g := a.Normalized()
	if NotApprox(g.Length(), 1) {
		t.Errorf("CPUVec Normalized failed")
	}

	h := g.Scale(23.1)
	if NotApprox(h.Length(), 23.1) {
		t.Errorf("CPUVec Scale failed")
	}

}

func NotApprox(a, b float64) bool {
	if math.Abs(a-b) > 0.000000001 {
		fmt.Printf("Difference %f\n", math.Abs(a-b))
		return true
	}
	return false
}
