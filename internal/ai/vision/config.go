package vision

import (
	"net/http"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

var (
	AssetsPath            = fs.Abs("../../../assets")
	FaceNetModelPath      = fs.Abs("../../../assets/facenet")
	NsfwModelPath         = fs.Abs("../../../assets/nsfw")
	CachePath             = fs.Abs("../../../storage/cache")
	DownloadUrl           = ""
	ServiceUri            = ""
	ServiceKey            = ""
	ServiceTimeout        = time.Minute
	ServiceMethod         = http.MethodPost
	ServiceFileScheme     = scheme.Data
	ServiceRequestFormat  = ApiFormatVision
	ServiceResponseFormat = ApiFormatVision
	DefaultResolution     = 224
)
