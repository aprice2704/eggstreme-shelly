package gl

import (
	v3 "../vec"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

// Some colours
var (
	White   math32.Color = math32.Color{R: 1, G: 1, B: 1}
	Grey    math32.Color = math32.Color{R: .5, G: .5, B: .5}
	Red     math32.Color = math32.Color{R: 1, G: 0, B: 0}
	Blue    math32.Color = math32.Color{R: 0, G: 0, B: 1}
	Green   math32.Color = math32.Color{R: 0, G: 1, B: 0}
	Olive   math32.Color = math32.Color{R: 0, G: 0.5, B: 0}
	Yellow  math32.Color = math32.Color{R: 1, G: 1, B: 0}
	Fuchsia math32.Color = math32.Color{R: 1, G: 0, B: 1}
	Aqua    math32.Color = math32.Color{R: 0, G: 1, B: 1}
)

// ColourLine is a simple, evenly coloured line
type ColourLine struct {
	Start, End v3.Vec
	Colour     *math32.Color
}

// LineSet is a set of lines
type LineSet struct {
	graphic.Lines
	CLines []ColourLine
	mat    *material.Basic
}

// // VisPatch is a visible patch
// type VisPatch struct {
// 	*LineSet
// }

// NewLineSet sets it up, including the underlying gl stuff
func NewLineSet(lines []ColourLine, width float64) *LineSet {

	ls := LineSet{CLines: lines}
	nls := geometry.NewGeometry()
	ls.mat = material.NewBasic()
	ls.mat.SetLineWidth(float32(width))
	buff := math32.NewArrayF32(0, 12*len(lines))

	for _, l := range lines {
		buff = appendXZY(buff, l.Start)
		buff = appendColour(buff, *l.Colour)
		buff = appendXZY(buff, l.End)
		buff = appendColour(buff, *l.Colour)
	}

	nls.AddVBO(gls.NewVBO(buff).
		AddAttrib(gls.VertexPosition).
		AddAttrib(gls.VertexColor),
	)
	ls.Init(nls, ls.mat)
	ls.Lines.SetVisible(true)

	return &ls

}

// LinesForPatch makes an array of lines for the given patch
func LinesForPatch(p v3.Patch, norm bool, colour math32.Color) []ColourLine {

	a := p.Corner
	b := a.Add(p.Sides[0])
	c := a.Add(p.Sides[1])
	d := b.Add(p.Sides[1])

	lines := []ColourLine{
		ColourLine{Start: a, End: b, Colour: &colour},
		ColourLine{Start: a, End: c, Colour: &colour},
		ColourLine{Start: b, End: d, Colour: &colour},
		ColourLine{Start: c, End: d, Colour: &colour},
	}

	if norm {
		e := a.Add(d).Scale(0.5)
		f := e.Add(p.Normal.Scale(0.5)) //(p.Sides[0].Length() + p.Sides[1].Length()) / 4))
		lines = append(lines, ColourLine{Start: e, End: f, Colour: &White})
	}

	return lines

}

// Utils

func appendXZY(list []float32, vec v3.Vec) []float32 {
	return append(list, float32(vec.X()), float32(vec.Z()), float32(vec.Y()))
}

func appendColour(list []float32, c math32.Color) []float32 {
	return append(list, c.R, c.G, c.B)
}
