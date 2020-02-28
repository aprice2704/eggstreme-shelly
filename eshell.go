package main

import (
	"fmt"
	"log"
	"math"
	"sort"

	cam "./cam"
	ell "./ellipsoid"
	v3 "./vec"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/helper"

	"github.com/ztrue/tracerr"
)

// Global consts
var (
	DebugPurple = math32.Color{R: 0.9, G: 0, B: 0.9}
)

// Material is substance a panel may be made of
type Material struct {
	Base          MaterialBase // Basic substance
	Specific      string       // Specific variety e.g. alloy or steel or SS or Al etc.
	Thickness     float64      // what material thickness should it be made of?
	BendAllowance float64      // what length of unbent material does a 90deg bend require
	MinBendRadius float64      // what is the bend radius imparted by 90deg bend
}

// MaterialBase is the basic substance a panel may be made of
type MaterialBase int

// Values of Material
const (
	MatColdRolled MaterialBase = iota // Cold rolled steel
	MatHotRolled                      // Hot rolled steel
	MatStainless                      // 304
	MatAl                             // 6061
	MatTi                             // Titanium
	MatCu                             // Copper
	MatBrass                          // Brass
	MatExotic                         // Maraging steel etc., hardface, carbon fibre, glass, plastic
)

// FinishType is the basic variety of finish, more detail given in Specific
type FinishType int

// Values of FinishType
const (
	FinTypeNone     FinishType = iota // As it came from the factory
	FinTypeAbraded                    // simply sanded to some grade (see specific)
	FinTypeMetalDip                   // Dipped in a liquid metal, e.g. hot dipped galv
	FinTypeElectro                    // Electroplated
	FinTypeCoating                    // Coated in some non-metallic way
)

// SurfaceFinish is the basic type of finish to apply
type SurfaceFinish struct {
	Basic    FinishType // basic type of finish
	Specific string     // the colour, grade etc. wanted
}

// EShell is a set of panels covering an ellipsoid from its apex (+Z) to some horizontal plane (Z=base)
type EShell struct {
	E           ell.Ellipsoid // Ellipsoid shape on which this is based
	Base        float64       // Z=base is bottom plane
	Vertices    []Vertex      // all of them, indexed by int serial #
	Edges       []Edge        // all of them, indexed by int serial #
	Panels      []Panel       // all of them, indexed by int serial #
	PanelSize   float64       // desired panelsize during initial tessellation
	Tolerance   float64       // tolerance during panel edge length estimation
	FlangeWidth float64       // normal flange width expected for this design
	Step        int           //moribund?
	Cuts        []CutSegment  //TODO
	DebugLines  []DebugLine   //TODO
	// showSegs    []v3.Segment
	// showTris    []v3.Patch
}

// EShellMesh is just the g3n mesh
type EShellMesh struct {
	graphic.Mesh
	normals *helper.Normals
}

// CutSegment is a new segment defined by a cut
type CutSegment struct {
	start v3.Vec
	end   v3.Vec
}

// DebugLine is used for debugging
type DebugLine struct {
	Start  v3.Vec
	End    v3.Vec
	Colour math32.Color
}

// ██╗   ██╗███████╗██████╗ ████████╗███████╗██╗  ██╗
// ██║   ██║██╔════╝██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝
// ██║   ██║█████╗  ██████╔╝   ██║   █████╗   ╚███╔╝
// ╚██╗ ██╔╝██╔══╝  ██╔══██╗   ██║   ██╔══╝   ██╔██╗
//  ╚████╔╝ ███████╗██║  ██║   ██║   ███████╗██╔╝ ██╗
//   ╚═══╝  ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝

// Constraint is a function that enforces constraints on vertices
//type Constraint func(e *EShell, p v3.Vec) v3.Vec

// Constraints are functions that enforce constraints on vertices,
//   called during movement etc.
type Constraints []*func(e *EShell, p v3.Vec) v3.Vec

// Vertex is a point where panels meet
type Vertex struct {
	Serial      int
	Position    v3.Vec
	Normal      v3.Vec    // the normal at this vertex = average of normals of panels
	Edges       []int     // indexes into edges
	Panels      []int     // indexes into Panels
	V           v3.SimVec // velocity
	Shell       *EShell
	Alive       bool
	Constraints Constraints
}

// NiceString returns one to look at
func (v Vertex) NiceString() string {
	s := fmt.Sprintf("Vertex %d has %d edges %s and %d panels %s",
		v.Serial, len(v.Edges), List2String(v.Edges), len(v.Panels), List2String(v.Panels))
	return s
}

// OnEllipsoid forces the vertex to be on the surface of the ellipsoid
var OnEllipsoid = func(e *EShell, p v3.Vec) v3.Vec {
	return e.E.Surface(p)
}

// OnBase forces the vertex to be at the height of the base
var OnBase = func(e *EShell, p v3.Vec) v3.Vec {
	p.SetZ(e.Base)
	return p
}

// Move moves a vertex to a new position, while respecting contraints. Returns actual new position.
func (v *Vertex) Move(p v3.Vec) v3.Vec {
	dest := p
	for _, cst := range v.Constraints {
		dest = (*cst)(v.Shell, dest)
	}
	v.Position = dest
	return dest
}

// ComputeNormal averages the normals of the panels for this vertex and stores it
func (v *Vertex) ComputeNormal() {
	n := 0
	tot := v.Position.New(0, 0, 0)
	for _, panel := range v.Panels {
		p := &v.Shell.Panels[panel]
		if p.Alive {
			n++
			tot = tot.Add(p.Normal)
		}
	}
	v.Normal = tot.Scale(1 / float64(n))
}

// Combine does so to two lists of constraints producing a single sensible set
func Combine(c1, c2 Constraints) Constraints {
	c3 := c1
	for _, c := range c2 {
		found := false
		for _, d := range c1 {
			if c != d {
				found = true
				break
			}
		}
		if !found {
			c3 = append(c3, c)
		}
	}
	return c3
}

