package entity

import "strings"

type CameraType = string

const (
	CameraTypeUnknown      CameraType = ""             // Default
	CameraType3D           CameraType = "3d"           // Stereo 3D
	CameraType360          CameraType = "360-degree"   // 360 Degree
	CameraTypeAction       CameraType = "action"       // GoPro
	CameraTypeComputer     CameraType = "computer"     // Webcams
	CameraTypeSurveillance CameraType = "surveillance" // Surveillance
	CameraTypeBody         CameraType = "body"         // SLR, DSLR, Mirrorless, Medium Format
	CameraTypeCompact      CameraType = "compact"      // Compact Cameras
	CameraTypeInstant      CameraType = "instant"      // Polaroid
	CameraTypePhone        CameraType = "phone"        // Mobile Phones
	CameraTypeTablet       CameraType = "tablet"       // Tablets
	CameraTypeFilm         CameraType = "film"         // Scanned Films
	CameraTypeScanner      CameraType = "scanner"      // Other Scanners
	CameraTypeVideo        CameraType = "video"        // Video Cameras
)

// CameraTypes maps internal model identifiers to camera types.
var CameraTypes = map[string]CameraType{
	"360 CAM":          CameraType360,
	"Mi Sphere":        CameraType360,
	"Scan":             CameraTypeScanner,
	"Scanner":          CameraTypeScanner,
	ModelMSScanner:     CameraTypeScanner,
	ModelPhotoScan:     CameraTypeScanner,
	"170 7472F20EEC14": CameraTypeScanner,
	ModelCanoScan:      CameraTypeScanner,
	ModelScanSnap:      CameraTypeScanner,
	ModelOpticFilm:     CameraTypeScanner,
	ModelSlideNScan:    CameraTypeFilm,
	"RODFS40":          CameraTypeFilm,
	"RODFS50":          CameraTypeFilm,
	"RODFS60":          CameraTypeFilm,
	"RODFS70":          CameraTypeFilm,
	"RODFS80":          CameraTypeFilm,
	"RODFS90":          CameraTypeFilm,
}

// GetCameraType determines the camera type based on the make and model name.
func GetCameraType(makeName string, modelName string) CameraType {
	// Detect the device type based on the exact model name.
	if result, found := CameraTypes[modelName]; found {
		return result
	}

	// Detect the device type for common brands.
	switch makeName {
	case MakeMotorola, MakeHTC, MakeLG, MakeOnePlus, MakeGarmin:
		return CameraTypePhone
	case MakeMinolta, MakeKonicaMinolta, MakePentax, MakeHasselblad, MakeSigma:
		return CameraTypeBody
	case MakeGoPro:
		return CameraTypeAction
	case MakeRaspberryPi:
		return CameraTypeComputer
	case MakePolaroid:
		return CameraTypeInstant
	case MakeRicoh:
		return CameraTypeCompact
	case MakeReolink, MakeVenTrade:
		return CameraTypeSurveillance
	case MakeApple:
		if strings.HasPrefix(modelName, ModelIPhone) {
			return CameraTypePhone
		} else if strings.HasPrefix(modelName, ModelIPad) {
			return CameraTypeTablet
		}
	case MakeGoogle:
		if strings.HasPrefix(modelName, "Pixel ") {
			if n := strings.ToLower(modelName); strings.Contains(n, "tab") || strings.Contains(n, "slate") {
				return CameraTypeTablet
			} else {
				return CameraTypePhone
			}
		}
	case MakeCanon:
		if strings.HasPrefix(modelName, "Lide") {
			return CameraTypeScanner
		} else if strings.HasPrefix(modelName, "EOS ") {
			if strings.HasPrefix(modelName, "EOS C") {
				return CameraTypeVideo
			}

			return CameraTypeBody
		} else if strings.HasPrefix(modelName, "Power") {
			return CameraTypeCompact
		}
	case MakeSony:
		if strings.HasPrefix(modelName, "Alpha") {
			return CameraTypeBody
		} else if strings.HasPrefix(modelName, "Xperia") {
			return CameraTypePhone
		} else {
			return CameraTypeCompact
		}
	case MakeHuawei:
		if n := strings.ToLower(modelName); strings.Contains(n, "tab") {
			return CameraTypeTablet
		} else if strings.HasPrefix(modelName, "P") ||
			strings.HasPrefix(modelName, "Mate ") ||
			strings.HasPrefix(modelName, "Honor ") {
			return CameraTypePhone
		}
	case MakeXiaomi:
		if n := strings.ToLower(modelName); strings.Contains(n, "tab") || strings.Contains(n, "pad") {
			return CameraTypeTablet
		} else {
			return CameraTypePhone
		}
	case MakeSamsung:
		if strings.HasPrefix(modelName, "Galaxy Tab") {
			return CameraTypeTablet
		} else if strings.HasPrefix(modelName, "Galaxy ") {
			return CameraTypePhone
		}
	case MakeHewlettPackard:
		if n := strings.ToLower(modelName); strings.Contains(n, "scan") ||
			strings.Contains(n, "laser") ||
			strings.Contains(n, "office") {
			return CameraTypeScanner
		}
	}

	// Try to recognize the device type by brand name.
	if n := strings.ToLower(makeName); n != "" {
		if strings.Contains(n, "scan") {
			return CameraTypeScanner
		}
	}

	// Try to recognize the device type by model name.
	if n := strings.ToLower(modelName); n != "" {
		if strings.Contains(n, "scan") {
			return CameraTypeScanner
		} else if strings.Contains(n, "pad") || strings.Contains(n, "tab") {
			return CameraTypeTablet
		} else if strings.Contains(n, "phone") || strings.Contains(n, "fone") || strings.Contains(n, "blackberry") {
			return CameraTypePhone
		} else if strings.Contains(n, "360") {
			return CameraType360
		}
	}

	return CameraTypeUnknown
}
