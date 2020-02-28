// ██╗   ██╗███████╗ ██████╗
// ██║   ██║██╔════╝██╔════╝
// ██║   ██║█████╗  ██║
// ╚██╗ ██╔╝██╔══╝  ██║
//  ╚████╔╝ ███████╗╚██████╗
//   ╚═══╝  ╚══════╝ ╚═════╝

package vec

import (
	"fmt"
	"math"
)

// Vec must do 3 vector things
type Vec interface {
	New(x, y, z float64) Vec
	X() float64
	Y() float64
	Z() float64
	SetX(x float64)
	SetY(y float64)
	SetZ(z float64)
	Length() float64
	LengthSq() float64
	Normalized() Vec
	Scale(f float64) Vec
	Dot(w Vec) float64
	Cross(w Vec) Vec
	Add(w Vec) Vec
	Subtract(w Vec) Vec
	String() string
	Stl() string
}

// unit axes etc.
var (
	X      = NewSimVec(1, 0, 0)
	Y      = NewSimVec(0, 1, 0)
	Z      = NewSimVec(0, 0, 1)
	Origin = NewSimVec(0, 0, 0)
)

// ███████╗██╗███╗   ███╗██╗   ██╗███████╗ ██████╗
// ██╔════╝██║████╗ ████║██║   ██║██╔════╝██╔════╝
// ███████╗██║██╔████╔██║██║   ██║█████╗  ██║
// ╚════██║██║██║╚██╔╝██║╚██╗ ██╔╝██╔══╝  ██║
// ███████║██║██║ ╚═╝ ██║ ╚████╔╝ ███████╗╚██████╗
// ╚══════╝╚═╝╚═╝     ╚═╝  ╚═══╝  ╚══════╝ ╚═════╝

// SimVec is a naive 3 vector
type SimVec struct {
	x, y, z float64
}

// SetX is obvious
func (v SimVec) SetX(newVal float64) {
	v.x = newVal
}

// SetY is obvious
func (v SimVec) SetY(newVal float64) {
	v.y = newVal
}

// SetZ is obvious
func (v SimVec) SetZ(newVal float64) {
	v.z = newVal
}

// Stl renders it as a string suitable for output in an stl file
func (v SimVec) Stl() string {
	return fmt.Sprintf("%E %E %E", v.X(), v.Y(), v.Z())
}

// String renders it as a string
func (v SimVec) String() string {
	return fmt.Sprintf("[X:%.3g, Y:%.3g, Z:%.3g 📏%.3g]", v.x, v.y, v.z, v.Length())
}

// Used only for calls to new
var simvec = SimVec{}

// NewSimVec makes one
func NewSimVec(x, y, z float64) SimVec {
	return simvec.New(x, y, z).(SimVec)
}

// New makes a new one from the given components. Values in receiver are ignored.
func (v SimVec) New(x, y, z float64) Vec {
	return SimVec{x: x, y: y, z: z}
}

// LengthSq returns the square of the length
func (v SimVec) LengthSq() float64 {
	return v.x*v.x + v.y*v.y + v.z*v.z
}

// Length returns the length
func (v SimVec) Length() float64 {
	return math.Sqrt(v.LengthSq())
}

// Normalized returns a copy, length 1
func (v SimVec) Normalized() Vec {
	l := v.Length()
	return v.New(v.x/l, v.y/l, v.z/l)
}

// X returns the X component
func (v SimVec) X() float64 {
	return v.x
}

// Y returns the Y component
func (v SimVec) Y() float64 {
	return v.y
}

// Z returns the Z component
func (v SimVec) Z() float64 {
	return v.z
}

// Scale returns a scaled version
func (v SimVec) Scale(f float64) Vec {
	return v.New(v.x*f, v.y*f, v.z*f)
}

// Add returns v+w
func (v SimVec) Add(w Vec) Vec {
	return v.New(v.x+w.X(), v.y+w.Y(), v.z+w.Z())
}

// Subtract returns v-w
func (v SimVec) Subtract(w Vec) Vec {
	return v.New(v.x-w.X(), v.y-w.Y(), v.z-w.Z())
}

// Dot is the dot product
func (v SimVec) Dot(w Vec) float64 {
	return v.x*w.X() + v.y*w.Y() + v.z*w.Z()
}

