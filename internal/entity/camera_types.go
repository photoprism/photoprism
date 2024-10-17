package entity

type CameraType = string

const (
	CameraUnknown    CameraType = ""
	CameraZLR        CameraType = "zlr"
	CameraSLR        CameraType = "slr"
	CameraDSLR       CameraType = "dslr"
	CameraMirrorless CameraType = "mirrorless"
	CameraCCTV       CameraType = "cctv"
	CameraSpy        CameraType = "spy"
	CameraAction     CameraType = "action"
	CameraWebcam     CameraType = "webcam"
	CameraDashcam    CameraType = "dashcam"
	CameraAnalog     CameraType = "analog"
	CameraBridge     CameraType = "bridge"
	CameraCompact    CameraType = "compact"
	CameraInstant    CameraType = "instant" // Polaroid
	CameraStereo     CameraType = "stereo"  // 3D
	CameraOmni       CameraType = "omni"    // 360 Degree
	CameraPhone      CameraType = "phone"
	CameraTablet     CameraType = "tablet"
	CameraMobile     CameraType = "mobile"
	CameraScanner    CameraType = "scanner"
	CameraMovie      CameraType = "movie"
	CameraVideo      CameraType = "video"
	CameraSoftware   CameraType = "software"
	CameraOther      CameraType = "other"
)

// CameraTypes maps internal model identifiers to camera types.
var CameraTypes = map[string]CameraType{
	"Scan":             CameraScanner,
	"Scanner":          CameraScanner,
	"MS Scanner":       CameraScanner,
	"PhotoScan":        CameraScanner,
	"170 7472F20EEC14": CameraScanner,
	"Slide N Scan":     CameraScanner,
	"RODFS40":          CameraScanner,
	"RODFS50":          CameraScanner,
	"RODFS60":          CameraScanner,
	"RODFS70":          CameraScanner,
	"RODFS80":          CameraScanner,
	"RODFS90":          CameraScanner,
}
