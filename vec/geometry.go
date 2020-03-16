// ██╗   ██╗███████╗ ██████╗     ██╗ ██████╗ ███████╗ ██████╗ ███╗   ███╗██╗
// ██║   ██║██╔════╝██╔════╝    ██╔╝██╔════╝ ██╔════╝██╔═══██╗████╗ ████║╚██╗
// ██║   ██║█████╗  ██║         ██║ ██║  ███╗█████╗  ██║   ██║██╔████╔██║ ██║
// ╚██╗ ██╔╝██╔══╝  ██║         ██║ ██║   ██║██╔══╝  ██║   ██║██║╚██╔╝██║ ██║
//  ╚████╔╝ ███████╗╚██████╗    ╚██╗╚██████╔╝███████╗╚██████╔╝██║ ╚═╝ ██║██╔╝
//   ╚═══╝  ╚══════╝ ╚═════╝     ╚═╝ ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝

package vec

import (
	"fmt"
	"math"
)

// Very short lengths
const (
	PlanckLength    = 1e-12 // 1fM
	PlanckFactor    = 1 + PlanckLength
	mayAsWellBeZero = 1e-12
)

// Manipulations of simple geometrical things

// ██╗     ██╗███╗   ██╗███████╗
// ██║     ██║████╗  ██║██╔════╝
// ██║     ██║██╔██╗ ██║█████╗
// ██║     ██║██║╚██╗██║██╔══╝
// ███████╗██║██║ ╚████║███████╗
// ╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝

// Line is an infinite line, all points P such that: P = d*In + On
type Line struct {
	PointOn Vec // a point on the line
	AlongN  Vec // a vector along the line, length 1
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

func (l Line) String() string {
	return fmt.Sprintf("Line, passes through %s direction %s", l.PointOn, l.AlongN)
}

// ███████╗███████╗ ██████╗ ███╗   ███╗███████╗███╗   ██╗████████╗
// ██╔════╝██╔════╝██╔════╝ ████╗ ████║██╔════╝████╗  ██║╚══██╔══╝
// ███████╗█████╗  ██║  ███╗██╔████╔██║█████╗  ██╔██╗ ██║   ██║
// ╚════██║██╔══╝  ██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║
// ███████║███████╗╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║
// ╚══════╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝

// Segment is a finite line on P = d*In + On, where MinD < d < MaxD
type Segment struct {
	Line
	MinD, MaxD float64
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

func (seg Segment) String() string {
	return fmt.Sprintf("Segment, passes through %s direction %s, minD %4.1f and maxD %4.1f", seg.PointOn, seg.AlongN, seg.MinD, seg.MaxD)
}

// ██████╗ ██╗      █████╗ ███╗   ██╗███████╗
// ██╔══██╗██║     ██╔══██╗████╗  ██║██╔════╝
// ██████╔╝██║     ███████║██╔██╗ ██║█████╗
// ██╔═══╝ ██║     ██╔══██║██║╚██╗██║██╔══╝
// ██║     ███████╗██║  ██║██║ ╚████║███████╗
// ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝

// Plane is an infinite flat surface: all points P such that: (P - In) dot Normal = 0
type Plane struct {
	PointOn Vec // a point in the plane
	Normal  Vec // a vector, length 1, normal to the plane
}

// Planes where X=0, Y=0 and Z=0
var (
	XPlane = NewPlane(Origin, X)
	YPlane = NewPlane(Origin, Y)
	ZPlane = NewPlane(Origin, Z)
)

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

func (p Plane) String() string {
	return fmt.Sprintf("Plane, contains %s normal %s", p.PointOn, p.Normal)
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

// RotateZ rotates the plane by a radians about Z through its contained point
func (p *Plane) RotateZ(a Radians) *Plane {
	p.Normal = p.Normal.RotateZ(a)
	return p
}

// Translate moves a plane
func (p *Plane) Translate(by Vec) *Plane {
	p.PointOn = p.PointOn.Add(by)
	return p
}

// NormalSide returns true iff the given point is on
//   the side of the plane to which the normal points
func (p Plane) NormalSide(poi Vec) bool {
	r := poi.Subtract(p.PointOn).Dot(p.Normal)
	if r < 0 {
		return false
	}
	return true
}

// ██████╗  █████╗ ████████╗ ██████╗██╗  ██╗
// ██╔══██╗██╔══██╗╚══██╔══╝██╔════╝██║  ██║
// ██████╔╝███████║   ██║   ██║     ███████║
// ██╔═══╝ ██╔══██║   ██║   ██║     ██╔══██║
// ██║     ██║  ██║   ██║   ╚██████╗██║  ██║
// ╚═╝     ╚═╝  ╚═╝   ╚═╝    ╚═════╝╚═╝  ╚═╝

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

func (pa Patch) String() string {
	return fmt.Sprintf("Patch, contains %s normal %s\nSides: %s and %s", pa.PointOn, pa.Normal, pa.Sides[0], pa.Sides[1])
}

// RotateZ rotates in place about Z axis
func (pa *Patch) RotateZ(a Radians) *Patch {
	pa.Plane.RotateZ(a)
	pa.Sides = []Vec{pa.Sides[0].RotateZ(a), pa.Sides[1].RotateZ(a)}
	return pa
}

// RotateZAbout rotates about Z axis through p
func (pa *Patch) RotateZAbout(p Vec, a Radians) *Patch {
	pa.Plane.RotateZ(a)
	cp := pa.Corner.Add(p)
	s0 := pa.Sides[0].Subtract(cp).RotateZ(a).Add(cp)
	s1 := pa.Sides[1].Subtract(cp).RotateZ(a).Add(cp)
	pa.Sides = []Vec{s0, s1}
	return pa
}

// Translate moves a patch
func (pa *Patch) Translate(by Vec) *Patch {
	pa.Plane.Translate(by)
	pa.Corner = pa.Corner.Add(by)
	return pa
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

	b := pa.Corner.Add(pa.Sides[0])
	c := pa.Corner.Add(pa.Sides[1])

	if inTriangle(whu, pa.Corner, b, c) {
		return whu, true
	}

	if inTriangle(whu, b, b.Add(pa.Sides[1]), c) {
		return whu, true
	}

	return where, false

}
