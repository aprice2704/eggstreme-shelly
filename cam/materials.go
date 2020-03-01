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
	Display      string  // Name to display
	ID           GaugeID //
	Thickness    float64 // m
	ArealDensity float64 // kg/m2
}

// MaterialID is a unique identifier of a material
type MaterialID string

// Material is substance a panel may be made of -- this is as it arrives
type Material struct {
	ID            MaterialID
	Base          MaterialBase // Basic substance
	Specific      string       // Specific variety e.g. alloy or steel or SS or Al etc.
	DisplayName   string       // Human friendly name
	Thickness     float64      // what material thickness should it be made of?
	BendAllowance float64      // what length of unbent material does a 90deg bend require
	MinBendRadius float64      // what is the bend radius imparted by 90deg bend
	Density       float64      // Kg/m3
	SheetData     GaugeStats   //
	Element       string       // dominant constituent element -- chemical symbol
}

// MaterialSet is just a map of them
type MaterialSet map[MaterialID]Material

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

	Materials["Stainless"] = Material{}

	// mildgauges := []gauge{
	// 	{display: "28ga", id: "28", thickness: 0.378 / 1000},
	// 	{display: "24ga", id: "24", thickness: 0.607 / 1000},
	// 	{display: "22ga", id: "22", thickness: 0.759 / 1000},
	// 	{display: "20ga", id: "20", thickness: 0.911 / 1000},
	// 	{display: "18ga", id: "18", thickness: 1.214 / 1000},
	// 	{display: "16ga", id: "16", thickness: 1.518 / 1000},
	// 	{display: "14ga", id: "14", thickness: 1.897 / 1000},
	// 	{display: "0.5in", id: "0000000", thickness: 12.7 / 1000},
	// 	{display: "1in", id: "00000000", thickness: 25.5 / 1000},
	// }

	// densities := []density{
	// 	{display: "Steel", element: "Fe", rho: 7874},
	// 	{display: "Aluminium", element: "Al", rho: 2700},
	// 	{display: "Titanium", element: "Ti", rho: 4506},
	// }

}
