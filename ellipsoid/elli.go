package ellipsoid

import (
	"math"
	"math/rand"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"

	v3 "../vec"
)

// Unit axes
var (
	X = v3.NewSimVec(1, 0, 0)
	Y = v3.NewSimVec(0, 1, 0)
	Z = v3.NewSimVec(0, 0, 1)
)

// Ellipsoid is the surface of an ellipsoid centered at the origin and aligned with the X, Y and Z axes.
type Ellipsoid struct {
	L, W, H       float64 // RADIAL size, not diameteral
	LL, WW, HH    float64 // squares of dims
	oL, oW, oH    float64 // one over the each dimensions
	oLL, oWW, oHH float64 // one over the square of each dimension
}

// Set sets the fields of an Ellipsoid, including time savers
func (e *Ellipsoid) Set(l, w, h float64) {
	e.L = l
	e.W = w
	e.H = h
	e.LL = l * l
	e.WW = w * w
	e.HH = h * h
	e.oL = 1 / l
	e.oW = 1 / w
	e.oH = 1 / h
	e.oLL = 1 / (l * l)
	e.oWW = 1 / (w * w)
	e.oHH = 1 / (h * h)
}

// XGivenYZ finds positive X coord of a point on the surface given others
func (e *Ellipsoid) XGivenYZ(y, z float64) float64 {
	return math.Sqrt(e.LL * (1 - ((y * y / e.WW) + (z * z / e.HH))))
}

// YGivenXZ finds positive X coord of a point on the surface given others
func (e *Ellipsoid) YGivenXZ(x, z float64) float64 {
	return math.Sqrt(e.WW * (1 - ((x * x / e.LL) + (z * z / e.HH))))
}

// ZGivenXY finds positive X coord of a point on the surface given others
func (e *Ellipsoid) ZGivenXY(x, y float64) float64 {
	return math.Sqrt(e.HH * (1 - ((x * x / e.LL) + (y * y / e.WW))))
}

// Surface finds where the vector, assumed to start at the origin, intersects with the surface of the ellipsoid
func (e Ellipsoid) Surface(dir v3.Vec) v3.Vec {
	v := dir.Normalized()
	k := math.Sqrt(1 / ((v.X() * v.X() * e.oLL) + (v.Y() * v.Y() * e.oWW) + (v.Z() * v.Z() * e.oHH)))
	//	fmt.Printf("k is %f, ", k)
	return v.Scale(k)
}

// PointDistant -- find a point s along the line starting at p defined by g projected onto e that is L from p (straight line) +- no more than l*tolerance
func (e Ellipsoid) PointDistant(p v3.Vec, g v3.Vec, L float64, tolerance float64) v3.Vec {

	P := p.Length()
	PP := P * P
	gN := g.Normalized()
	pN := p.Normalized()

	// Approximate e with a circle
	// K and M are lengths of components parallel and ortho to g in the plane of p & g
	// K = L^2/(2P), M = sqrt( L^2(1-L^2/4P^2))
	LL := L * L
	P2 := P * 2
	K := LL / P2
	M := math.Sqrt(LL * (1 - (LL / (4 * PP))))

	//fmt.Printf("L %f\nP %f\nk %f\nm %f\n", L, P, K, M)
	//fmt.Printf("Should be %f is %f\n", L, math.Sqrt(K*K+M*M))

	wN := gN.Cross(pN)
	vN := pN.Cross(wN)
	//fmt.Printf("wN %s\nvN %s\n", wN, vN)
	// s = p -pN*k + vN*M
	s := p.Subtract(pN.Scale(K)).Add(vN.Scale(M))

	estimate := e.Surface(s)
	diff := estimate.Subtract(p)
	actL := diff.Length()

	tries := 0
	delta := math.Abs(L - actL)
	for (delta > tolerance) && (tries < 10) {
		//		fmt.Printf("est %s;    Wanted %f got %f (δ %f)\n", estimate, L, actL, delta)
		diff = estimate.Subtract(p)
		actL = diff.Length()
		estimate = e.Surface(p.Add(diff.Scale(L / actL)))
		delta = math.Abs(estimate.Subtract(p).Length() - L)
		tries++
	}

	//	fmt.Printf("Final %s;    Wanted %f got %f (δ %f)\n", estimate, L, actL, L-actL)
	return estimate

}

// fmt.Printf("p   %s\nq   %s\ns   %s\nest %s\nWanted %f got %f (δ %f)\n",
// 	p, g, s, estimate, L, actL, L-actL)

// Humpty is an ellipsoid composed of lines
type Humpty struct {
	graphic.Lines
}

