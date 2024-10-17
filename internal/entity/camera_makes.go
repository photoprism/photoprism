package entity

const (
	MakeNone           = ""
	MakeAcer           = "Acer"
	MakeApple          = "Apple"
	MakeAsus           = "ASUS"
	MakeCanon          = "Canon"
	MakeNikon          = "NIKON"
	MakeGoogle         = "Google"
	MakeMotorola       = "Motorola"
	MakeLG             = "LG"
	MakeHTC            = "HTC"
	MakeGoPro          = "GoPro"
	MakeCasio          = "CASIO"
	MakeKodak          = "KODAK"
	MakeLeica          = "Leica"
	MakeOlympus        = "Olympus"
	MakeMinolta        = "Minolta"
	MakeKonicaMinolta  = "Konica Minolta"
	MakePentax         = "PENTAX"
	MakeSamsung        = "Samsung"
	MakeSony           = "SONY"
	MakeSharp          = "SHARP"
	MakeHuawei         = "HUAWEI"
	MakeXiaomi         = "Xiaomi"
	MakeFuji           = "FUJIFILM"
	MakeBlackBerry     = "BlackBerry"
	MakeRaspberryPi    = "Raspberry Pi"
	MakePolaroid       = "Polaroid"
	MakeHasselblad     = "Hasselblad"
	MakeSigma          = "SIGMA"
	MakeOnePlus        = "OnePlus"
	MakeHewlettPackard = "HP"
	MakeGarmin         = "Garmin"
	MakeRicoh          = "RICOH"
	MakeReolink        = "Reolink"
	MakeVenTrade       = "VenTrade"
)

// CameraMakes maps internal make identifiers to normalized names.
var CameraMakes = map[string]string{
	"acer":                        MakeAcer,
	"ACER":                        MakeAcer,
	"asus":                        MakeAsus,
	"Asus":                        MakeAsus,
	"ASUS_AI2302":                 MakeAsus,
	"apple":                       MakeApple,
	"Casio":                       MakeCasio,
	"CASIO COMPUTER":              MakeCasio,
	"CASIO COMPUTER CO":           MakeCasio,
	"CASIO COMPUTER CO.":          MakeCasio,
	"CASIO COMPUTER CO.,LTD":      MakeCasio,
	"CASIO COMPUTER CO.,LTD.":     MakeCasio,
	"CASIO CORPORATION":           MakeCasio,
	"Fujifilm":                    MakeFuji,
	"FUJIFILM CORPORATION":        MakeFuji,
	"Garmin-Asus":                 MakeGarmin,
	"google":                      MakeGoogle,
	"GOOGLE":                      MakeGoogle,
	"U63667E4U111368":             MakeGoogle,
	"Hewlett-Packard":             MakeHewlettPackard,
	"htc":                         MakeHTC,
	"Kodak":                       MakeKodak,
	"EASTMAN KODAK":               MakeKodak,
	"EASTMAN KODAK COMPANY":       MakeKodak,
	"GCMC":                        MakeKodak,
	"Leica Camera AG":             MakeLeica,
	"LEICA":                       MakeLeica,
	"LG Electronics":              MakeLG,
	"LGE":                         MakeLG,
	"lge":                         MakeLG,
	"Minolta Co., Ltd.":           MakeMinolta,
	"MINOLTA CO.,LTD":             MakeMinolta,
	"KONICA MINOLTA":              MakeKonicaMinolta,
	"Motorol":                     MakeMotorola,
	"Motorola Mobility":           MakeMotorola,
	"motorola":                    MakeMotorola,
	"samsung":                     MakeSamsung,
	"SAMSUNG":                     MakeSamsung,
	"Samsung Electronics":         MakeSamsung,
	"SAMSUNG TECHWIN Co.":         MakeSamsung,
	"SAMSUNG TECHWIN":             MakeSamsung,
	"Samsung Techwin":             MakeSamsung,
	"sharp":                       MakeSharp,
	"Sharp":                       MakeSharp,
	"sigma":                       MakeSigma,
	"Sigma":                       MakeSigma,
	"OLYMPUS":                     MakeOlympus,
	"OLYMPUS CORPORATION":         MakeOlympus,
	"OLYMPUS DIGITAL CAMERA":      MakeOlympus,
	"OLYMPUS IMAGING CORP.":       MakeOlympus,
	"OLYMPUS OPTICAL CO.,LTD":     MakeOlympus,
	"ONEPLUS":                     MakeOnePlus,
	"Nikon":                       MakeNikon,
	"NIKON CORPORATION":           MakeNikon,
	"Huawei":                      MakeHuawei,
	"XIAOMI":                      MakeXiaomi,
	"RaspberryPi":                 MakeRaspberryPi,
	"RICOH IMAGING COMPANY, LTD.": MakeRicoh,
	"Ricoh":                       MakeRicoh,
	"Pentax":                      MakePentax,
	"PENTAX Corporation":          MakePentax,
	"PENTAX CORPORATION":          MakePentax,
	"Blackberry":                  MakeBlackBerry,
	"Research In Motion":          MakeBlackBerry,
	"Sony":                        MakeSony,
	"VenTrade GmbH, Germany":      MakeVenTrade,
}
