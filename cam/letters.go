package cam

// Letter is a character glyph of some sort
type Letter struct {
	Width, Height float64
	Draw          func(t *Turtle)
}

// Font is a map of (probably single letter) strings of the functions
//   which write those letters using a given turtle
type Font map[string]Letter

// Plain is a very basic plain font intended for rapid plasma cnc cutting
var Plain Font

// TypeTo outputs a string of letters to the given turtle
func (f Font) TypeTo(t *Turtle, txt string, spacing float64) *Turtle {
	for _, c := range txt {
		if letter, ok := f[string(c)]; ok {
			t.Mark()
			letter.Draw(t)
			t.Return().F(letter.Width + t.TextSpacing)
		}
	}
	return t
}

// GetLetter looks one up
func (f Font) GetLetter(txt string) Letter {
	if letter, ok := f[txt]; ok {
		return letter
	}
	return f["?"]
}

func init() {

	Plain = make(Font)

	space := func(t *Turtle) {}
	one := func(t *Turtle) { t.L().Jump(7).F(2).R().F(3).R().F(9).R().F(2).R().F(7).L().F(1) }
	two := func(t *Turtle) {
		t.L().F(3).Strafe(3, 3).F(1).L().F(3).R().F(2).R().F(4)
		t.MoveBy(1, -1).R().F(3).Strafe(3, 3).L().F(3).R().F(2).R().F(5)
	}
	three := func(t *Turtle) {
		t.F(5).L().F(9).L().F(5).L().F(2).L().F(3).R().F(1)
		t.Strafe(1, 2).Strafe(1, -2).F(2).R().F(3).L().F(2)
	}
	four := func(t *Turtle) {
		t.PenUp().F(3).PenDown().L().F(3).L().F(3).R().F(1)
		t.Strafe(5, 4).R().F(1).R().F(9).R().F(2)
	}
	five := func(t *Turtle) {
		t.F(5).L().F(5).Strafe(2, -3).R().F(3).L().F(2).L().F(5).L().F(3).Strafe(2, -3).F(2).R().F(3).L().F(2)
	}
	six := func(t *Turtle) {
		t.L().Jump(1).F(7).Strafe(1, 1).R().F(4).R().F(2).R().F(3).L().F(2).L().F(2).Strafe(1, 1).R().F(3).Strafe(1, 1).R().F(3).Strafe(1, 1)
	}
	seven := func(t *Turtle) {
		t.Jump(2).L().F(4).Strafe(3, 1).L().F(3).R().F(2).R().F(5).R().F(1).Strafe(4, 1).F(4).R().F(2)
	}
	eight := func(t *Turtle) {
		t.L().Jump(1).F(3).Strafe(1, 1).Strafe(1, -1).F(2).Strafe(1, 1).R().F(3).Strafe(1, 1).R().F(2).Strafe(1, 1).Strafe(1, -1).F(3)
		t.Strafe(1, 1).R().F(3).Strafe(1, 1)
	}
	nine := func(t *Turtle) {
		t.Jump(3).L().F(5).L().F(2).Strafe(1, 1).R().F(2).Strafe(1, 1).R().F(3).Strafe(1, 1).R().F(8).R().F(2)
	}
	zero := func(t *Turtle) {
		t.L().Jump(3).F(3).Strafe(3, 2).R().F(1).R().Strafe(3, -2).F(3).Strafe(3, 2).R().F(1).R().Strafe(3, -2)
	}
	tri := func(t *Turtle) {
		t.F(6).LDeg(120).F(6).LDeg(120).F(6)
	}
	edge := func(t *Turtle) {
		t.F(2).L().F(6).L().F(2).L().F(6)
	}
	open := func(t *Turtle) {
		t.F(4).L().F(8).L().F(4).L().F(2).L().F(2).R().F(4).R().F(2).L().F(2)
	}
	closed := func(t *Turtle) {
		t.F(4).L().F(4).L().F(4).L().F(4)
	}

	Plain["?"] = Letter{Width: 3, Height: 9, Draw: space} // TODO replace with a real 'not found' glyph
	Plain[" "] = Letter{Width: 3, Height: 9, Draw: space}
	Plain["1"] = Letter{Width: 3, Height: 9, Draw: one}
	Plain["2"] = Letter{Width: 5, Height: 9, Draw: two}
	Plain["3"] = Letter{Width: 5, Height: 9, Draw: three}
	Plain["4"] = Letter{Width: 5, Height: 9, Draw: four}
	Plain["5"] = Letter{Width: 5, Height: 9, Draw: five}
	Plain["6"] = Letter{Width: 5, Height: 9, Draw: six}
	Plain["7"] = Letter{Width: 5, Height: 9, Draw: seven}
	Plain["8"] = Letter{Width: 5, Height: 9, Draw: eight}
	Plain["9"] = Letter{Width: 5, Height: 9, Draw: nine}
	Plain["0"] = Letter{Width: 5, Height: 9, Draw: zero}
	Plain["P"] = Letter{Width: 6, Height: 6, Draw: tri}
	Plain["E"] = Letter{Width: 2, Height: 6, Draw: edge}
	Plain["O"] = Letter{Width: 4, Height: 8, Draw: open}
	Plain["C"] = Letter{Width: 4, Height: 4, Draw: closed}

}
