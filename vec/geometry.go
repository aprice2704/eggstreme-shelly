package vec

import (
	"fmt"
	"math"
)

const (
	mayAsWellBeZero = 1e-40
)

// Manipulations of simple geometrical things

// Line is an infinite line, all points P such that: P = d*In + On
type Line struct {
	PointOn Vec // a point on the line
	AlongN  Vec // a vector along the line, length 1
}

// Segment is a finite line on P = d*In + On, where MinD < d < MaxD
type Segment struct {
	Line
	MinD, MaxD float64
}

// Plane is an infinite flat surface: all points P such that: (P - In) dot Normal = 0
type Plane struct {
	PointOn Vec // a point in the plane
	Normal  Vec // a vector, length 1, normal to the plane
}

// Patch is a parallagram shaped area on a plane, defined by two edges in the plane.
type Patch struct {
	Plane        // The plane the patch lies on
	Corner Vec   // One corner of the area that is on the plane
	Sides  []Vec // vectors from Corner that lie in the plane
}

// NewPatch makes a new one
func NewPatch(newPoint, newNormal, side1, side2 Vec) Patch {
	np := Patch{}
	np.Plane = NewPlane(newPoint, newNormal)
	np.Corner = newPoint
	np.Sides = append(np.Sides, side1, side2)
	return np
}

// NewLine makes one given the a point on it and a vector along it
func NewLine(newOn, newAlong Vec) Line {
	nl := Line{}
	nl.PointOn = newOn
	nl.AlongN = newAlong.Normalized()
	return nl
}

// NewLineTwoPoints makes one given two points on it
func NewLineTwoPoints(p1, p2 Vec) Line {
	return NewLine(p1, p2.Subtract(p1))
}

// NewPlane makes a new one given a point and a normal
func NewPlane(newPoint, newNormal Vec) Plane {
	np := Plane{}
	np.PointOn = newPoint
	np.Normal = newNormal.Normalized()
	return np
}

// NewPlane3Points makes a new one given three points in the plane. Points should be provided clockwise along desired normal dir.
func NewPlane3Points(p0, p1, p2 Vec) Plane {
	p0p1 := p1.Subtract(p0)
	p0p2 := p2.Subtract(p0)
	return NewPlane(p0, p0p1.Cross(p0p2))
}

// NewSegment makes a new one
func NewSegment(l Line, minD, maxD float64) Segment {
	return Segment{Line: l, MinD: minD, MaxD: maxD}
}

// NewSegment2Ends returns a segment defined by points at each end
func NewSegment2Ends(p0, p1 Vec) Segment {
	return Segment{Line: NewLineTwoPoints(p0, p1), MinD: 0, MaxD: p1.Subtract(p0).Length()}
}

// Start returns the point at the start of the segment (=MinD)
func (seg Segment) Start() Vec {
	return seg.Line.PointOn.Add(seg.Line.AlongN.Scale(seg.MinD))
}

// End returns the point at the end of the segment (=MaxD)
func (seg Segment) End() Vec {
	return seg.Line.PointOn.Add(seg.Line.AlongN.Scale(seg.MaxD))
}

func (l Line) String() string {
	return fmt.Sprintf("Line, passes through %s direction %s", l.PointOn, l.AlongN)
}

func (seg Segment) String() string {
	return fmt.Sprintf("Segment, passes through %s direction %s, minD %4.1f and maxD %4.1f", seg.PointOn, seg.AlongN, seg.MinD, seg.MaxD)
}

func (p Plane) String() string {
	return fmt.Sprintf("Plane, contains %s normal %s", p.PointOn, p.Normal)
}

func (pa Patch) String() string {
	return fmt.Sprintf("Patch, contains %s normal %s\nSides: %s and %s", pa.PointOn, pa.Normal, pa.Sides[0], pa.Sides[1])
}

// IntersectLine determines whether the given line intersects this plane, and if so, where. hits = false -> line is parallel to plane.
func (p Plane) IntersectLine(l Line) (where Vec, hits bool) {
	ldotn := l.AlongN.Dot(p.Normal)
	if math.Abs(ldotn) < mayAsWellBeZero { // effectively zero
		return where, false
	}
	d := (p.PointOn.Subtract(l.PointOn).Dot(p.Normal) / ldotn)
	where = l.PointOn.Add(l.AlongN.Scale(d))
	return where, true
}

// IntersectSegment determines whether the given segment intersects this plane,
//   and if so, where. hits = false -> line is parallel to plane.
func (p Plane) IntersectSegment(s Segment) (where Vec, hits bool) {
	whu, anyHit := p.IntersectLine(s.Line)
	if !anyHit { // bail, they are parallel
		return where, false
	}
	howFar := whu.Subtract(s.PointOn).Dot(s.AlongN)
	if howFar <= s.MinD || howFar >= s.MaxD { // on the line, but beyond the segment
		return where, false
	}
	return whu, true
}

