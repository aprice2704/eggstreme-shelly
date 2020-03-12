package cam

// Materials is basic data for everything we use
var Materials MaterialSet

// MaterialBase is the basic substance a panel may be made of
type MaterialBase int

// Values of MaterialBase
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

// GaugeID is the unique (for this material) gauge name
type GaugeID string

// GaugeStats is the data for all likely gauges of a material
type GaugeStats map[GaugeID]SheetGauge

// SheetGauge in the info for a particular gauge of a particular material
type SheetGauge struct {
	Display       string  // Name to display
	ID            GaugeID //
	Thickness     float64 // m
	ArealDensity  float64 // kg/m2
	BendAllowance float64 // what length of unbent material does a 90deg bend require
	MinBendRadius float64 // what is the bend radius imparted by 90deg bend
}

// MaterialID is a unique identifier of a material
type MaterialID string

// Material is substance a panel may be made of -- this is as it arrives
type Material struct {
	ID          MaterialID
	Base        MaterialBase // Basic substance
	Specific    string       // Specific variety e.g. alloy or steel or SS or Al etc.
	DisplayName string       // Human friendly name
	Density     float64      // Kg/m3, estimated
	Element     string       // dominant constituent elements -- chemical symbols
	SheetData   GaugeStats   // used for display & estimation
}

// MaterialSet is just a map of them
type MaterialSet map[MaterialID]Material

// InputSheetTypeID identifier for sheet type
type InputSheetTypeID string

// InputSheetType a type of sheet used to make a component
type InputSheetType struct {
	ID       InputSheetTypeID // unique id
	Material MaterialID       // substance its made of
	Gauge    GaugeID          // what gauge
}

// FinishType is the basic variety of finish, more detail given in Specific
type FinishType int

// Values of FinishType
const (
	FinTypeNone     FinishType = iota // As it came from the factory
	FinTypeAbraded                    // simply sanded to some grade (see specific)
	FinTypeMetalDip                   // Dipped in a liquid metal, e.g. hot dipped galv
	FinTypeElectro                    // Electroplated
	FinTypeEPolish                    // Electro polished
	FinTypeCoating                    // Coated in some non-metallic way
)

// SurfaceFinish is the basic type of finish to apply
type SurfaceFinish struct {
	Basic    FinishType // basic type of finish
	Specific string     // the colour, grade etc. wanted
}

func init() {

	Materials = make(MaterialSet)

	mildgauges := GaugeStats{
		"28ga":       SheetGauge{Display: "28ga", ID: "28ga", Thickness: 0.378 / 1000},
		"24ga":       SheetGauge{Display: "24ga", ID: "24ga", Thickness: 0.607 / 1000},
		"22ga":       SheetGauge{Display: "22ga", ID: "22ga", Thickness: 0.759 / 1000},
		"20ga":       SheetGauge{Display: "20ga", ID: "20ga", Thickness: 0.911 / 1000},
		"18ga":       SheetGauge{Display: "18ga", ID: "18ga", Thickness: 1.214 / 1000},
		"16ga":       SheetGauge{Display: "16ga", ID: "16ga", Thickness: 1.518 / 1000},
		"14ga":       SheetGauge{Display: "14ga", ID: "14ga", Thickness: 1.897 / 1000},
		"0000000ga":  SheetGauge{Display: "0.5in", ID: "0000000ga", Thickness: 12.7 / 1000},
		"00000000ga": SheetGauge{Display: "1in", ID: "00000000ga", Thickness: 25.5 / 1000},
	}

	Materials["Stainless304"] = Material{ID: "Stainless304", Base: MatStainless, Specific: "304",
		DisplayName: "Stainless steel: 304", Density: 8030, Element: "Fe,Cr",
		SheetData: mildgauges} // TODO WRONG! update properly

	// densities := []density{
	// 	{display: "Steel", element: "Fe", rho: 7874},
	// 	{display: "Aluminium", element: "Al", rho: 2700},
	// 	{display: "Titanium", element: "Ti", rho: 4506},
	// }

}
