package vec

import (
	"fmt"
	"testing"
)

func TestLine(t *testing.T) {

	origin := NewSimVec(0, 0, 0)

	pLine := NewSimVec(5, 5, 0)
	alongLine := NewSimVec(0, 0, 20)
	l := NewLine(pLine, alongLine)

	pPlane := NewSimVec(0, 0, 5)
	nPlane := NewSimVec(0, 0, 1)
	pl := NewPlane(pPlane, nPlane)

	ix, hits := pl.IntersectLine(l)

	fmt.Printf("Line: %s\n", l)
	fmt.Printf("Plane: %s\n", pl)
	fmt.Printf("Intersection: %t, %s\n", hits, ix)

	if !hits || NotApprox(ix.X(), 5) || NotApprox(ix.Y(), 5) || NotApprox(ix.Z(), 5) {
		t.Error("Line & plane intersection failed")
	}

	segment1 := NewSegment(l, 0, 4)
	_, hitsShort := pl.IntersectSegment(segment1)
	if hitsShort {
		t.Error("Short segment should not hit plane")
	}

	segmentLonger := NewSegment(l, 0, 6)
	ix, hitsPlane := pl.IntersectSegment(segmentLonger)
	if !hitsPlane || NotApprox(ix.X(), 5) || NotApprox(ix.Y(), 5) || NotApprox(ix.Z(), 5) {
		t.Error("Longer segment should reach plane")
	}

	// patchLarge := NewPatch(pPlane, nPlane, NewSimVec(20, 0, 0), NewSimVec(0, 20, 0))
	// ix, hitsPatchLarge := patchLarge.TriIntersectSegment(segmentLonger)
	// if !hitsPatchLarge || NotApprox(ix.X(), 5) || NotApprox(ix.Y(), 5) || NotApprox(ix.Z(), 5) {
	// 	t.Error("Segment should reach larger patch")
	// }

	// patchSmall := NewPatch(pPlane, nPlane, NewSimVec(4, 0, 0), NewSimVec(0, 4, 0))
	// ix, hitsPatchSmall := patchSmall.TriIntersectSegment(segmentLonger)
	// if hitsPatchSmall {
	// 	t.Error("Segment should not hit small patch")
	// }

	// 10 up
	myplane := NewPlane(NewSimVec(-10, -10, 10), NewSimVec(0, 0, -1)) // , NewSimVec(20, 0, 0), NewSimVec(20, 0, 0))
	fmt.Printf("\n\nPLANE LINE TESTS\n%s\n", myplane)
	var ls []Line

	ls = append(ls, NewLine(origin, NewSimVec(1, 1, 5)))
	ls = append(ls, NewLine(origin, NewSimVec(1, -1, 5)))
	ls = append(ls, NewLine(origin, NewSimVec(-1, 1, 5)))
	ls = append(ls, NewLine(origin, NewSimVec(-1, -1, 5)))

	ls = append(ls, NewLine(origin, NewSimVec(1, 1, 0.9)))
	ls = append(ls, NewLine(origin, NewSimVec(1, -1, 3)))
	ls = append(ls, NewLine(origin, NewSimVec(-1, 1, 0.9)))
	ls = append(ls, NewLine(origin, NewSimVec(-1, -1, 3)))

	for i := range ls {
		whu, hitl := myplane.IntersectLine(ls[i])
		if hitl {
			fmt.Printf("Line %s intersects plane at %s\n", ls[i], whu)
		} else {
			fmt.Println("No hit!")
		}
	}

	// 10 up
	mypatch := NewPatch(NewSimVec(-10, -10, 10), NewSimVec(0, 0, -1), NewSimVec(20, 0, 0), NewSimVec(0, 20, 0))
	fmt.Printf("\n\nPATCH LINE TESTS\n%s\n", mypatch)
	for i := range ls {
		sg := NewSegment(ls[i], 0, 50)
		whu, hitl := mypatch.TriIntersectSegment(sg)
		if hitl {
			fmt.Printf("Hit %s intersects patch at %s\n", sg, whu)
		} else {
			fmt.Printf("No hit! Segment: %s\n", sg)
		}
	}

	// if NotApprox(a.Length(), 5) {
	// 	t.Errorf("SimVec length failed")
	// }
	// if NotApprox(b.LengthSq(), 25) {
	// 	t.Errorf("SimVec length squared failed")
	// }

}