// ███████╗██╗      █████╗ ███╗   ██╗ ██████╗ ███████╗
// ██╔════╝██║     ██╔══██╗████╗  ██║██╔════╝ ██╔════╝
// █████╗  ██║     ███████║██╔██╗ ██║██║  ███╗█████╗
// ██╔══╝  ██║     ██╔══██║██║╚██╗██║██║   ██║██╔══╝
// ██║     ███████╗██║  ██║██║ ╚████║╚██████╔╝███████╗
// ╚═╝     ╚══════╝╚═╝  ╚═╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝

// FlangeStyle is the overall style of the flange
type FlangeStyle int

// FlangeStyle values
const (
	FStyleNone      FlangeStyle = iota // no flange
	FStyleGroundMk1                    // Simple ground flanges with holes for bolts or ground anchors, for structures up to 600 sq ft
	FStyleDoorMk1                      // Simple door flange for smallish doors (up to 8'x8')
)

// Flange is a rectangular flappy thing attached to an edge
type Flange struct {
	Edge    int         // which edge this flange is attached to
	Style   FlangeStyle // what sort of flange it is
	Depth   float64     // m, negative means away from origin
	Normal  v3.Vec      // normal to the plane of the flange
	Corners []v3.Vec    // 0..3 corners of the flange, in world coords
	Holes   []v3.Vec    // positions of any holes in the flange, world coords
	Dias    []float64   // and the diameters of the holes, m
}

// ███████╗██████╗  ██████╗ ███████╗
// ██╔════╝██╔══██╗██╔════╝ ██╔════╝
// █████╗  ██║  ██║██║  ███╗█████╗
// ██╔══╝  ██║  ██║██║   ██║██╔══╝
// ███████╗██████╔╝╚██████╔╝███████╗
// ╚══════╝╚═════╝  ╚═════╝ ╚══════╝

// EdgeTreatment is a type of flange, hem etc. applied to an edge of a panel.
type EdgeTreatment int

// Edge represents an edge on the constructed shell
type Edge struct {
	Serial    int           // id number
	Vertices  []int         // indexes into Vertices
	Panels    []int         // indexes into Panels
	Along     v3.Vec        // Vector along the edge from vertices[0] to vertices[1], not normalized
	Length    float64       // length of the edge
	Tension   float64       // negative is pull, positive is push
	Shell     *EShell       // shell its part of
	Alive     bool          // still part of display?
	Treatment EdgeTreatment // what type if edge should it be?
	HemSize   float64       // if a hem, this is the 'size' in m == distance from finished outer face to bottom most point/face of hem
	// note, therefore, that a closed/open pair of hems will have difference sizes in order to nest properly with outer faces even and
	// therefore, will depend on the thickness of the panel.
}

// EdgeTreatment values
const (
	ETreatAsCut        EdgeTreatment = iota // As it comes out of the CNC
	ETreatOpenHemMk1                        // regular open structural hem, worked with ClosedHemMk1
	ETreatClosedHemMk1                      // regualar close structural hem, works with OpenHemMk1
	ETreatTeardropHem                       // Small teardrop-style hem, no structural intent, merely smooth
	ETreatSmooth                            // Simply ground smooth with file, angle-grinder, dull beaver
	ETreatFlange                            // Details in separate struct
)

// Update recalcs the along vector after vertices have moved
func (ed *Edge) Update(e *EShell) {
	if !ed.Alive {
		return
	}
	ed.Along = e.Vertices[ed.Vertices[1]].Position.Subtract(e.Vertices[ed.Vertices[0]].Position)
}

// NiceString returns one to look at
func (ed Edge) NiceString() string {
	s := fmt.Sprintf("Edge %d has %d vertices %s and %d panels %s",
		ed.Serial, len(ed.Vertices), List2String(ed.Vertices), len(ed.Panels), List2String(ed.Panels))
	return s
}

// OtherEnd -- finds the vertex no of the end other than the one supplied
func (ed Edge) OtherEnd(this int) int {
	nd := ed.Vertices[0]
	if nd == this {
		nd = ed.Vertices[1]
	}
	return nd
}

