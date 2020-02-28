//  ██████╗ █████╗ ███╗   ███╗
// ██╔════╝██╔══██╗████╗ ████║
// ██║     ███████║██╔████╔██║
// ██║     ██╔══██║██║╚██╔╝██║
// ╚██████╗██║  ██║██║ ╚═╝ ██║
//  ╚═════╝╚═╝  ╚═╝╚═╝     ╚═╝

package cam

import (
	"fmt"
	"image/color"
	"math"
	"os/exec"

	"github.com/llgcode/draw2d/draw2dpdf"
	"github.com/llgcode/draw2d/draw2dsvg"
)

// Stuff for outputting to cnc

const (
	pi     = math.Pi
	deg360 = 2 * pi // radians for 360
	deg180 = pi
	deg90  = pi / 2
	deg45  = pi / 4
	d2r    = (2 * pi) / 360
	r2d    = 360 / (2 * pi)
)

// CurveTolerance is the allowable deviation from perfect curve, in mm nominally
var CurveTolerance = 0.05

// ██╗   ██╗███████╗ ██████╗██████╗
// ██║   ██║██╔════╝██╔════╝╚════██╗
// ██║   ██║█████╗  ██║      █████╔╝
// ╚██╗ ██╔╝██╔══╝  ██║     ██╔═══╝
//  ╚████╔╝ ███████╗╚██████╗███████╗
//   ╚═══╝  ╚══════╝ ╚═════╝╚══════╝

// Vec2 is a 2D vector, nominally in mm
type Vec2 struct {
	X, Y float64 // mm
}

// Origin is (0,0)
var Origin = Vec2{X: 0, Y: 0}

// NewVec2 makes a new one
func NewVec2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
}

// Add adds a vector to this one
func (v Vec2) Add(v2 Vec2) Vec2 {
	return Vec2{X: v.X + v2.X, Y: v.Y + v2.Y}
}

// Subtract adds a vector to this one
func (v Vec2) Subtract(v2 Vec2) Vec2 {
	return Vec2{X: v.X - v2.X, Y: v.Y - v2.Y}
}

// Scale does so uniformly in x & y
func (v Vec2) Scale(k float64) Vec2 {
	return Vec2{X: k * v.X, Y: k * v.Y}
}

// Rotate does so by a radians -- note positive rotations are ANTI-CLOCKWISE!!! not like headings!!!
func (v Vec2) Rotate(a float64) Vec2 {
	cos := math.Cos(a)
	sin := math.Sin(a)
	return Vec2{X: cos*v.X - sin*v.Y, Y: sin*v.X + cos*v.Y}
}

// Length returns the length of the vector
func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) String() string {
	return fmt.Sprintf("(%.4g,%.4g 📏%.4g)", v.X, v.Y, v.Length())
}

// ██████╗  █████╗ ████████╗██╗  ██╗
// ██╔══██╗██╔══██╗╚══██╔══╝██║  ██║
// ██████╔╝███████║   ██║   ███████║
// ██╔═══╝ ██╔══██║   ██║   ██╔══██║
// ██║     ██║  ██║   ██║   ██║  ██║
// ╚═╝     ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝

// PathKind is the kind of path
type PathKind int

// Possible PathKind values
const (
	EdgePath PathKind = iota // a complete cut at the edge of a part or hole
	FoldPath                 // where work (eg sheet metal) is to be bent
	MarkPath                 // where the work is to be marked
	MetaPath                 // markings on the drawing not intended for cnc
)

// String renders pathkind in text
func (k PathKind) String() string {
	var s string
	switch k {
	case EdgePath:
		s = "Edge"
	case FoldPath:
		s = "Fold"
	case MarkPath:
		s = "Mark"
	case MetaPath:
		s = "Meta"
	default:
		s = "Unknown"
	}
	return s
}

// Segment is a straight line portion of a path
type Segment struct {
	Kind       PathKind
	Start, End Vec2 // position vectors of its start and end
}

// String renders a segment in text
func (s Segment) String() string {
	return fmt.Sprintf("Kind: %s, %s -> %s, δ%s", s.Kind, s.Start, s.End, s.End.Subtract(s.Start))
}

// Path is a set of linked segments in 2D
type Path struct {
	Segments []Segment
	Closed   bool
}

// Add adds a segment to the path
func (p *Path) Add(s Segment) *Path {
	p.Segments = append(p.Segments, s)
	return p
}

// Close joins the end of the path to its beginning
func (p *Path) Close() *Path {
	p.Closed = true
	l := len(p.Segments)
	if l > 0 {
		s := Segment{Kind: p.Segments[l-1].Kind,
			Start: p.Segments[l-1].End, End: p.Segments[0].Start}
		p.Add(s)
	}
	return p
}

