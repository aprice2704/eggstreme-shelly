package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"

	ell "./ellipsoid"
	v3 "./vec"

	_ "./statik"
	"github.com/rakyll/statik/fs"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/experimental/collision"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

const (
	m2ft     = 3.28084     // 1m in ft
	ft2m     = 1 / 3.28084 // 1' in m
	m2mm     = 1000.0      // 1m in mm
	mm2m     = 0.001       // 1mm in m
	sqM2sqFt = 10.7639     // 1 sq m to 1 sq ft
	sqFt2sqM = 1 / 10.7639 // other way
	deg90    = math.Pi / 2
)

type density struct {
	display string  // Name to display
	element string  // chemical symbol
	rho     float64 // density of substance in kg/m3
}

type gauge struct {
	display   string  // Name to display
	id        string  // integer
	thickness float64 // m
}

var showTris []v3.Patch
var showSegs []v3.Segment

// ███╗   ███╗ █████╗ ██╗███╗   ██╗
// ████╗ ████║██╔══██╗██║████╗  ██║
// ██╔████╔██║███████║██║██╔██╗ ██║
// ██║╚██╔╝██║██╔══██║██║██║╚██╗██║
// ██║ ╚═╝ ██║██║  ██║██║██║ ╚████║
// ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝

func main() {

	// cam.Opengltest()

	// Some local aliases
	pi := math.Pi
	cos := math.Cos
	sin := math.Sin
	deg60 := pi / 3

	gauges := []gauge{
		{display: "28ga", id: "28", thickness: 0.378 / 1000},
		{display: "24ga", id: "24", thickness: 0.607 / 1000},
		{display: "22ga", id: "22", thickness: 0.759 / 1000},
		{display: "20ga", id: "20", thickness: 0.911 / 1000},
		{display: "18ga", id: "18", thickness: 1.214 / 1000},
		{display: "16ga", id: "16", thickness: 1.518 / 1000},
		{display: "14ga", id: "14", thickness: 1.897 / 1000},
		{display: "0.5in", id: "0000000", thickness: 12.7 / 1000},
		{display: "1in", id: "00000000", thickness: 25.5 / 1000},
	}

	densities := []density{
		{display: "Steel", element: "Fe", rho: 7874},
		{display: "Aluminium", element: "Al", rho: 2700},
		{display: "Titanium", element: "Ti", rho: 4506},
	}

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	desiredL := 1.1     // desired size of panels
	tolerance := 0.0001 // tolerance in length approximations = 1/10th mm

	headroom := 12 * ft2m
	midWidth := 30 * ft2m
	midLength := 26 * ft2m
	midHeight := 20 * ft2m

	semiWidth := midWidth / 2
	semiLength := midLength / 2
	semiHeight := midHeight / 2
	midplaneRaised := headroom - semiHeight

	// Display shell as wireframe and/or shell
	wire := true
	shell := false

	// Show colourful ellipsoid
	ellipy := false

	// Show normal helpers
	norms := false

	doorColour := math32.Color{R: 0.7, G: 0.7, B: 0.9}

	ellipsoid := ell.Ellipsoid{}
	ellipsoid.Set(semiWidth, semiLength, semiHeight)
	wht := math32.Color{R: 1, G: 1, B: 1}
	eloid := ellipsoid.LatLong(60, 60, 100, wht)
	eloid.SetVisible(ellipy)

	eshell := EShell{E: ellipsoid}
	eshell.Base = -midplaneRaised
	eshell.PanelSize = desiredL
	eshell.Tolerance = tolerance
	eshell.FlangeWidth = 0.05 // 50 mm flanges when doubled over

	wireframe := &ShellLines{}

	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Add some furniture
	var grid *helper.Grid
	var ground *graphic.Mesh

	//steps := 0

	// ██╗   ██╗██╗
	// ██║   ██║██║
	// ██║   ██║██║
	// ██║   ██║██║
	// ╚██████╔╝██║
	//  ╚═════╝ ╚═╝

	col1 := float32(50)
	col2 := float32(140)
	col3 := float32(200)
	row := float32(40)

	var mygui *gui.Panel
	mygui = gui.NewPanel(700, 1000)
	mygui.SetRenderable(true)
	mygui.SetEnabled(true)
	mygui.SetColor4(&math32.Color4{R: 0, G: 0, B: 0.1, A: 0.5})

	fontFile := "/RobotoMono-Regular.ttf"
	r, err := statikFS.Open(fontFile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	fontData, err := ioutil.ReadAll(r)
	statsFont, err := text.NewFontFromData(fontData)
	if err != nil {
		fmt.Printf("Could not load font from %s\n", fontFile)
	}

	stats := gui.NewLabel("")
	stats.SetFont(statsFont)
	stats.SetPosition(col1, 400)
	mygui.Add(stats)

	inpFn := func(panel *gui.Panel, lab string, init string, unit string) *gui.Edit {
		lab1 := gui.NewLabel(lab)
		lab1.SetPosition(col1, row)
		var inp *gui.Edit
		inp = gui.NewEdit(50, init)
		inp.SetText(init)
		inp.SetPosition(col2, row)
		lab2 := gui.NewLabel(unit)
		lab2.SetPosition(col3, row)
		row += 22.0
		panel.Add(lab1)
		panel.Add(lab2)
		panel.Add(inp)
		return inp
	}

	lengthInput := inpFn(mygui, "Length", fmt.Sprintf("%4.1f", midLength*m2ft), "ft")
	widthInput := inpFn(mygui, "Width", fmt.Sprintf("%4.1f", midWidth*m2ft), "ft")
	heightInput := inpFn(mygui, "Height", fmt.Sprintf("%4.1f", midHeight*m2ft), "ft")
	headroomInput := inpFn(mygui, "Headroom", fmt.Sprintf("%4.1f", headroom*m2ft), "ft")
	panelInput := inpFn(mygui, "Panel", fmt.Sprintf("%4.1f", desiredL), "m")

	// ███████╗███████╗████████╗██╗   ██╗██████╗
	// ██╔════╝██╔════╝╚══██╔══╝██║   ██║██╔══██╗
	// ███████╗█████╗     ██║   ██║   ██║██████╔╝
	// ╚════██║██╔══╝     ██║   ██║   ██║██╔═══╝
	// ███████║███████╗   ██║   ╚██████╔╝██║
	// ╚══════╝╚══════╝   ╚═╝    ╚═════╝ ╚═╝

	var shellmesh *EShellMesh // the actual shell

	smat := material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
	smat.SetLineWidth(1)
	smat.SetWireframe(false)
	smat.SetSide(material.SideDouble)

	wiremat := material.NewBasic() // for the wireframe
	wiremat.SetLineWidth(2)
	wiremat.SetWireframe(true)
	//wiremat.SetSide(material.SideDouble)

	setupFunc := func() {

		eshell.MakeMesh(desiredL, tolerance) // compute the tris
		smat.SetWireframe(false)
		shellmesh = eshell.Prep(smat) // convert to opengl tris
		shellmesh.SetVisible(shell)
		scene.Add(shellmesh)

		shellmesh.normals = helper.NewNormals(shellmesh, 0.5, &math32.Color{R: 0, G: 0.7, B: 0}, 1)
		scene.Add(shellmesh.normals)
		shellmesh.normals.SetVisible(norms)

		wireframe = eshell.PrepLines(wiremat)
		wireframe.SetVisible(wire)
		scene.Add(wireframe)

		eloid = ellipsoid.LatLong(60, 60, 100, wht)
		eloid.SetVisible(ellipy)
		scene.Add(eloid)

		// Door tool 1
		doorPatch := v3.NewPatch(v3.NewSimVec(0, 0, eshell.Base),
			v3.X, v3.Y.Scale(8*ft2m), v3.Z.Scale(8*ft2m))
		door := NewGLPatch(doorPatch, doorColour)
		scene.Add(door)

		// Ground
		rgba, err := loadRGBA("/Nextgen_grass.jpg", statikFS)
		if err != nil {
			log.Fatal(err)
		}
		tex0 := texture.NewTexture2DFromRGBA(rgba)
		tex0.SetWrapS(gls.REPEAT)
		tex0.SetWrapT(gls.REPEAT)
		tex0.SetRepeat(100, 100)
		mat0 := material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
		mat0.AddTexture(tex0)
		mat0.SetSide(material.SideBack)
		groundGeom := geometry.NewSegmentedCube(100, 2)
		ground = graphic.NewMesh(groundGeom, nil)
		ground.AddGroupMaterial(mat0, 0)
		ground.SetVisible(shell)
		ground.RotateZ(-deg90)
		ground.SetPositionY(50 + float32(eshell.Base))

		// Add a grid
		gry := math32.Color{R: 0.2, G: 0.2, B: 0.2}
		grid = helper.NewGrid(20, 0.5, &gry)
		grid.TranslateY(float32(eshell.Base))
		grid.SetVisible(wire)
		scene.Add(grid)

		scene.Add(ground)

		stats.SetText(eshell.Stats(gauges, densities))

	}

	// ██████╗ ███████╗ ██████╗ ███████╗███╗   ██╗
	// ██╔══██╗██╔════╝██╔════╝ ██╔════╝████╗  ██║
	// ██████╔╝█████╗  ██║  ███╗█████╗  ██╔██╗ ██║
	// ██╔══██╗██╔══╝  ██║   ██║██╔══╝  ██║╚██╗██║
	// ██║  ██║███████╗╚██████╔╝███████╗██║ ╚████║
	// ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝

	// Regenerate the scene after the shell itself is changed
	regenFunc := func(name string, ev interface{}) {

		desiredL = floatIn(panelInput, desiredL)
		midLength = floatIn(lengthInput, midLength) * ft2m
		midWidth = floatIn(widthInput, midWidth) * ft2m
		headroom = floatIn(headroomInput, headroom) * ft2m
		midHeight = math.Max(floatIn(heightInput, midHeight)*ft2m, headroom*1.25) // >headroom
		heightInput.SetText(fmt.Sprintf("%4.1f", midHeight*m2ft))

		semiWidth := midWidth / 2
		semiLength := midLength / 2
		semiHeight := midHeight / 2
		midplaneRaised := headroom - semiHeight

		oldDebugs := eshell.DebugLines // preserve the debugs

		ellipsoid := ell.Ellipsoid{}
		ellipsoid.Set(semiWidth, semiLength, semiHeight)
		eshell = EShell{E: ellipsoid, DebugLines: oldDebugs}

		eshell.Base = -midplaneRaised
		eshell.PanelSize = desiredL
		eshell.Tolerance = tolerance
		eshell.FlangeWidth = 0.05 // 50 mm flanges when doubled over

		scene.Remove(shellmesh)
		eshell.MakeMesh(desiredL, tolerance) // compute the tris
		shellmesh = eshell.Prep(smat)
		shellmesh.SetVisible(shell)
		scene.Add(shellmesh)

		scene.Remove(wireframe)
		wireframe = eshell.PrepLines(wiremat) // convert to opengl tris
		wireframe.SetVisible(wire)
		scene.Add(wireframe)

		scene.Remove(shellmesh.normals)
		shellmesh.normals = helper.NewNormals(shellmesh, 0.5, &math32.Color{R: 0, G: 0.7, B: 0}, 1)
		shellmesh.normals.SetVisible(norms)
		scene.Add(shellmesh.normals)

		scene.Remove(eloid)
		eloid = ellipsoid.LatLong(60, 60, 100, wht)
		eloid.SetVisible(ellipy)
		scene.Add(eloid)

		scene.Remove(grid)
		gry := math32.Color{R: 0.2, G: 0.2, B: 0.2}
		grid = helper.NewGrid(20, 0.5, &gry)
		grid.TranslateY(float32(eshell.Base))
		grid.SetVisible(wire)
		scene.Add(grid)

		// Ground
		scene.Remove(ground)
		// tex0, err := texture.NewTexture2DFromImage("Nextgen_grass.jpg")
		// if err != nil {
		// 	log.Fatalf("Error loading texture: %s", err)
		// }
		// tex0.SetWrapS(gls.REPEAT)
		// tex0.SetWrapT(gls.REPEAT)
		// tex0.SetRepeat(100, 100)
		// mat0 := material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
		// mat0.AddTexture(tex0)
		// mat0.SetSide(material.SideBack)
		// groundGeom := geometry.NewSegmentedCube(100, 2)
		// ground = graphic.NewMesh(groundGeom, nil)
		// ground.AddGroupMaterial(mat0, 0)
		// ground.RotateZ(-deg90)
		// ground.SetPositionY(50 + float32(eshell.Base))
		// ground.SetVisible(shell)
		scene.Add(ground)

		stats.SetText(eshell.Stats(gauges, densities))

		//scramble = false

	}

	row += 15

	// wireframe button
	wireBtn := gui.NewButton("Wireframe")
	wireBtn.SetPosition(col1, row)
	wireBtn.SetSize(40, 18)
	wireBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		wire = !wire
		wireframe.SetVisible(wire)
		grid.SetVisible(wire)
	})
	mygui.Add(wireBtn)

	row += 25

	// shell button
	shellBtn := gui.NewButton("Textured, Shaded")
	shellBtn.SetPosition(col1, row)
	shellBtn.SetSize(40, 18)
	shellBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		shell = !shell
		shellmesh.SetVisible(shell)
		ground.SetVisible(shell)
	})
	mygui.Add(shellBtn)

	row += 25

	// Regen button
	regenBtn := gui.NewButton("Regenerate")
	regenBtn.SetPosition(col1, row)
	regenBtn.SetSize(40, 18)
	regenBtn.Subscribe(gui.OnClick, regenFunc)
	mygui.Add(regenBtn)

	row += 25

	// Cull edges button
	cullFunc := func(name string, ev interface{}) {
		//eshell.PruneEdges(desiredL * 0.1)
		// _, _, err := dlgs.FileMulti("Select files", "")
		// if err != nil {
		// 	panic(err)
		// }
	}
	cullBtn := gui.NewButton("Cull Short Edges")
	cullBtn.SetPosition(col1, row)
	cullBtn.SetSize(40, 18)
	cullBtn.Subscribe(gui.OnClick, cullFunc)
	mygui.Add(cullBtn)

	row += 40

	// normals button
	normsBtn := gui.NewButton("Normals")
	normsBtn.SetPosition(col1, row)
	normsBtn.SetSize(40, 18)
	normsBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		norms = !norms
		shellmesh.normals.SetVisible(norms)
	})
	mygui.Add(normsBtn)

	row += 25

	// ellipsoid button
	ellipyBtn := gui.NewButton("Ellipsoid")
	ellipyBtn.SetPosition(col1, row)
	ellipyBtn.SetSize(40, 18)
	ellipyBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		ellipy = !ellipy
		eloid.SetVisible(ellipy)
	})
	mygui.Add(ellipyBtn)

	row += 40

	// export STL button
	stlBtn := gui.NewButton("Export STL")
	stlBtn.SetPosition(col1, row)
	stlBtn.SetSize(40, 18)
	stlBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter filename: ")
		fname, _ := reader.ReadString('\n')
		fname = strings.TrimSpace(fname)
		if !strings.HasSuffix(fname, ".stl") {
			fname = fname + ".stl"
		}
		fmt.Printf("Will save in %s\n", fname)

		f, err := os.Create(fname)
		if err != nil {
			fmt.Printf("Error creating %s: %s\n", fname, err.Error())
			return
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		n, err := w.WriteString(eshell.STLString())
		if err != nil {
			fmt.Printf("Error writing %s: %s\n", fname, err.Error())
			return
		}
		w.Flush()
		fmt.Printf("Wrote %d bytes to %s\n", n, fname)

	})
	mygui.Add(stlBtn)

	scene.Add(mygui)

	// ███████╗ ██████╗███████╗███╗   ██╗███████╗
	// ██╔════╝██╔════╝██╔════╝████╗  ██║██╔════╝
	// ███████╗██║     █████╗  ██╔██╗ ██║█████╗
	// ╚════██║██║     ██╔══╝  ██║╚██╗██║██╔══╝
	// ███████║╚██████╗███████╗██║ ╚████║███████╗
	// ╚══════╝ ╚═════╝╚══════╝╚═╝  ╚═══╝╚══════╝

	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(-10, 10, 10)
	scene.Add(cam)
	orig := math32.Vector3{X: 0, Y: 0, Z: 0}
	zaxis := math32.Vector3{X: 0, Y: 1, Z: 0}
	cam.LookAt(&orig, &zaxis)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Scene setup
	onResize := func(evname string, ev interface{}) {
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	rc := collision.NewRaycaster(&math32.Vector3{}, &math32.Vector3{})

	onMouseDown := func(evname string, ev interface{}) {

		mev := ev.(*window.MouseEvent)
		if mev.Button != 1 {
			return
		}

		matrixWorld := (*shellmesh).MatrixWorld()
		var inverseMatrix math32.Matrix4
		inverseMatrix.GetInverse(&matrixWorld)

		width, height := a.GetSize()
		rcx := 2*(mev.Xpos/float32(width)) - 1
		rcy := -2*(mev.Ypos/float32(height)) + 1
		rc.SetFromCamera(cam, rcx, rcy)

		var ray math32.Ray
		ray.Copy(&rc.Ray).ApplyMatrix4(&inverseMatrix)

		rayOn := v3.NewSimVec(float64(ray.Origin().X), float64(ray.Origin().Z), float64(ray.Origin().Y))
		rayDir := v3.NewSimVec(float64(ray.Direction().X), float64(ray.Direction().Z), float64(ray.Direction().Y))

		seg := v3.NewSegment(v3.NewLine(rayOn, rayDir), 0.0, 50.0)
		showSegs = append(showSegs, seg)

		hitPanels, wheres := eshell.IntersectsPanels(seg)

		if len(hitPanels) > 0 {
			fmt.Printf("Hits: %d (%d)\n", len(hitPanels), len(wheres))
		} else {
			fmt.Println("MISSED!")
		}

	}

	a.Subscribe(window.OnMouseDown, onMouseDown)

	scene.Add(light.NewAmbient(&math32.Color{R: 1.0, G: 1.0, B: 1.0}, 0.6))

	var lights []*light.Point
	i := 0
	for ang := 0.0; ang < 2*pi; ang += deg60 {
		lights = append(lights, light.NewPoint(&math32.Color{R: 0.9, G: 0.9, B: 1}, 5000.0))
		lights[i].SetPosition(float32(-200*cos(ang)), 200, float32(200*sin(ang)))
		scene.Add(lights[i])
		i++
	}

	scene.Add(helper.NewAxes(0.5))

	//	a.Gls().ClearColor(0.53, 0.81, 0.92, 0.0) // sky blue
	a.Gls().ClearColor(0.0, 0.0, 0.0, 0.0) // sky blue

	stats.SetText(eshell.Stats(gauges, densities))

	// Compute the meshes etc.
	setupFunc()

	fmt.Printf("Panels: %d,  Edges: %d,  Vertices: %d\n", len(eshell.Panels), len(eshell.Edges), len(eshell.Vertices))

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})

}