// From returns the along vector of this edge, flipped if required, pointing away from the given vertex
func (ed Edge) From(v int) v3.Vec {
	fr := ed.Along
	if ed.Vertices[0] != v {
		if ed.Vertices[1] != v {
			err := tracerr.Errorf("Geometry error: vertex %d is not on edge %d at all (%v are)", v, ed.Serial, ed.Vertices)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
		fr = fr.Scale(-1)
	}
	return fr
}

// ██████╗  █████╗ ███╗   ██╗███████╗██╗
// ██╔══██╗██╔══██╗████╗  ██║██╔════╝██║
// ██████╔╝███████║██╔██╗ ██║█████╗  ██║
// ██╔═══╝ ██╔══██║██║╚██╗██║██╔══╝  ██║
// ██║     ██║  ██║██║ ╚████║███████╗███████╗
// ╚═╝     ╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝

// A panel is a triangular piece of geometry and the initial unit of tessellation
// It may be drawn in various ways for difference purposes, including being made up of sub pieces if it contains holes etc.

// PanelType defines whether the panel is a plain triangle, or is composed of several sub-triangles of geometry
//  which make up the full piece. (e.g. a panel with a hole in it, or around a hole corner).
type PanelType int

// PanelType values
const (
	PTypeNormal  PanelType = iota // a normal panel, unchanged in type from initial tessellation
	PTypeComplex                  // has a hole or internal corner, so needs extra information for display
)

// PanelAccessoryType is a type of accessory attached to this panel
type PanelAccessoryType int

// Panel is a single triangular panel of the structure
type Panel struct {
	Serial      int                // id number
	Corners     []int              // indexes into Vertices
	Edges       []int              // indexes into Edges
	Normal      v3.SimVec          // Normal, pointing away from origin
	InitNormal  v3.SimVec          // for flip detection
	Area        float64            // area in m2 of the outer extent of this panel
	Shell       *EShell            // Pointer back to owning shell
	OGLMaterial *material.Standard // moribund
	OGLVertices []int              // indices of the OpenGL vertex objects made for this panel (3 for each panel, not shared)
	Alive       bool               // should we render this panel in current software displays?
	Emit        bool               // should this panel be emitted as part of the final design?
	Accessory   PanelAccessoryType // what type of accessory, if any, is to be attached to this panel
	SubPanelOf  int                // serial number of panel from which this one was derived
	Kind        PanelType          // is this a simple, or complex, panel to render?
	Material    *Material          // what material should it be made from?
}

// Types of accessory on a panel
const (
	PAtypePlain     PanelAccessoryType = iota // No accessory
	PAtypeWindowMk1                           // Window, first version
	PAtypeVentMk1                             // Vent, first version
)

// Update recalculates the normal and area after an edge has moved
func (p *Panel) Update(e *EShell) {
	if !p.Alive {
		return
	}
	crx := e.Edges[p.Edges[0]].Along.Cross(e.Edges[p.Edges[1]].Along)
	p.Area = crx.Length() / 2
	p.Normal = crx.Normalized().(v3.SimVec)
}

// NiceString returns one to look at
func (p Panel) NiceString() string {
	s := fmt.Sprintf("Panel %d has %d edges %s and %d corners %s",
		p.Serial, len(p.Edges), List2String(p.Edges), len(p.Corners), List2String(p.Corners))
	return s
}

// STLString returns an STL rendering of this panel's outer geometrical face
func (p Panel) STLString() string {
	return fmt.Sprintf("facet normal %s\n outer loop\n  vertex %s\n  vertex %s\n  vertex %s\n endloop\nendfacet\n",
		p.Normal.Stl(), p.Shell.Vertices[p.Corners[0]].Position.Stl(), p.Shell.Vertices[p.Corners[1]].Position.Stl(),
		p.Shell.Vertices[p.Corners[2]].Position.Stl())
}

func (p Panel) EdgesWithCorner(c int) ([]int) {
	vNo := p.Corners[c]
	var es []int
	if p.Edges[0].Vertices[0] == vNo || p.Edges[0].Vertices[1] == vNo {
		es = append(es,p.Edges


// ClockwiseEdges ensures that the edges of a panel are listed in clockwise order
func (p *Panel) ClockwiseEdges() {
	vNo := p.Corners[0]

	e0 := &p.Shell.Edges[p.Edges[0]]
	e1 := &p.Shell.Edges[p.Edges[0]]
	e2 := &p.Shell.Edges[p.Edges[0]]
	if e0.Along.Cross(e1.Along).Dot(
}

// Draw does a CAM drawing of a panel
func (p *Panel) Draw(t *cam.Turtle) {
	if !p.Alive {
		return
	}
	for _, vNo := range p.Corners { // make sure the normals are accurate
		p.Shell.Vertices[vNo].ComputeNormal()
	}
	// Go around the edges
	for _, eNo := range p.Edges {
		e := &p.Shell.Edges[eNo]
		v0 := &p.Shell.Vertices[e.Vertices[0]]
		v1 := &p.Shell.Vertices[e.Vertices[1]]
		v0 = v1
	}
}

// HasVertex -- does this edge have an end at the given vertex
func (ed Edge) HasVertex(v int) bool {
	for _, vn := range ed.Vertices {
		if vn == v {
			return true
		}
	}
	return false
}

// ███████╗███████╗██╗  ██╗███████╗██╗     ██╗
// ██╔════╝██╔════╝██║  ██║██╔════╝██║     ██║
// █████╗  ███████╗███████║█████╗  ██║     ██║
// ██╔══╝  ╚════██║██╔══██║██╔══╝  ██║     ██║
// ███████╗███████║██║  ██║███████╗███████╗███████╗
// ╚══════╝╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝

// AddDryPanel checks if a panel would be entirely underwater, does not add if so
func (e *EShell) AddDryPanel(es []int, waterline float64) (pNo int, dry bool) {
	dry = false
	for _, ed := range es {
		for _, vNo := range e.Edges[ed].Vertices {
			if e.Vertices[vNo].Position.Z() > waterline {
				dry = true
			}
		}
	}
	if dry {
		pNo = e.AddPanel(es)
	}
	return pNo, dry
}

// AddPanel adds one to a shell
func (e *EShell) AddPanel(es []int) int {
	p := Panel{Accessory: PAtypePlain} // assume plain to begin with
	p.Edges = es
	crx := e.Edges[es[0]].Along.Cross(e.Edges[es[1]].Along)
	p.Area = crx.Length() / 2
	p.Normal = crx.Normalized().(v3.SimVec)
	v0 := e.Vertices[e.Edges[es[0]].Vertices[0]].Position
	if p.Normal.Dot(v0) < 0 { // its pointing inwards so flip it
		p.Normal = p.Normal.Scale(-1).(v3.SimVec)
	}
	p.InitNormal = p.Normal
	p.Shell = e
	p.Serial = len(e.Panels)
	p.Alive = true
	for _, f := range es { // record the new panel on each edge
		e.Edges[f].Panels = appendUnique(e.Edges[f].Panels, p.Serial)
		for _, v := range e.Edges[f].Vertices { // record the new panel on each vertex
			e.Vertices[v].Panels = appendUnique(e.Vertices[v].Panels, p.Serial)
			p.Corners = appendUnique(p.Corners, v)
		}
	}
	e.Panels = append(e.Panels, p)
	return p.Serial
}

// AddVertex adds one to a shell
func (e *EShell) AddVertex(v v3.Vec, cs Constraints) int {
	newV := Vertex{Position: v.(v3.SimVec), Serial: len(e.Vertices), Alive: true}
	e.Vertices = append(e.Vertices, newV)
	return newV.Serial
}

// RemovePanel removes one from a shell
func (e *EShell) RemovePanel(pNo int) {
	e.Panels[pNo].Alive = false
}

// RemoveVertex removes one for a shell
func (e *EShell) RemoveVertex(vNo int) {
	e.Vertices[vNo].Alive = false
}

// RemoveEdge removes one from a shell
func (e *EShell) RemoveEdge(eNo int) {
	e.Edges[eNo].Alive = false
}

// AddEdge adds one to a shell, does vertex housekeeping too
func (e *EShell) AddEdge(vs []int) int {
	al := e.Vertices[vs[1]].Position.Subtract(e.Vertices[vs[0]].Position)
	eno := len(e.Edges)
	newE := Edge{Vertices: vs, Along: al, Length: al.Length(), Serial: eno, Alive: true}
	e.Edges = append(e.Edges, newE)
	for _, v := range vs {
		e.Vertices[v].Edges = appendUnique(e.Vertices[v].Edges, eno)
	}
	return eno
}

// AntiSpike fills in gaps e=1p,v=6e,e=1p
func (e *EShell) AntiSpike() bool {
	var any bool
NextEdge:
	for _, ed := range e.Edges {
		edge1 := ed.Serial
		if len(ed.Panels) == 1 && ed.Alive == true {
			for _, vNo := range ed.Vertices {
				var v6s []int
				if len(e.Vertices[vNo].Edges) == 6 {
					v6s = append(v6s, vNo)
				}
				for _, v6 := range v6s { // all the v6s
					for _, edge2 := range e.Vertices[v6].Edges {
						if (edge2 != edge1) && len(e.Edges[edge2].Panels) == 1 { // we have some winners!
							if e.Vertices[v6].Position.Z() > e.Base { // only for dry v6's
								newEdge := e.AddEdge([]int{e.Edges[edge1].OtherEnd(v6), e.Edges[edge2].OtherEnd(v6)})
								e.AddPanel([]int{edge1, newEdge, edge2})
								any = true
								continue NextEdge
							}
						}
					}
				}
			}
		}
	}
	return any
}

// Spike adds a single tri to an edge if it is at least partly above the waterline
func (e *EShell) Spike(desiredL float64, tolerance float64) bool {
	var any bool
	for _, edge := range e.Edges {
		if len(edge.Panels) == 1 && edge.Alive == true { // this edge is part of only one panel
			//		eNo := edge.Serial
			p := e.Panels[edge.Panels[0]] // that panel
			vNo := edge.Vertices[0]       // one end of this edge
			// Find the edge of the panel that does not share the first vertex
			for _, ep := range p.Edges {
				oe := e.Edges[ep]
				if !oe.HasVertex(vNo) { // the one we want
					a := oe.From(edge.Vertices[1]).Scale(-1) // other end of this edge
					newPoint := e.E.PointDistant(e.Vertices[vNo].Position, a, desiredL, tolerance)
					if (newPoint.Z() > e.Base) ||
						(e.Vertices[vNo].Position.Z() > e.Base) ||
						(e.Vertices[edge.Vertices[1]].Position.Z() > e.Base) {
						newV := e.AddVertex(newPoint, Constraints{&OnEllipsoid})
						edge2 := e.AddEdge([]int{vNo, newV})
						edge3 := e.AddEdge([]int{newV, edge.Vertices[1]})
						e.AddPanel([]int{edge.Serial, edge2, edge3})
						any = true
						break
					}
				}
			}
		}
	}
	return any
}

// FillIn tris
func (e *EShell) FillIn(desiredL float64, tolerance float64) bool {
	var any bool
	for _, vertex := range e.Vertices {
		if len(vertex.Edges) == 5 && vertex.Alive == true && (vertex.Position.Z() > e.Base) { // 5 edges
			var twoEdges []int
			for _, eNo := range vertex.Edges {
				if len(e.Edges[eNo].Panels) == 1 { // edges part of only 1 panel
					twoEdges = append(twoEdges, eNo)
				}
			}
			if len(twoEdges) == 2 { // there should be 2
				me := vertex.Serial
				e1 := twoEdges[0]
				e2 := twoEdges[1]
				// Fill in with one tri or two?
				theta := math.Acos(e.Edges[e1].From(me).Normalized().Dot(e.Edges[e2].From(me).Normalized())) // angle between em
				//				fmt.Printf("Angle is %5.1f\n", theta*180/math.Pi)
				if theta < math.Pi/2 { // <90degrees so just one tri to fill in
					ne1 := e.AddEdge([]int{e.Edges[e1].OtherEnd(me), e.Edges[e2].OtherEnd(me)})
					e.AddPanel([]int{e1, ne1, e2})
					any = true
				} else { // two tris
					g := e.Edges[e1].From(me).Add(e.Edges[e2].From(me))
					p := e.E.PointDistant(vertex.Position, g, desiredL, tolerance) // new position
					pNo := e.AddVertex(p, Constraints{&OnEllipsoid})
					oe1 := e.Edges[e1].OtherEnd(vertex.Serial) // find the other ends
					oe2 := e.Edges[e2].OtherEnd(vertex.Serial)
					ne1 := e.AddEdge([]int{oe1, pNo})
					ne2 := e.AddEdge([]int{pNo, vertex.Serial})
					ne3 := e.AddEdge([]int{oe2, pNo})
					e.AddPanel([]int{e1, ne1, ne2})
					any = true
					e.AddPanel([]int{ne2, ne3, e2})
					any = true
				}
			}
		}
	}
	return any
}

// ShellLines is a wireframe version of a shell
type ShellLines struct {
	graphic.Lines
}

// PrepLines makes a Lines object for use in g3n for the eshell
func (e *EShell) PrepLines(mat *material.Basic) *ShellLines {

	s := ShellLines{}

	geom := geometry.NewGeometry()
	buff := math32.NewArrayF32(0, 3*2*6*(len(e.Panels)+len(e.Cuts)+len(showSegs)+len(showTris)))

	appendColour := func() {
		buff = append(buff, 1.0, 1.0, 0)
	}

	for _, panel := range e.Panels {

		if panel.Alive {

			if len(panel.Edges) != 3 {
				fmt.Printf("Geometry error! Panel %d has %d sides\n", panel.Serial, len(panel.Edges))
			}

			e0 := &e.Edges[panel.Edges[0]]
			e1 := &e.Edges[panel.Edges[1]]
			e2 := &e.Edges[panel.Edges[2]]

			vs := []int{e0.Vertices[0]}
			vs = appendUnique(vs, e0.Vertices[1])
			vs = appendUnique(vs, e1.Vertices[0])
			vs = appendUnique(vs, e1.Vertices[1])
			vs = appendUnique(vs, e2.Vertices[0])
			vs = appendUnique(vs, e2.Vertices[1])

			if len(vs) != 3 {
				fmt.Printf("Geometry error! Panel %d has %d edges and %d vertices\n", panel.Serial, len(panel.Edges), len(vs))
			}

			buff = appendXZY(buff, e.Vertices[vs[0]].Position)
			appendColour()
			buff = appendXZY(buff, e.Vertices[vs[1]].Position)
			appendColour()

			buff = appendXZY(buff, e.Vertices[vs[1]].Position)
			appendColour()
			buff = appendXZY(buff, e.Vertices[vs[2]].Position)
			appendColour()

			buff = appendXZY(buff, e.Vertices[vs[2]].Position)
			appendColour()
			buff = appendXZY(buff, e.Vertices[vs[0]].Position)
			appendColour()
		}
	}

	// Add the cut lines
	for _, ce := range e.Cuts {
		buff = appendXZY(buff, ce.start)
		buff = append(buff, 1.0, 0, 0)
		buff = appendXZY(buff, ce.end)
		buff = append(buff, 1.0, 0, 0)
	}

	// Add the debuglines
	for _, dl := range e.DebugLines {
		buff = appendXZY(buff, dl.Start)
		buff = append(buff, dl.Colour.R, dl.Colour.G, dl.Colour.B)
		buff = appendXZY(buff, dl.End)
		buff = append(buff, dl.Colour.R, dl.Colour.G, dl.Colour.B)
	}

	//	fmt.Printf("Tris %d, segs %d, debugs %d\n", len(showTris), len(showSegs), len(e.DebugLines))

	for _, seg := range showSegs {
		buff = appendXZY(buff, seg.Start())
		buff = append(buff, 1.0, 0, 0)
		buff = appendXZY(buff, seg.End())
		buff = append(buff, 1.0, 0, 0)
	}

	geom.AddVBO(
		gls.NewVBO(buff).
			AddAttrib(gls.VertexPosition).
			AddAttrib(gls.VertexColor),
	)
	s.Init(geom, mat)

	return &s
}

// Prep makes an OpenGL shellmesh for use in g3n for the eshell
func (e *EShell) Prep(mat *material.Standard) *EShellMesh {

	geom := geometry.NewGeometry()
	positions := math32.NewArrayF32(0, 2*3*3*len(e.Panels)) //
	normals := math32.NewArrayF32(0, 3*3*len(e.Panels))
	//	colours := math32.NewArrayF32(0, 3*3*len(e.Panels))
	indices := math32.NewArrayU32(0, 3*len(e.Panels))
	var idx uint32 // running index of the vertices

	for _, panel := range e.Panels {

		if panel.Alive {

			if len(panel.Edges) != 3 {
				fmt.Printf("Geometry error! Panel %d has %d sides\n", panel.Serial, len(panel.Edges))
			}

			e0 := &e.Edges[panel.Edges[0]]
			e1 := &e.Edges[panel.Edges[1]]
			e2 := &e.Edges[panel.Edges[2]]

			vs := []int{e0.Vertices[0]}
			vs = appendUnique(vs, e0.Vertices[1])
			vs = appendUnique(vs, e1.Vertices[0])
			vs = appendUnique(vs, e1.Vertices[1])
			vs = appendUnique(vs, e2.Vertices[0])
			vs = appendUnique(vs, e2.Vertices[1])

			if len(vs) != 3 {
				fmt.Printf("Geometry error! Panel %d has %d edges and %d vertices\n", panel.Serial, len(panel.Edges), len(vs))
			}

			positions = appendXZY(positions, e.Vertices[vs[0]].Position)
			positions = appendXZY(positions, e.Vertices[vs[1]].Position)
			positions = appendXZY(positions, e.Vertices[vs[2]].Position)

			normals = appendXZY(normals, panel.Normal)
			normals = appendXZY(normals, panel.Normal)
			normals = appendXZY(normals, panel.Normal)

			indices = append(indices, idx, idx+1, idx+2)
			idx += 3
		}
	}

	geom.SetIndices(indices)
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	//	geom.AddVBO(gls.NewVBO(colours).AddAttrib(gls.VertexColor))

	shell := EShellMesh{}
	shell.Mesh.Init(geom, mat)
	return &shell
}

// GLPatch is a visible patch
type GLPatch struct {
	graphic.Lines
}

// NewGLPatch makes a new visible asset for the given patch
func NewGLPatch(p v3.Patch, colour math32.Color) *GLPatch {

	mat := material.NewBasic()

	pat := geometry.NewGeometry()
	buff := math32.NewArrayF32(0, 4*2*2)

	a := p.Corner
	b := a.Add(p.Sides[0])
	c := a.Add(p.Sides[1])
	d := b.Add(p.Sides[1])

	buff = appendXZY(buff, a)
	buff = appendColour(buff, colour)
	buff = appendXZY(buff, b)
	buff = appendColour(buff, colour)

	buff = appendXZY(buff, a)
	buff = appendColour(buff, colour)
	buff = appendXZY(buff, c)
	buff = appendColour(buff, colour)

	buff = appendXZY(buff, b)
	buff = appendColour(buff, colour)
	buff = appendXZY(buff, d)
	buff = appendColour(buff, colour)

	buff = appendXZY(buff, c)
	buff = appendColour(buff, colour)
	buff = appendXZY(buff, d)
	buff = appendColour(buff, colour)

	pat.AddVBO(gls.NewVBO(buff).
		AddAttrib(gls.VertexPosition).
		AddAttrib(gls.VertexColor),
	)

	glpat := GLPatch{}
	glpat.Init(pat, mat)
	return &glpat

}

// STLString returns an STL representation of the panels in the shell
func (e EShell) STLString() string {
	s := "solid Eggstreme\n"
	// for i := 0; i < 3; i++ {
	// 	p := e.Panels[i]
	for _, p := range e.Panels {
		s += p.STLString()
	}
	s += "endsolid Eggstreme\n"
	return s
}

// ███████╗████████╗ █████╗ ████████╗███████╗
// ██╔════╝╚══██╔══╝██╔══██╗╚══██╔══╝██╔════╝
// ███████╗   ██║   ███████║   ██║   ███████╗
// ╚════██║   ██║   ██╔══██║   ██║   ╚════██║
// ███████║   ██║   ██║  ██║   ██║   ███████║
// ╚══════╝   ╚═╝   ╚═╝  ╚═╝   ╚═╝   ╚══════╝

// Stats is
func (e EShell) Stats(gs []gauge, ds []density) string {
	area := 0.0
	nPanels := 0
	nEdges := 0
	nSeams := 0
	nVertices := 0
	totPerim := 0.0
	for _, p := range e.Panels {
		if p.Alive {
			perim := 0.0
			for _, eNo := range p.Edges {
				perim += e.Edges[eNo].Along.Length()
			}
			area += p.Area + perim*2*e.FlangeWidth // doubled over flange
			totPerim += perim
			nPanels++
		}
	}
	for _, ed := range e.Edges {
		if ed.Alive {
			nEdges++
			if len(ed.Panels) == 2 {
				nSeams++
			}
		}
	}
	for _, v := range e.Vertices {
		if v.Alive {
			nVertices++
		}
	}

	s1 := fmt.Sprintf("Panels: %d,  Edges: %d inc %d Seamed,  Vertices: %d\nMidplane: %4.1f'x%4.1f' (%4.1fx%4.1fm)   Area: %4.0fsqft (%4.0fm2)",
		nPanels, nEdges, nSeams, nVertices,
		2*e.E.W*m2ft, 2*e.E.L*m2ft, 2*e.E.W, 2*e.E.L, e.E.W*m2ft*e.E.L*m2ft*math.Pi, e.E.W*e.E.L*math.Pi)

	s := fmt.Sprintf("%s\nMetal area needed: %4.1f sq ft (%4.1f sq m)\n", s1, area*sqM2sqFt, area)

	s += "       "
	for _, den := range ds {
		s += fmt.Sprintf("%10s", den.display)
	}
	s += "\n"
	for _, ga := range gs {
		s += fmt.Sprintf("%7s", ga.display)
		for _, de := range ds {
			s += fmt.Sprintf("  %8.0f", area*ga.thickness*de.rho)
		}
		s += "\n"
	}

	l2gal := 0.264172
	beadVol := 1000 * (totPerim / 2) * 0.003 * 0.003 * math.Pi / 4

	s += fmt.Sprintf("Total panel perimeter: %5.1f' (%5.1fm), 3mm bead volume: %.2gl (%.2ggal)\n", totPerim*m2ft, totPerim, beadVol, beadVol*l2gal)
	// Floor area calcs
	floorX := e.E.XGivenYZ(0, e.Base)
	floorY := e.E.YGivenXZ(0, e.Base)
	s += fmt.Sprintf("Floor is at %4.1g' (%4.1gm), peak is %4.1f' above it\n   It is %4.1f' x %4.1f' (%4.1fm x %4.1fm)   Area %4.1fsqft (%4.1fsqm)\n",
		e.Base*m2ft, e.Base, ((e.E.H)-e.Base)*m2ft, floorX*2*m2ft, floorY*2*m2ft, floorX*2, floorY*2, math.Pi*floorX*m2ft*floorY*m2ft, math.Pi*floorX*floorY)

	return fmt.Sprintf("%s\nStep %d", s, e.Step)
}

// Cleanup makes sure the references are consistent
// func (e *EShell) Cleanup() {
// 	a := e.CoupDeGrace()
// 	b := e.Undertaker()
// 	for a || b {
// 		a = e.CoupDeGrace()
// 		b = e.Undertaker()
// 	}
// }

// MakeMesh makes the initial mesh
func (e *EShell) MakeMesh(desiredL float64, tolerance float64) {

	pi := math.Pi
	cos := math.Cos
	sin := math.Sin
	deg60 := pi / 3

	// Start with a hexagonal patch at the zenith
	zenith := e.E.Surface(ell.Z)
	var ang float64
	e.AddVertex(zenith, Constraints{&OnEllipsoid}) // first vertex at zenith
	for i := 0; i < 6; i++ {
		e.AddVertex(e.E.PointDistant(zenith, ell.X.Scale(cos(ang)).Add(ell.Y.Scale(sin(ang))),
			desiredL, tolerance), Constraints{&OnEllipsoid})
		ang += deg60
	}
	e.AddEdge([]int{1, 2}) // nb order matters
	e.AddEdge([]int{2, 3})
	e.AddEdge([]int{3, 4})
	e.AddEdge([]int{4, 5})
	e.AddEdge([]int{5, 6})
	e.AddEdge([]int{6, 1})
	e.AddEdge([]int{0, 1})
	e.AddEdge([]int{0, 2})
	e.AddEdge([]int{0, 3})
	e.AddEdge([]int{0, 4})
	e.AddEdge([]int{0, 5})
	e.AddEdge([]int{0, 6})
	e.AddPanel([]int{6, 0, 7})
	e.AddPanel([]int{7, 1, 8})
	e.AddPanel([]int{8, 2, 9})
	e.AddPanel([]int{9, 3, 10})
	e.AddPanel([]int{10, 4, 11})
	e.AddPanel([]int{11, 5, 6})

	didSomething := true
	for didSomething {
		a := e.AntiSpike()
		b := e.FillIn(desiredL, tolerance)
		c := e.Spike(desiredL, tolerance)
		didSomething = a || b || c
	}

	e.CutFloor()

}

// func remove(value int, from []int) []int {
// 	n := []int{}
// 	for _, v := range from {
// 		if v != value {
// 			n = append(n, v)
// 		}
// 	}
// 	return n
// }

// CombineVertices transfers all references to v1 onto v0 and moves it to p
// func (e *EShell) CombineVertices(vNo0, vNo1 int, p v3.Vec) {
// 	v0 := &e.Vertices[vNo0]
// 	v1 := &e.Vertices[vNo1]
// 	v0.Position = p

// } TODO TODO

type edgeRef struct {
	serial int
	length float64
}

// PruneEdges tries to eliminate very short edges
func (e *EShell) PruneEdges(lengthLim float64) {

	var shorts []edgeRef

	for edi := range e.Edges {
		ed := &e.Edges[edi]
		eNo := ed.Serial
		if ed.Alive && (ed.Length < lengthLim) {
			shorts = append(shorts, edgeRef{serial: eNo, length: ed.Length})
		}
	}

	sort.Slice(shorts, func(i, j int) bool {
		return shorts[i].length < shorts[j].length
	})

	//	fmt.Printf("SHORTS: %s\n", shorts)

}

// CutPatch cuts all of the panels that intersect the given patch, tags the edges so formed with et
func (e *EShell) CutPatch(p v3.Patch, et EdgeTreatment) {

	// for pNo := range e.Panels {
	// }

}

// CutFloor cuts off all panels projecting below the floor
func (e *EShell) CutFloor() {
	floor := v3.NewPlane(v3.NewSimVec(0, 0, e.Base), v3.NewSimVec(0, 0, 1))
	//	debug := math32.Color{R: 0.1, G: 0.7, B: 0.4}
	for pNo := range e.Panels {
		p := &e.Panels[pNo]
		cutEnds := []v3.Vec{}
		for _, eNo := range p.Edges {
			ed := &e.Edges[eNo]
			vNo := ed.Vertices[0]
			f := ed.From(vNo)
			l := v3.NewLine(e.Vertices[vNo].Position, f)
			s := v3.NewSegment(l, 0.0, f.Length())
			//			e.DebugLines = append(e.DebugLines, DebugLine{Start: s.Start().Scale(1.01), End: s.End().Scale(1.01), Colour: debug})
			intersect, itsCut := floor.IntersectSegment(s)
			if itsCut {
				cutEnds = append(cutEnds, intersect)
				//		e.DebugLines = append(e.DebugLines, DebugLine{Start: v3.Origin, End: intersect, Colour: DebugPurple})
			}
		}
		if len(cutEnds) == 2 {
			//			e.Cuts = append(e.Cuts, CutSegment{start: cutEnds[0], end: cutEnds[1]})
			var aboves []int
			for _, vNo := range p.Corners {
				if e.Vertices[vNo].Position.Z() > e.Base {
					aboves = append(aboves, vNo)
				}
			}
			//			fmt.Printf("Panel %d above cut %d\n", pNo, len(aboves))
			if len(aboves) == 1 { // make one triangle
				vNew0 := e.AddVertex(e.E.Surface(cutEnds[0]), Constraints{&OnBase, &OnEllipsoid})
				vNew1 := e.AddVertex(e.E.Surface(cutEnds[1]), Constraints{&OnBase, &OnEllipsoid})
				eNew0 := e.AddEdge([]int{vNew0, vNew1})
				eNew1 := e.AddEdge([]int{aboves[0], vNew0})
				eNew2 := e.AddEdge([]int{aboves[0], vNew1})
				e.AddPanel([]int{eNew0, eNew1, eNew2})
			} else if len(aboves) == 2 { // need to make two
				vNew0 := e.AddVertex(e.E.Surface(cutEnds[0]), Constraints{&OnBase, &OnEllipsoid})
				vNew1 := e.AddVertex(e.E.Surface(cutEnds[1]), Constraints{&OnBase, &OnEllipsoid})
				eNew0 := e.AddEdge([]int{vNew0, vNew1})
				eNew1 := e.AddEdge([]int{aboves[0], vNew0})
				eNew2 := e.AddEdge([]int{aboves[0], vNew1})
				eNew3 := e.AddEdge([]int{aboves[1], vNew1})
				eNew4 := e.AddEdge([]int{aboves[1], aboves[0]})
				e.AddPanel([]int{eNew0, eNew1, eNew2})
				e.AddPanel([]int{eNew2, eNew3, eNew4})
			}
			e.RemovePanel(pNo)
		} else {
			if len(cutEnds) != 0 {
				fmt.Printf("ERROR: Panel %d has %d cut ends\n", pNo, len(cutEnds))
			}
		}

	}
}

// CalcTensions computes the tension/compression in each edge
func (e *EShell) CalcTensions(desired float64, k float64) {
	for eNo := 0; eNo < len(e.Edges); eNo++ {
		if e.Edges[eNo].Alive {
			e.Edges[eNo].Along = e.Vertices[e.Edges[eNo].Vertices[1]].Position.Subtract(e.Vertices[e.Edges[eNo].Vertices[0]].Position)
			e.Edges[eNo].Tension = k * math.Pow((e.Edges[eNo].Along.Length()-desired), 5) // tension = +ve
		}
	}
}

// MoveVertices moves them under action of the edges
func (e *EShell) MoveVertices(elli ell.Ellipsoid, moveFactor float64, slowFactor float64) {
	for vNo := 0; vNo < len(e.Vertices); vNo++ {
		var f v3.SimVec
		for _, eNo := range e.Vertices[vNo].Edges {
			if vNo == e.Edges[eNo].Vertices[0] {
				f = f.Add(e.Edges[eNo].Along.Normalized().Scale(e.Edges[eNo].Tension)).(v3.SimVec)
			} else {
				f = f.Add(e.Edges[eNo].Along.Normalized().Scale(-e.Edges[eNo].Tension)).(v3.SimVec)
			}
		}
		e.Vertices[vNo].V = e.Vertices[vNo].V.Add(f.Scale(moveFactor)).Scale(slowFactor).(v3.SimVec)
		e.Vertices[vNo].Position = elli.Surface(e.Vertices[vNo].Position.Add(e.Vertices[vNo].V)).(v3.SimVec)
	}
	for eNo := 0; eNo < len(e.Edges); eNo++ {
		if e.Edges[eNo].Alive {
			e.Edges[eNo].Update(e)
		}
	}
	for pNo := 0; pNo < len(e.Panels); pNo++ {
		if e.Panels[pNo].Alive {
			e.Panels[pNo].Update(e)
		}
	}
}

// IntersectsPanels find which panels a segment intersects
func (e *EShell) IntersectsPanels(seg v3.Segment) (panels []int, wheres []v3.Vec) {
	dnorm := math32.Color{R: 1, G: 0, B: 0}
	dsides := math32.Color{R: 0, G: 1, B: 1}
	showSegs = append(showSegs, seg)
	for i := 0; i < len(e.Panels); i++ {
		p := &e.Panels[i]
		ed := &e.Edges[p.Edges[0]]
		v := &e.Vertices[ed.Vertices[0]]
		v1 := &e.Vertices[ed.Vertices[1]]
		var ed2 *Edge // need to find which edge also has this vertex
		if e.Edges[p.Edges[1]].HasVertex(v.Serial) {
			ed2 = &e.Edges[p.Edges[1]]
		} else {
			ed2 = &e.Edges[p.Edges[2]]
		}
		f0 := &e.Vertices[ed2.Vertices[0]]
		f1 := &e.Vertices[ed2.Vertices[1]]
		tri := v3.NewPatch(v.Position, p.Normal, ed.From(v.Serial), ed2.From(v.Serial))
		whu, hits := tri.TriIntersectSegment(seg)
		if hits {
			panels = append(panels, p.Serial)
			wheres = append(wheres, whu)
			e.DebugLines = append(e.DebugLines, DebugLine{Start: v.Position, End: whu, Colour: DebugPurple})
			e.DebugLines = append(e.DebugLines, DebugLine{Start: v.Position, End: v.Position.Add(p.Normal), Colour: dnorm})
			e.DebugLines = append(e.DebugLines, DebugLine{Start: v.Position, End: v1.Position, Colour: dsides})
			e.DebugLines = append(e.DebugLines, DebugLine{Start: f0.Position, End: f1.Position, Colour: dsides})
			showTris = append(showTris, tri)
		}
	}
	return panels, wheres
}

// CheckGeometry does some basic checks on shell geometry
func (e *EShell) CheckGeometry() {
	for i := range e.Vertices {
		v := &e.Vertices[i]
		ne := len(v.Edges)
		if ne < 2 {
			err := tracerr.Errorf("Geometry error: vertex %d is on an incorrect number of edges: %d (%v)", v.Serial, ne, v.Edges)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
		np := len(v.Panels)
		if np > 6 || ne < 1 {
			err := tracerr.Errorf("Geometry error: vertex %d is on an incorrect number of panels: %d (%v)", v.Serial, np, v.Panels)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
	}
	for i := range e.Edges {
		ed := &e.Edges[i]
		nv := len(ed.Vertices)
		if nv != 2 {
			err := tracerr.Errorf("Geometry error: edge %d should have 2 vertices, has %d (%v)", ed.Serial, nv, ed.Vertices)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
		np := len(ed.Panels)
		if np > 2 || np < 1 {
			err := tracerr.Errorf("Geometry error: edge %d should be on 1 or 2 panels, is on %d (%v)", ed.Serial, np, ed.Panels)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
	}
	for i := range e.Panels {
		p := &e.Panels[i]
		nv := len(p.Corners)
		if nv != 3 {
			err := tracerr.Errorf("Geometry error: Panel %d should have 3 corners, has %d (%v)", p.Serial, nv, p.Corners)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
		ne := len(p.Edges)
		if ne != 3 {
			err := tracerr.Errorf("Geometry error: panel %d should have 3 edges, has %d (%v)", p.Serial, ne, p.Edges)
			tracerr.PrintSourceColor(err, 5, 2)
			log.Fatal(err)
		}
	}

}

// ██╗   ██╗████████╗██╗██╗     ███████╗
// ██║   ██║╚══██╔══╝██║██║     ██╔════╝
// ██║   ██║   ██║   ██║██║     ███████╗
// ██║   ██║   ██║   ██║██║     ╚════██║
// ╚██████╔╝   ██║   ██║███████╗███████║
//  ╚═════╝    ╚═╝   ╚═╝╚══════╝╚══════╝

func appendXZY(list []float32, vec v3.Vec) []float32 {
	return append(list, float32(vec.X()), float32(vec.Z()), float32(vec.Y()))
}

func appendColour(list []float32, c math32.Color) []float32 {
	return append(list, c.R, c.G, c.B)
}