// String prints out a path in text
func (p Path) String() string {
	s := fmt.Sprintf("Path has %d segments:\n", len(p.Segments))
	for _, seg := range p.Segments {
		s += fmt.Sprintf("   %s\n", seg)
	}
	if p.Closed {
		s += "Is Closed\n"
	} else {
		s += "Is Open\n"
	}
	return s
}

// ████████╗██╗   ██╗██████╗ ████████╗██╗     ███████╗
// ╚══██╔══╝██║   ██║██╔══██╗╚══██╔══╝██║     ██╔════╝
//    ██║   ██║   ██║██████╔╝   ██║   ██║     █████╗
//    ██║   ██║   ██║██╔══██╗   ██║   ██║     ██╔══╝
//    ██║   ╚██████╔╝██║  ██║   ██║   ███████╗███████╗
//    ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚══════╝

// Turtle is a logo-type graphics turtle
type Turtle struct {
	Position    Vec2
	Heading     float64  // radians, clockwise, from +Y
	TrailKind   PathKind // what kind of trail it is laying down currently
	Trailing    bool     // whether or not it is marking -> cf pen down/up
	Trail       Path     // The path so far ...
	Font        Font     // current font being used
	TextSpacing float64  // spacing between text letters
	forward     Vec2     // unit vector facing forward
	wasAt       Vec2     // remembered location
	wasFacing   float64  // remembered heading
}

// NewTurtle makes a default one, at the origin, facing 0, metamarking
func NewTurtle() Turtle {
	t := Turtle{Position: Origin, Heading: 0, TrailKind: MetaPath, Trailing: true}
	t.TurnTo(t.Heading)
	return t
}

func (t Turtle) String() string {
	s := fmt.Sprintf("At: %s, Heading: %.2g, Kind: %s\n", t.Position, t.Heading, t.TrailKind)
	s += t.Trail.String()
	return s
}

// F moves turtle forward
func (t *Turtle) F(distance float64) *Turtle {
	s := Segment{Kind: t.TrailKind, Start: t.Position,
		End: t.Position.Add(t.forward.Scale(distance))}
	t.Position = s.End
	if t.Trailing {
		t.Trail.Add(s)
	}
	return t
}

// B moves turtle backward
func (t *Turtle) B(distance float64) *Turtle {
	t.F(-distance)
	return t
}

// TurnBy turns r radians from current heading
func (t *Turtle) TurnBy(r float64) *Turtle {
	t.Heading += r
	t.forward.X = math.Sin(t.Heading)
	t.forward.Y = math.Cos(t.Heading)
	return t
}

// TurnTo turns to a given heading
func (t *Turtle) TurnTo(r float64) *Turtle {
	t.Heading = r
	t.forward.X = math.Sin(t.Heading)
	t.forward.Y = math.Cos(t.Heading)
	return t
}

// MoveTo moves directly to (x,y)
func (t *Turtle) MoveTo(x, y float64) *Turtle {
	s := Segment{Kind: t.TrailKind, Start: t.Position}
	t.Position.X = x
	t.Position.Y = y
	s.End = t.Position
	if t.Trailing {
		t.Trail.Add(s)
	}
	return t
}

// MoveBy moves relatively by (x,y)
func (t *Turtle) MoveBy(x, y float64) *Turtle {
	return t.MoveByVec(NewVec2(x, y))
}

// MoveByVec moves relatively BUT in *world* coords by vec2, use Strafe to move rel to heading
func (t *Turtle) MoveByVec(v Vec2) *Turtle {
	s := Segment{Kind: t.TrailKind, Start: t.Position}
	t.Position = t.Position.Add(v)
	s.End = t.Position
	if t.Trailing {
		t.Trail.Add(s)
	}
	return t
}

// JumpTo moves to (x,y) without leaving a trail, whatever the Trailing setting
func (t *Turtle) JumpTo(x, y float64) *Turtle {
	amTrailing := t.Trailing
	t.Trailing = false
	t.MoveTo(x, y)
	t.Trailing = amTrailing
	return t
}

// Jump moves forward by given distance without leaving a trail
func (t *Turtle) Jump(d float64) *Turtle {
	amTrailing := t.Trailing
	t.Trailing = false
	t.F(d)
	t.Trailing = amTrailing
	return t
}

// R turn +90 degrees
func (t *Turtle) R() *Turtle {
	t.TurnBy(deg90)
	return t
}

// R45 turns +45 degrees
func (t *Turtle) R45() *Turtle {
	t.TurnBy(deg45)
	return t
}

// L turns -90 degrees
func (t *Turtle) L() *Turtle {
	t.TurnBy(-deg90)
	return t
}

// L45 turns -45 degrees
func (t *Turtle) L45() *Turtle {
	t.TurnBy(-deg45)
	return t
}

// RDeg turn d degrees clockwise
func (t *Turtle) RDeg(d float64) *Turtle {
	t.TurnBy(d * d2r)
	return t
}