// ██╗   ██╗████████╗██╗██╗     ███████╗
// ██║   ██║╚══██╔══╝██║██║     ██╔════╝
// ██║   ██║   ██║   ██║██║     ███████╗
// ██║   ██║   ██║   ██║██║     ╚════██║
// ╚██████╔╝   ██║   ██║███████╗███████║
//  ╚═════╝    ╚═╝   ╚═╝╚══════╝╚══════╝

// loadRGBA loads an image from a filesystem
func loadRGBA(name string, fs http.FileSystem) (rgba *image.RGBA, err error) {
	imfile, err := fs.Open(name)
	if err != nil {
		return rgba, fmt.Errorf("Error loading texture %s : %s", name, err)
	}
	img, _, err := image.Decode(imfile)
	rgba = image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return rgba, fmt.Errorf("Unsupported stride in %s", name)
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
	return rgba, nil
}

// Replace all occurences of a value with another
func replace(l []int, f, t int) []int {
	for i, v := range l {
		if v == f {
			l[i] = t
		}
	}
	return l
}

// append an int to an []int only if it that value is not already in the list
func appendUnique(l []int, x int) []int {
	for _, v := range l {
		if x == v {
			return l
		}
	}
	return append(l, x)
}

// append a slice of ints to another, only unique values
func appendSliceUnique(l []int, x []int) []int {
	var l2 []int
	for _, v := range x {
		l2 = appendUnique(l, v)
	}
	return l2
}

