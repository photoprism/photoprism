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
}

// Known maps names to standard projection types.
type Known map[string]Type