// Cross is the cross product
func (v SimVec) Cross(w Vec) Vec {
	cx := v.y*w.Z() - v.z*w.Y()
	cy := v.z*w.X() - v.x*w.Z()
	cz := v.x*w.Y() - v.y*w.X()
	return v.New(cx, cy, cz)
}

//  ██████╗██████╗ ██╗   ██╗██╗   ██╗███████╗ ██████╗
// ██╔════╝██╔══██╗██║   ██║██║   ██║██╔════╝██╔════╝
// ██║     ██████╔╝██║   ██║██║   ██║█████╗  ██║
// ██║     ██╔═══╝ ██║   ██║╚██╗ ██╔╝██╔══╝  ██║
// ╚██████╗██║     ╚██████╔╝ ╚████╔╝ ███████╗╚██████╗
//  ╚═════╝╚═╝      ╚═════╝   ╚═══╝  ╚══════╝ ╚═════╝

// CPUVec is a 3D vector optimized for CPU speed over storage size
type CPUVec struct {
	SimVec
	//	x, y, z    float64
	l, ll      float64 // length and length squared
	xN, yN, zN float64 // normalized x,y,z
}

// Used only for calls to new
var cpuvec = CPUVec{}

// Stl renders at Vector3 as a string suitable for output in an stl file
// func (v CPUVec) Stl() string {
// 	return fmt.Sprintf("%f %f %f", v.X(), v.Y(), v.Z())
// }

// // String renders a Vec as a string
// func (v CPUVec) String() string {
// 	return fmt.Sprintf("[X:%.3g, Y:%.3g, Z:%.3g 📏%.3g]", v.x, v.y, v.z, v.Length())
// }

// New makes a new one from the given components. Values in receiver are ignored.
func (v CPUVec) New(x, y, z float64) Vec {
	n := CPUVec{}
	n.x = x
	n.y = y
	n.z = z
	n.ll = x*x + y*y + z*z
	n.l = math.Sqrt(n.ll)
	n.xN = x / n.l
	n.yN = y / n.l
	n.zN = z / n.l
	return n
}

// X returns the X component
// func (v CPUVec) X() float64 {
// 	return v.x
// }

// // Y returns the Y component
// func (v CPUVec) Y() float64 {
// 	return v.y
// }

// // Z returns the Z component
// func (v CPUVec) Z() float64 {
// 	return v.z
// }

// // SetX is obvious
// func (v CPUVec) SetX(newVal float64) {
// 	v.x = newVal
// }

// // SetY is obvious
// func (v CPUVec) SetY(newVal float64) {
// 	v.y = newVal
// }

// // SetZ is obvious
// func (v CPUVec) SetZ(newVal float64) {
// 	v.z = newVal
// }

// NewCPUVec makes one
func NewCPUVec(x, y, z float64) CPUVec {
	return cpuvec.New(x, y, z).(CPUVec)
}

// LengthSq returns the square of the length
func (v CPUVec) LengthSq() float64 {
	return v.ll
}

// Length returns the length
func (v CPUVec) Length() float64 {
	return v.l
}

// Normalized returns a naive copy, length 1
func (v CPUVec) Normalized() Vec {
	return v.New(v.xN, v.yN, v.zN)
}

// Scale returns a scaled version
func (v CPUVec) Scale(f float64) Vec {
	return v.New(v.x*f, v.y*f, v.z*f)
}

// Simple returns a SimVec copy
func (v CPUVec) Simple() SimVec {
	return NewSimVec(v.x, v.y, v.z)
}

// Dot is the dot product
func (v CPUVec) Dot(w Vec) float64 {
	return v.x*w.X() + v.y*w.Y() + v.z*w.Z()
}

// Cross is the cross product
func (v CPUVec) Cross(w Vec) Vec {
	cx := v.y*w.Z() - v.z*w.Y()
	cy := v.z*w.X() - v.x*w.Z()
	cz := v.x*w.Y() - v.y*w.X()
	return v.New(cx, cy, cz)
}

// Add returns v+w
func (v CPUVec) Add(w Vec) Vec {
	return v.New(v.x+w.X(), v.y+w.Y(), v.z+w.Z())
}

// Subtract returns v-w
func (v CPUVec) Subtract(w Vec) Vec {
	return v.New(v.x-w.X(), v.y-w.Y(), v.z-w.Z())
}