// NewHumpty makes one
func (e Ellipsoid) NewHumpty(n int, color math32.Color) *Humpty {

	hu := new(Humpty)
	r := rand.New(rand.NewSource(99))
	positions := math32.NewArrayF32(0, 0)

	for i := 0; i < n; i++ {
		p := v3.NewSimVec(2*(r.Float64()-0.5), 2*(r.Float64()-0.5), 2*(r.Float64()-0.5))
		q := e.Surface(p)
		positions.Append(
			0, 0, 0, color.R, color.G, color.B,
			float32(q.X()), float32(q.Z()), float32(q.Y()), color.R, color.G, color.B)
	}

	// Create geometry
	geom := geometry.NewGeometry()
	geom.AddVBO(
		gls.NewVBO(positions).
			AddAttrib(gls.VertexPosition).
			AddAttrib(gls.VertexColor),
	)

	// Create material
	mat := material.NewBasic()

	// Initialize lines with the specified geometry and material
	hu.Lines.Init(geom, mat)
	return hu

}

// Hat is a bunch of lines
type Hat struct {
	graphic.Lines
}

// NewHat makes one
func (e Ellipsoid) NewHat(p v3.Vec, dist float64, n int, color math32.Color) *Hat {

	hat := new(Hat)
	r := rand.New(rand.NewSource(99))
	positions := math32.NewArrayF32(0, 0)

	for i := 0; i < n; i++ {
		p2 := v3.NewSimVec(p.X()+2*(r.Float64()-0.5), p.Y()+2*(r.Float64()-0.5), p.Z())
		q := e.PointDistant(p, p2, dist, 0.00001)
		positions.Append(
			float32(p.X()), float32(p.Z()), float32(p.Y()), color.R, color.G, color.B,
			float32(q.X()), float32(q.Z()), float32(q.Y()), color.R, color.G, color.B)
	}

	// Create geometry
	geom := geometry.NewGeometry()
	geom.AddVBO(
		gls.NewVBO(positions).
			AddAttrib(gls.VertexPosition).
			AddAttrib(gls.VertexColor),
	)

	// Create material
	mat := material.NewBasic()

	// Initialize lines with the specified geometry and material
	hat.Lines.Init(geom, mat)
	return hat

}

// LatLongEllipsoid is a cage outline
type LatLongEllipsoid struct {
	graphic.Lines
}

// LatLong makes a conventional lat/long cage
func (e Ellipsoid) LatLong(nLat, nLong int, segs int, color math32.Color) *LatLongEllipsoid {

	eloid := new(LatLongEllipsoid)
	positions := math32.NewArrayF32(0, 0)

	halfPi := math.Pi / 2
	segStep := 2 * math.Pi / float64(segs)

	latStep := math.Pi / float64(nLat)
	lat := -halfPi
	for i := 0; i < nLat; i++ {
		z := math.Sin(lat)
		r := math.Cos(lat)
		var theta float64
		last := e.Surface(v3.NewSimVec(r*math.Cos(0), r*math.Sin(0), z))
		for j := 0; j <= segs; j++ {
			theta += segStep
			p := e.Surface(v3.NewSimVec(r*math.Cos(theta), r*math.Sin(theta), z))
			positions.Append(
				float32(last.X()), float32(last.Z()), float32(last.Y()), float32(math.Cos(theta)), float32(math.Sin(theta)), float32(r),
				float32(p.X()), float32(p.Z()), float32(p.Y()), float32(math.Cos(theta)), float32(math.Sin(theta)), float32(r),
			)
			last = p
		}
		lat += latStep
	}

	lonStep := math.Pi / float64(nLong)
	lon := -halfPi
	for i := 0; i < nLong; i++ {
		var theta float64
		last := e.Surface(v3.NewSimVec(math.Cos(lon), math.Sin(lon), 0))
		for j := 0; j <= segs; j++ {
			theta += segStep
			z := math.Sin(theta)
			r := math.Cos(theta)
			p := e.Surface(v3.NewSimVec(r*math.Cos(lon), r*math.Sin(lon), z))
			positions.Append(
				float32(last.X()), float32(last.Z()), float32(last.Y()), float32(math.Cos(lon)), float32(z), float32(r),
				float32(p.X()), float32(p.Z()), float32(p.Y()), float32(math.Cos(lon)), float32(z), float32(r),
			)
			last = p
		}
		lon += lonStep
	}

	// Create geometry
	geom := geometry.NewGeometry()
	geom.AddVBO(
		gls.NewVBO(positions).
			AddAttrib(gls.VertexPosition).
			AddAttrib(gls.VertexColor),
	)

	// Create material
	mat := material.NewBasic()

	// Initialize lines with the specified geometry and material
	eloid.Lines.Init(geom, mat)
	return eloid

}