// remove any instances of a value from an []int
func remove(l []int, x int) []int {
	r := []int{}
	for _, v := range l {
		if v != x {
			r = append(r, v)
		}
	}
	return r
}

// List2String makes a nice compact string of an slice of ints
func List2String(l []int) string {
	s := "["
	for _, i := range l {
		s += fmt.Sprintf("%d,", i)
	}
	return s[0:len(s)-1] + "]"
}

// Read a float from the given text input, returning old if there is an error
func floatIn(ed *gui.Edit, old float64) float64 {
	s := strings.TrimSpace(ed.Text())
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Printf("Float conversion error text %s %s\n", s, err)
		return old
	}
	//	fmt.Printf("floatin old %4.1f text %s new %4.1f\n", old, s, f)
	return f
}

// DebugLines is a set of lines to help with debugging
type DebugLines struct {
	graphic.Lines
}

// MakeDebugs makes opengl structures to display debug data lines
func MakeDebugs() *DebugLines {
	d := &DebugLines{}
	return d
}

// if scramble || (steps > 0) {
// 	//	eshell.RemoveShortEdges(desiredL * 0.8)
// 	eshell.CalcTensions(desiredL, 0.1)
// 	eshell.MoveVertices(ellipsoid, 0.01, 0.955)
// 	eshell.Step++
// 	redraw = true
// 	steps--
// }
// if redraw {
// 	scene.Remove(shellmesh)
// 	shellmesh = eshell.Prep(smat)
// 	scene.Add(shellmesh)
// 	stats.SetText(eshell.Stats(gauges, densities))
// 	redraw = false
// }

