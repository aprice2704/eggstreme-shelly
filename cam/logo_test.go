package cam

import (
	"fmt"
	"testing"
)

func TestTurtle(t *testing.T) {

	xeno := NewTurtle()
	xeno.JumpTo(50, 50)
	xeno.R().F(20).R().F(20).R().F(20)
	if xeno.Position.Subtract(NewVec2(50, 30)).Length() > 0.05 {
		fmt.Printf("%s", xeno)
		t.Error("Normal commands wandered off target")
	}

	touche := NewTurtle()
	touche.Curl(1000, deg180, 0.005)

	if touche.Position.Subtract(Vec2{2000, 0}).Length() > 0.05 {
		fmt.Printf("%s", touche)
		t.Error("Curl wandered off target")
	}

	mini := NewTurtle()
	//	mini.Curl(50, deg180, 0.005)

	// mini.MoveTo(1, 1).TurnTo(deg90)
	// Plain["2"].Draw(&mini)

	// mini.MoveTo(6, 1).TurnTo(deg90)
	// Plain["1"].Draw(&mini)

	mini.JumpTo(5, 5).F(20).R().F(20).SetFont(Plain, 1).Type("1234567890PEOC").F(20).R45().F(5).L().F(5).L().F(5).L().F(5)
	mini.JumpTo(5, 50).TurnTo(90*d2r).SetFont(Plain, 1).Type("P42 E280 O7")

	//	Plain.TypeTo(&mini, "12122", 2)

	fmt.Printf("%s", mini)
	mini.OutputSVG()

}
