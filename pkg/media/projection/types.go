package projection

const (
	Unknown                     Type = ""
	Equirectangular             Type = "equirectangular"
	Cubestrip                   Type = "cubestrip"
	Cylindrical                 Type = "cylindrical"
	TransverseCylindrical       Type = "transverse-cylindrical"
	PseudocylindricalCompromise Type = "pseudocylindrical-compromise"
	Other                       Type = "other"
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
