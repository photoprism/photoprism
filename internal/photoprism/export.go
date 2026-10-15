package photoprism

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/proc"
)

var (
	// ErrExportUnavailable reports an unsupported or unavailable source or rendition.
	ErrExportUnavailable = errors.New("media export unavailable")
	// ErrExportNotReady reports an absent or unverified sidecar output.
	ErrExportNotReady = errors.New("media export not ready")
	// ErrExportSourceChanged reports a source that no longer matches its indexed identity.
	ErrExportSourceChanged = errors.New("media export source changed")
	// ErrExportBusy reports that the bounded export capacity is occupied.
	ErrExportBusy    = errors.New("media export preparation busy")
	exportSlots      = make(chan struct{}, 4)
	exportProbeSlots = make(chan struct{}, 4)
)

// ExportOutput describes a verified or preparable media rendition.
type ExportOutput struct {
	Ready      bool   `json:"ready"`
	Preparable bool   `json:"preparable"`
	Format     string `json:"format"`
	Codec      string `json:"codec"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Bytes      int64  `json:"bytes"`
}

// ExportProbe describes source metadata and available video renditions.
type ExportProbe struct {
	Codec       string        `json:"codec"`
	Duration    float64       `json:"duration"`
	CaptureRole string        `json:"captureRole,omitempty"`
	AppleKey    string        `json:"appleKey"`
	Avc         *ExportOutput `json:"avc,omitempty"`
	VideoPart   *ExportOutput `json:"videoPart,omitempty"`
}

// exportSource retains the indexed identity and observed source version.
type exportSource struct {
	file        *entity.File
	media       *MediaFile
	stat        os.FileInfo
	version     *avcSourceVersion
	info        video.Info
	captureRole string
	companion   *exportSource
}

// exportSource resolves a selected file and admits only same-photo capture companions.
func (w *Convert) exportSource(f *entity.File) (*exportSource, error) {
	if f == nil || !fs.IsHash(f.FileHash) || f.FileMissing || f.FileError != "" || f.FileName == "" {
		return nil, ErrExportUnavailable
	}
	name := ConfigFileName(w.conf, f.FileRoot, f.FileName)
	version, err := newAvcSourceVersion(name)
	if err != nil || !version.info.Mode().IsRegular() || version.info.Size() == 0 {
		return nil, ErrExportUnavailable
	}
	m, err := NewMediaFile(name)
	if err != nil {
		return nil, ErrExportUnavailable
	}
	// Indexed metadata prevents a probe from discovering or creating unrelated sidecars.
	m.metaOnce.Do(func() {})
	m.metaData.Codec, m.metaData.Duration, m.metaData.Frames = f.FileCodec, f.FileDuration, f.FileFrames
	m.width, m.height = f.FileWidth, f.FileHeight
	info := video.NewInfo()
	if f.FileVideo || f.MediaType == entity.MediaLive || f.MediaType == entity.MediaVideo {
		info, err = w.exportVideoInfo(name)
		if err != nil {
			return nil, err
		}
	}
	m.videoOnce.Do(func() { m.videoInfo = info })
	if err = version.check(name); err != nil {
		return nil, err
	}
	m.hash = f.FileHash
	capture := FindInsta360Capture(m)
	m.exportCaptureResolved = true
	m.exportCapture = capture
	version.capture = capture
	var mate *entity.File
	if capture != nil {
		for _, member := range (MediaFiles{capture.Left, capture.Right}) {
			if member == nil {
				continue
			}
			if f.PhotoID != 0 && member.FileName() != m.FileName() {
				mate = &entity.File{}
				if err = entity.Db().Where("photo_id = ? AND file_root = ? AND file_name = ? AND file_missing = FALSE AND file_error = ?", f.PhotoID, f.FileRoot, member.RootRelName(), "").First(mate).Error; err != nil {
					return nil, ErrExportUnavailable
				}
				member.hash = mate.FileHash
			}
			member.metaOnce.Do(func() {})
			memberInfo, memberErr := w.exportVideoInfo(member.FileName())
			if memberErr != nil {
				return nil, memberErr
			}
			member.videoOnce.Do(func() { member.videoInfo = memberInfo })
			member.width, member.height = memberInfo.VideoWidth, memberInfo.VideoHeight
			member.metaData.Codec = string(memberInfo.VideoCodec)
		}
		if capture.Right == nil {
			version.missing = append(version.missing, capture.Name.FileName(media.Insta360VideoRight))
		}
		if capture.Left == nil {
			version.missing = append(version.missing, capture.Name.FileName(media.Insta360VideoLeft))
		}
	}
	var companion *exportSource
	role := ""
	if capture.ValidPair() {
		role = "secondary"
		if capture.Left.FileName() == m.FileName() {
			role = "main"
			if !capture.Dewarpable() {
				return nil, ErrExportUnavailable
			}
			right, versionErr := newAvcSourceVersion(capture.Right.FileName())
			if versionErr != nil {
				return nil, ErrExportSourceChanged
			}
			version.related = append(version.related, right)
			if mate != nil {
				companion = &exportSource{file: mate, media: capture.Right, stat: right.info, version: right}
			}
		}
	}
	return &exportSource{file: f, media: m, stat: version.info, version: version, info: info, captureRole: role, companion: companion}, nil
}

// exportVideoInfo probes containers directly, using a bounded FFprobe fallback for other videos.
func (w *Convert) exportVideoInfo(name string) (video.Info, error) {
	select {
	case exportProbeSlots <- struct{}{}:
		defer func() { <-exportProbeSlots }()
	default:
		return video.NewInfo(), ErrExportBusy
	}
	file, err := exportOpenFile(name, false)
	if err != nil {
		return video.NewInfo(), ErrExportUnavailable
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return video.NewInfo(), ErrExportUnavailable
	}
	info, err := video.Probe(file)
	info.FileSize, info.FileName = stat.Size(), name
	exportVideoDimensions(&info)
	if info.Encrypted {
		return info, ErrExportUnavailable
	}
	if err == nil && info.Compatible && info.VideoCodec != video.CodecUnknown && !info.Encrypted {
		return info, nil
	}
	if w.conf.FFprobeBin() == "" {
		return info, ErrExportUnavailable
	}
	args := []string{"-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name,width,height,duration:format=duration,format_name", "-of", "json"}
	args = append(args, encode.InputFormatArgs(name)...)
	args = append(args, name)
	cmd := exec.Command(w.conf.FFprobeBin(), args...) //nolint:gosec // source is resolved from an admitted indexed file
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = proc.Run(cmd, 30*time.Second); err != nil {
		return info, ErrExportUnavailable
	}
	var result struct {
		Streams []struct {
			Codec    string `json:"codec_name"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			Duration string `json:"duration"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if json.Unmarshal(out.Bytes(), &result) != nil || len(result.Streams) != 1 || result.Streams[0].Width <= 0 || result.Streams[0].Height <= 0 {
		return info, ErrExportUnavailable
	}
	stream := result.Streams[0]
	info.VideoCodec = video.Codec(stream.Codec)
	if stream.Codec == "h264" {
		info.VideoCodec = video.CodecAvc1
	}
	info.VideoWidth, info.VideoHeight = stream.Width, stream.Height
	duration, _ := strconv.ParseFloat(result.Format.Duration, 64)
	if duration <= 0 {
		duration, _ = strconv.ParseFloat(stream.Duration, 64)
	}
	info.Duration = time.Duration(duration * float64(time.Second))
	return info, nil
}

// exportVideoDimensions fills HEVC dimensions from the sole visual track header.
func exportVideoDimensions(info *video.Info) {
	if (info.VideoWidth <= 0 || info.VideoHeight <= 0) && len(info.TrackSizes) == 1 {
		info.VideoWidth, info.VideoHeight = info.TrackSizes[0].Width, info.TrackSizes[0].Height
	}
}

// ProbeExport reads source capabilities without starting a conversion or writing cache files.
func (w *Convert) ProbeExport(f *entity.File) (ExportProbe, error) {
	source, err := w.exportSource(f)
	if err != nil {
		return ExportProbe{}, err
	}
	result := ExportProbe{CaptureRole: source.captureRole, Codec: string(source.info.VideoCodec), Duration: source.info.Duration.Seconds(), AppleKey: w.exportAppleKey(f)}
	if source.captureRole == "secondary" {
		return result, nil
	}
	if f.MediaType == entity.MediaLive {
		if source.info.Compatible && !source.info.Encrypted && source.info.VideoOffset > 0 && source.info.VideoSize() > 0 && source.info.VideoWidth > 0 && source.info.VideoHeight > 0 {
			output := ExportOutput{Ready: true, Preparable: true, Format: strings.TrimPrefix(source.info.VideoFileExt(), "."), Codec: string(source.info.VideoCodec), Width: source.info.VideoWidth, Height: source.info.VideoHeight, Bytes: source.info.VideoSize()}
			result.VideoPart = &output
		}
	} else if (f.FileVideo || f.MediaType == entity.MediaVideo || f.MediaType == entity.MediaAnimated || f.IsAnimated()) && (!exportAvc(result.Codec) || source.captureRole == "main") {
		output := ExportOutput{Format: "mp4", Codec: string(video.CodecAvc1)}
		output.Preparable = w.conf.FFmpegEnabled() && w.conf.FFmpegBin() != "" && w.conf.SidecarWritable() && w.FFmpegAllowed(source.media) && source.media.CheckType() == nil
		if name, nameErr := w.avcName(source.media); nameErr != nil {
			output.Preparable = false
		} else if stat, statErr := os.Lstat(name); statErr == nil {
			output.Preparable = output.Preparable && stat.Mode().IsRegular() && stat.Size() == 0
		} else if !errors.Is(statErr, os.ErrNotExist) {
			output.Preparable = false
		}
		if _, ready, readyErr := w.exportSidecar(source.media); readyErr == nil {
			output = ready
		}
		if output.Ready || output.Preparable {
			result.Avc = &output
		}
	}
	return result, nil
}

// exportAvc recognizes source codec names and AVC sample entries.
func exportAvc(codec string) bool {
	switch strings.ToLower(codec) {
	case "avc", "avc1", "avc2", "avc3", "avc4", "h264", "h.264", "x264":
		return true
	default:
		return false
	}
}

// exportAppleKey reads only explicit Apple pairing tags from existing hash-bound metadata.
func (w *Convert) exportAppleKey(f *entity.File) string {
	if f.MediaType != entity.MediaImage && f.MediaType != entity.MediaVideo {
		return ""
	}
	dir, err := fs.CachePath(w.conf.CachePath(), f.FileHash, "json", false)
	if err != nil {
		return ""
	}
	name := filepath.Join(dir, f.FileHash+"_exiftool.json")
	stat, err := os.Lstat(name)
	if err != nil || !stat.Mode().IsRegular() {
		return ""
	}
	file, err := exportOpenFile(name, true)
	if err != nil {
		return ""
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 4*1024*1024+1))
	if err != nil || len(data) > 4*1024*1024 || !gjson.ValidBytes(data) {
		return ""
	}
	root := gjson.ParseBytes(data)
	if root.IsArray() {
		root = root.Get("0")
	}
	for _, tag := range []string{"ContentIdentifier", "MediaGroupUUID"} {
		value := root.Get(tag).String()
		if value == "" {
			root.ForEach(func(key, val gjson.Result) bool {
				if strings.HasSuffix(key.String(), ":"+tag) {
					value = val.String()
					return false
				}
				return true
			})
		}
		if value != "" && len(value) <= 128 {
			return value
		}
	}
	return ""
}