// spotLight := light.NewSpot(&math32.Color{R: 0.9, G: 0.9, B: 1}, 10000.0)
// spotLight.SetPosition(40, 40, 40)
// spotLight.SetDirection(-1, -1, -1)
// spotLight.SetVisible(true)
// scene.Add(spotLight)

// for _, v := range s.Vertices {
// 	fmt.Println(v.NiceString())
// }

// for _, ed := range s.Edges {
// 	fmt.Println(ed.NiceString())
// }

// for _, p := range s.Panels {
// 	fmt.Println(p.NiceString())
// }

// eshell.MakeMesh(desiredL, tolerance) // compute the tris
// shellmesh = eshell.Prep(smat)        // convert to opengl tris
// scene.Add(shellmesh)

// Scramble button
// scrambleBtn := gui.NewButton("Scramble")
// var scramble bool // run the sim
//var redraw bool   // redraw the mesh
// scrambleBtn.SetPosition(50, 40)
// scrambleBtn.SetSize(40, 18)
// scrambleBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
// 	scramble = !scramble
// })
// mygui.Add(scrambleBtn)

// Some local aliases
// pi := math.Pi
// cos := math.Cos
// sin := math.Sin
// deg60 := pi / 3

// btnVisFunc := func(text string, toggle *bool, thing core.INode) {
// 	visBtn := gui.NewButton(text)
// 	visBtn.SetPosition(col1, row)
// 	visBtn.SetSize(40, 18)
// 	row += 22
// 	visBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
// 		*toggle = !*toggle
// 		thing.SetVisible(*toggle)
// 	})
// 	mygui.Add(visBtn)
// }

// btnVisFunc("Wireframe", &wire, wireframe)
// btnVisFunc("Shell", &shell, shellmesh)
// btnVisFunc("Ellipsoid", &ellipy, eloid)

//btnVisFunc("Normals", &norms, shellmesh.normals)

// Ground
// tex0, err := texture.NewTexture2DFromImage("Nextgen_grass.jpg")
// if err != nil {
// 	log.Fatalf("Error loading texture: %s", err)
// }
// mat0 := material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
// mat0.AddTexture(tex0)
// mat0.SetSide(material.SideDouble)
// groundGeom := geometry.NewSegmentedCube(100, 2)
// ground := graphic.NewMesh(groundGeom, nil)
// ground.AddGroupMaterial(mat0, 0)
// ground.SetVisible(true)
// ground.RotateZ(-deg90)
// ground.SetPositionY(50 + float32(eshell.Base))

// scene.Add(ground)
// fname, ok, err := dlgs.File("Select filename for export", "*.stl", false)
// 		if err != nil {
// 			fmt.Println("Error displaying file dialog")
// 			return
// 		}
// 		if !ok {
// 			fmt.Println("No filename chosen")
// 			return
// 		}
// 		fmt.Printf("Filename chosen: %s\n", fname)
// 	})