// LDeg turn d degrees anti-clockwise
func (t *Turtle) LDeg(d float64) *Turtle {
	t.TurnBy(-d * d2r)
	return t
}

// Strafe moves f forwards and r right (- = left) in a single line without changing heading
func (t *Turtle) Strafe(f, r float64) *Turtle {
	return t.MoveByVec(t.forward.Scale(f).Add(t.forward.Rotate(-deg90).Scale(r)))
}

// PenUp stops marking
func (t *Turtle) PenUp() *Turtle {
	t.Trailing = false
	return t
}

// PenDown starts marking
func (t *Turtle) PenDown() *Turtle {
	t.Trailing = true
	return t
}

// SetKind sets trail kind
func (t *Turtle) SetKind(k PathKind) *Turtle {
	t.TrailKind = k
	return t
}

// Curl turns the angle given at the radius given with
//   deviation from circularity less than tolerance
func (t *Turtle) Curl(radius float64, angle float64, tolerance float64) *Turtle {
	delta := 2 * (math.Acos(1 - tolerance/radius)) // angle of each step
	nSteps := int(angle / delta)
	if nSteps > 10000 { // prevent too many steps
		nSteps = 10000
	}
	delta = angle / float64(nSteps) // adjust to integer no of steps
	l := radius * math.Sin(delta)
	fmt.Printf("delta %.3g, steps %d, steplen %.3g\n", delta, nSteps, l)
	for i := 0; i <= nSteps; i++ { // TODO adjust to absolute perhaps?
		t.F(l).TurnBy(delta)
	}
	return t
}

// Mark sets the place and heading for Return
func (t *Turtle) Mark() *Turtle {
	t.wasAt = t.Position
	t.wasFacing = t.Heading
	return t
}

// Return goes back to the mark
func (t *Turtle) Return() *Turtle {
	return t.JumpTo(t.wasAt.X, t.wasAt.Y).TurnTo(t.wasFacing)
}

// SetFont sets the font and spacing
func (t *Turtle) SetFont(f Font, spacing float64) *Turtle {
	t.Font = f
	t.TextSpacing = spacing
	return t
}

// Type outputs a string of letters using the given font
func (t *Turtle) Type(txt string) *Turtle {
	for _, c := range txt {
		letter := t.Font.GetLetter(string(c))
		t.Mark()
		letter.Draw(t)
		t.Return().Jump(letter.Width + t.TextSpacing)
	}
	return t
}

// OutputPDF is
func (t Turtle) OutputPDF() {
	// Initialize the graphic context on an RGBA image
	dest := draw2dpdf.NewPdf("L", "mm", "A4")
	gc := draw2dpdf.NewGraphicContext(dest)

	// Set some properties
	gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(0.5)

	// Draw a closed shape
	gc.MoveTo(0, 0) // should always be called first for a new path
	for _, s := range t.Trail.Segments {
		gc.MoveTo(s.Start.X, 200-s.Start.Y)
		gc.LineTo(s.End.X, 200-s.End.Y)
	}
	gc.Close()
	gc.FillStroke()

	// Save to file
	draw2dpdf.SaveToPdfFile("turtle.pdf", dest)
	cmd := exec.Command("cmd", "/C start turtle.pdf")
	cmd.Start()
}

// OutputSVG is
func (t Turtle) OutputSVG() {
	// Initialize the graphic context on an RGBA image
	dest := draw2dsvg.NewSvg() //    NewSVG("L", "mm", "A4")
	gc := draw2dsvg.NewGraphicContext(dest)

	// Set some properties
	gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(0.5)

	// Draw a closed shape
	gc.MoveTo(0, 0) // should always be called first for a new path
	for _, s := range t.Trail.Segments {
		gc.MoveTo(s.Start.X, 200-s.Start.Y)
		gc.LineTo(s.End.X, 200-s.End.Y)
	}
	gc.Close()
	gc.FillStroke()

	// Save to file
	draw2dsvg.SaveToSvgFile("turtle.svg", dest)
	cmd := exec.Command("cmd", "/C start turtle.svg")
	cmd.Start()
}

// ██████╗ ██████╗  █████╗ ██╗    ██╗██╗███╗   ██╗ ██████╗
// ██╔══██╗██╔══██╗██╔══██╗██║    ██║██║████╗  ██║██╔════╝
// ██║  ██║██████╔╝███████║██║ █╗ ██║██║██╔██╗ ██║██║  ███╗
// ██║  ██║██╔══██╗██╔══██║██║███╗██║██║██║╚██╗██║██║   ██║
// ██████╔╝██║  ██║██║  ██║╚███╔███╔╝██║██║ ╚████║╚██████╔╝
// ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝ ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝

// Up is +Y, right is +X, origin is bottom left

// Drawing is a complete drawing designed for output to a CAM file of some sort
type Drawing struct {
	Name  string
	ID    int
	Paths []Path
}