// sameSide is from https://blackpawn.com/texts/pointinpoly/default.html
func sameSide(p1, p2, a, b Vec) bool {
	bSubA := b.Subtract(a)
	cp1 := bSubA.Cross(p1.Subtract(a))
	cp2 := bSubA.Cross(p2.Subtract(a))
	if cp1.Dot(cp2) >= 0 {
		return true
	}
	return false
}

// inTriangle
func inTriangle(p, a, b, c Vec) bool {
	if sameSide(p, a, b, c) && sameSide(p, b, a, c) && sameSide(p, c, a, b) {
		return true
	}
	return false
}

// TriIntersectSegment determines whether the given segment
//   intersects the triangle defined by the sides of this patch,
//   and if so, where. hits = false -> line is parallel to plane.
func (pa Patch) TriIntersectSegment(s Segment) (where Vec, hits bool) {

	const r2d = 180 / math.Pi

	whu, anyHit := pa.Plane.IntersectSegment(s) // does the segment intersect my containing plane?
	if !anyHit {                                // nope, bail
		return where, false
	}

	if !inTriangle(whu, pa.Corner, pa.Corner.Add(pa.Sides[0]), pa.Corner.Add(pa.Sides[1])) {
		return where, false
	}

	return whu, true
}

// ParaIntersectSegment determines whether the given segment
//   intersects the parallelagram defined by the sides of this patch,
//   and if so, where. hits = false -> line is parallel to plane.
func (pa Patch) ParaIntersectSegment(s Segment) (where Vec, hits bool) {
	whu, anyHit := pa.Plane.IntersectSegment(s) // does the segment intersect my containing plane?
	if !anyHit {                                // nope, bail
		return where, false
	}
	diff := pa.Corner.Subtract(whu).Normalized()
	for _, side := range pa.Sides {
		d := diff.Dot(side)
		if d > side.Length() { // outside the parallelagram
			return where, false
		}
	}
	return whu, true
}

// s0 := pa.Sides[0]
// 	l0 := s0.Length()
// 	s0N := s0.Normalized()

// 	s1 := pa.Sides[1]
// 	l1 := s1.Length()
// 	s1N := s1.Normalized()

// 	D := whu.Subtract(pa.Corner)
// 	d := D.Length()

// 	d0 := D.Dot(s0N)
// 	d1 := D.Dot(s1N)

// 	//	fmt.Printf("D and projs: %s, %6.3f vs %6.3f OR %6.3f vs %6.3f\n", D, d0, l0, d1, l1)

// 	if (d0 > l0) || (d0 <= 0) || (d1 > l1) || (d1 <= 0) {
// 		fmt.Println("Early bail")
// 		return where, false
// 	}

// 	cosTheta := s0N.Dot(s1N)
// 	//	theta := math.Acos(cosTheta)
// 	//	fmt.Printf("theta: %5.2f,  alpha: %5.2f, beta: %5.2f\n", theta*r2d, math.Acos(d0/d)*r2d, math.Acos(d1/d)*r2d)

// 	if (d0 + d1) > d*(1+cosTheta) {
// 		fmt.Printf("D and projs: %s, d0: %6.3f d1: %6.3f  d0+d1: %6.3f vs d*(1+cosT) %6.3f\n", D, d0, d1, d0+d1, d*(1+cosTheta))
// 		return where, false
// 	}

// 	// if (math.Acos(d0/d) > theta) || (math.Acos(d1/d) > theta) {
// 	// 	return where, false
// 	// }

// 	//lTot := d0/l0 + d1/l1       // sum of the sides as fractions of sides
// 	//	limit := 1.0 + s0N.Dot(s1N) // i.e. 1+cos(theta) scaled to D

// 	// if lTot > limit {
// 	// 	return where, false
// 	// }

// 	fmt.Println("HIT!")

// Barycenter technique
// Compute vectors
// v0 = C - A
// v1 = B - A
// v2 = P - A

// // Compute dot products
// dot00 = dot(v0, v0)
// dot01 = dot(v0, v1)
// dot02 = dot(v0, v2)
// dot11 = dot(v1, v1)
// dot12 = dot(v1, v2)

// // Compute barycentric coordinates
// invDenom = 1 / (dot00 * dot11 - dot01 * dot01)
// u = (dot11 * dot02 - dot01 * dot12) * invDenom
// v = (dot00 * dot12 - dot01 * dot02) * invDenom

// // Check if point is in triangle
// return (u >= 0) && (v >= 0) && (u + v < 1)
//   For a triangle with sides ~a, ~b and origin ~O any point ~p in it obeys:
//      norm(~p-~O).norm(~a) + norm(~p-~O).norm(b~) < 1 + cos(theta)
//   where theta is the angle between ~a & ~b.
