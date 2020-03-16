package main

// ██████╗  ██████╗  ██████╗ ██████╗
// ██╔══██╗██╔═══██╗██╔═══██╗██╔══██╗
// ██║  ██║██║   ██║██║   ██║██████╔╝
// ██║  ██║██║   ██║██║   ██║██╔══██╗
// ██████╔╝╚██████╔╝╚██████╔╝██║  ██║
// ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝

import (
	"math"

	ell "./ellipsoid"
	gl "./gl"
	v3 "./vec"
)

// DoorKind is what basic type of door is it
type DoorKind int

// Values of DoorKind
const (
	Hole DoorKind = iota // No door
	Rollup
	TiltUp
	SingleSwing
	DoubleSwing
)

// DoorOpens says how it will open
type DoorOpens int

// Values of DoorOpens
const (
	AlwaysOpen DoorOpens = iota // No door
	LeftIn
	LeftOut
	RightIn
	RightOut
	CenterIn
	CenterOut
	Bottom
	Top // (?)
)

// Clamp says what the door is clamped to in UI
type Clamp int

// Door is a door UI with associated geometry
type Door struct {
	*v3.Cutter
	Name          string
	Width, Height v3.Meters
	Opens         DoorOpens
	Kind          DoorKind
	Clamps        []Clamp // How is it clamped?
	//	Cutter        v3.Cutter
	Shell *EShell
}

// Values of Clamp
const (
	ClampNone    Clamp = iota // not clamped
	ClampFaceX                // Facing along X axis towards center
	ClampFaceY                // Facing along Y axis towards center
	ClampTangent              // Tangiental to ellipsoid
	ClampCenter               // Face the center of the ellipsoid
	ClampOnX                  // Position is on x axis
	ClampOnY                  // Position is on y axis
)

// clampFunc enforces clamping geometry constraints
type clampFunc = func(e ell.Ellipsoid, pos v3.Vec, norm v3.Vec) (v3.Vec, v3.Vec) // Transforms pos and normal

// NewDoor makes one
func NewDoor(eshell *EShell, width v3.Meters, height v3.Meters) *Door {

	clps := []Clamp{ClampTangent}
	d := Door{Width: width, Height: height, Clamps: clps,
		Kind: Hole, Opens: AlwaysOpen}

	p := v3.Y.Scale(eshell.E.W + 1).Add(v3.Z.Scale(eshell.Base * 1.3))
	//	a := v3.Deg90
	d.Cutter = v3.NewCutter(width, height, p, v3.Y.Scale(-1))
	d.Shell = eshell
	// doorPatch = v3.NewPatch(, v3.Y.Scale(-1), doorWide, doorHigh)

	return &d

}

// Translate by a vector
func (d *Door) Translate(v v3.Vec) *Door {
	//	fmt.Printf("Delta %s\n", v)
	d.Cutter = v3.NewCutter(d.Width, d.Height, d.Corner.Add(v), d.Normal)
	return d
}

// RotateZ rotates about Z axis
func (d *Door) RotateZ(a v3.Radians) *Door {
	d.Cutter = v3.NewCutter(d.Width, d.Height, d.Corner, d.Normal.RotateZ(a))
	return d
}

var noClamp clampFunc = func(e ell.Ellipsoid, pos v3.Vec, norm v3.Vec) (v3.Vec, v3.Vec) {
	return pos, norm
}

// clampFuncs clamp the door to particular constraints
var clampFuncs = map[Clamp]clampFunc{
	ClampNone: noClamp,
	ClampFaceX: func(e ell.Ellipsoid, pos v3.Vec, norm v3.Vec) (v3.Vec, v3.Vec) {
		if pos.Dot(v3.X) > 0 {
			return pos, v3.X.Scale(-1)
		}
		return pos, v3.X
	},
	ClampFaceY: func(e ell.Ellipsoid, pos v3.Vec, norm v3.Vec) (v3.Vec, v3.Vec) {
		if pos.Dot(v3.Y) > 0 {
			return pos, v3.Y.Scale(-1)
		}
		return pos, v3.Y
	},
	ClampTangent: func(e ell.Ellipsoid, pos v3.Vec, norm v3.Vec) (v3.Vec, v3.Vec) {
		a := v3.Radians(math.Atan(pos.X() / pos.Y()))
		return pos, e.NormalAt(a)
	},
}

// DoClamps applies the clamps
func (d *Door) DoClamps() {
	p := d.Cutter.Patch.Corner
	n := d.Cutter.Normal
	for _, c := range d.Clamps {
		p, n = clampFuncs[c](d.Shell.E, p, n)
	}
	d.Cutter = v3.NewCutter(d.Width, d.Height, p, n)
}

//pos := v3.NewSimVec(e.W*v3.Sin(a)*1.1, e.L*v3.Cos(a)*1.1, bf).Subtract(c.Wide.Scale(0.5))

// Display generates the lines to display a door
func (d *Door) Display(e *EShell) []gl.ColourLine {

	ls := []gl.ColourLine{}

	ls = append(ls, gl.LinesForPatch(d.Cutter.Patch, true, gl.Blue)...)

	for _, p := range d.Cutter.Walls {
		ls = append(ls, gl.LinesForPatch(p, true, gl.Blue)...)
		ls = append(ls, e.CutWithPatch(p)...)
	}

	return ls
}
