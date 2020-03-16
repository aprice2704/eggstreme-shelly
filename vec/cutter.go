package vec

//  ██████╗██╗   ██╗████████╗████████╗███████╗██████╗
// ██╔════╝██║   ██║╚══██╔══╝╚══██╔══╝██╔════╝██╔══██╗
// ██║     ██║   ██║   ██║      ██║   █████╗  ██████╔╝
// ██║     ██║   ██║   ██║      ██║   ██╔══╝  ██╔══██╗
// ╚██████╗╚██████╔╝   ██║      ██║   ███████╗██║  ██║
//  ╚═════╝ ╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═╝

// Cutter is a planar rectangular cutting tool with horizontal top and bottom and vertical sides
type Cutter struct {
	Patch          // The 'face' of the cutter, position is BL corner
	Width  Meters  // Width
	Height Meters  // Height
	Wide   Vec     // Vector from BL corner to BR corner
	High   Vec     // Vector from BL corner to TL corner
	Walls  []Patch // Sides of the cutter (original only)
}

// Locations of the sides in the array
const (
	CutterWallBottom = iota
	CutterWallTop
	CutterWallLeft
	CutterWallRight
	CutterWallNearEnd
	CutterWallFarEnd
)

// SidesOnly are the four Walls that are sides, not ends
var SidesOnly = []int{CutterWallBottom, CutterWallTop, CutterWallLeft, CutterWallRight}

// Translate by a vector
func (c Cutter) Translate(v Vec) *Cutter {
	newC := NewCutter(c.Width, c.Height, c.Corner.Add(v), c.Normal)
	return newC
}

// RotateZ rotates about Z axis
func (c Cutter) RotateZ(a Radians) *Cutter {
	newNorm := c.Normal.RotateZ(a)
	newC := NewCutter(c.Width, c.Height, c.Corner, newNorm)
	return newC
}

// SidesContain returns true iff the four sides (not ends) contain the given point
func (c Cutter) SidesContain(v Vec) bool {
	inside := true
	for _, s := range SidesOnly {
		if !c.Walls[s].Plane.NormalSide(v) {
			inside = false
			break
		}
	}
	return inside
}

// NewCutter makes a new one of width & height and position, at angle a (0=x,ccw)
func NewCutter(w, h Meters, p, normal Vec) *Cutter {

	// We are given the position of the bottom center of the door, need bottom left
	c := Cutter{Width: w, Height: h}

	hf := float64(h)
	wf := float64(w)
	c.Wide = Z.Cross(normal).Scale(-wf) // NewSimVec(wf*Cos(a), wf*Sin(a), 0)
	c.High = NewSimVec(0, 0, float64(hf))
	pos := p //p.Subtract(c.Wide.Scale(0.5))

	c.Patch = NewPatch(pos, normal, c.Wide, c.High)

	endPlane := YPlane
	fNormal := Y
	if c.Patch.Normal.X() < c.Patch.Normal.Y() { // we are facing the X plane plane more than the Y
		endPlane = XPlane
		fNormal = X
	}

	//	fmt.Printf("Cutter wide: %s\nEnd plane: %s\n", c.Wide, endPlane)

	blCorner := pos
	tlCorner := c.Patch.Corner.Add(c.High)
	brCorner := c.Patch.Corner.Add(c.Wide)
	trCorner := brCorner.Add(c.High)

	tlRay := NewLine(tlCorner, normal)
	trRay := NewLine(trCorner, normal)
	blRay := NewLine(c.Patch.Corner, normal)
	brRay := NewLine(brCorner, normal)

	blHit, hit0 := endPlane.IntersectLine(blRay)
	brHit, hit1 := endPlane.IntersectLine(brRay)
	tlHit, hit2 := endPlane.IntersectLine(tlRay)
	trHit, hit3 := endPlane.IntersectLine(trRay)

	if hit0 && hit1 && hit2 && hit3 {
		cwn := c.Wide.Normalized()

		bPatch := NewPatch(blCorner, Z, blHit.Subtract(blCorner), c.Wide)
		lPatch := NewPatch(blCorner, cwn, blHit.Subtract(blCorner), c.High)
		rPatch := NewPatch(brCorner, cwn.Scale(-1), brHit.Subtract(brCorner), c.High)
		tPatch := NewPatch(tlCorner, Z.Scale(-1), tlHit.Subtract(tlCorner), c.Wide)
		fPatch := NewPatch(blHit, fNormal, brHit.Subtract(blHit), trHit.Subtract(brHit))

		c.Walls = []Patch{bPatch, tPatch, lPatch, rPatch, fPatch}
	}

	return &c

}

// InitDisplay sets up displayable items for a cutter
// func (c *Cutter) InitDisplay() *Cutter {
// 	return c
// }
