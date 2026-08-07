package projection

const (
	// Unknown projection.
	Unknown Type = ""
	// Equirectangular projection type.
	Equirectangular Type = "equirectangular"
	// Cubestrip projection type.
	Cubestrip Type = "cubestrip"
	// Cylindrical projection type.
	Cylindrical Type = "cylindrical"
	// TransverseCylindrical projection type.
	TransverseCylindrical Type = "transverse-cylindrical"
	// PseudocylindricalCompromise projection type.
	PseudocylindricalCompromise Type = "pseudocylindrical-compromise"
	// Fisheye projection type, e.g. a single circular fisheye lens capture.
	Fisheye Type = "fisheye"
	// DualFisheye projection type, e.g. two side-by-side fisheye circles from a 360° camera.
	DualFisheye Type = "dual-fisheye"
	// Other projection type.
	Other Type = "other"
)

// Types maps identifiers to known types.
var Types = Known{
	string(Unknown):                     Unknown,
	string(Equirectangular):             Equirectangular,
	string(Cubestrip):                   Cubestrip,
	string(Cylindrical):                 Cylindrical,
	string(TransverseCylindrical):       TransverseCylindrical,
	string(PseudocylindricalCompromise): PseudocylindricalCompromise,
	string(Fisheye):                     Fisheye,
	string(DualFisheye):                 DualFisheye,
}

// Known maps names to standard projection types.
type Known map[string]Type
