package ytdl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Info youtube-dl info
type Info struct {
	// Generated from youtube-dl README using:
	// sed -e 's/ - `\(.*\)` (\(.*\)): \(.*\)/\1 \2 `json:"\1"` \/\/ \3/' | sed -e 's/numeric/float64/' | sed -e 's/boolean/bool/' | sed -e 's/_id/ID/'  | sed -e 's/_count/Count/'| sed -e 's/_uploader/Uploader/' | sed -e 's/_key/Key/' | sed -e 's/_year/Year/' | sed -e 's/_title/Title/' | sed -e 's/_rating/Rating/'  | sed -e 's/_number/Number/'  | awk '{print toupper(substr($0, 0, 1))  substr($0, 2)}'
	ID                 string  `json:"id"`                   // Video identifier
	Title              string  `json:"title"`                // Video title
	URL                string  `json:"url"`                  // Video URL
	AltTitle           string  `json:"alt_title"`            // A secondary title of the video
	DisplayID          string  `json:"display_id"`           // An alternative identifier for the video
	Uploader           string  `json:"uploader"`             // Full name of the video uploader
	License            string  `json:"license"`              // License name the video is licensed under
	Creator            string  `json:"creator"`              // The creator of the video
	ReleaseDate        string  `json:"release_date"`         // The date (YYYYMMDD) when the video was released
	Timestamp          float64 `json:"timestamp"`            // UNIX timestamp of the moment the video became available
	UploadDate         string  `json:"upload_date"`          // Video upload date (YYYYMMDD)
	UploaderID         string  `json:"uploader_id"`          // Nickname or id of the video uploader
	Channel            string  `json:"channel"`              // Full name of the channel the video is uploaded on
	ChannelID          string  `json:"channel_id"`           // Id of the channel
	Location           string  `json:"location"`             // Physical location where the video was filmed
	Duration           float64 `json:"duration"`             // Length of the video in seconds
	ViewCount          float64 `json:"view_count"`           // How many users have watched the video on the platform
	LikeCount          float64 `json:"like_count"`           // Number of positive ratings of the video
	DislikeCount       float64 `json:"dislike_count"`        // Number of negative ratings of the video
	RepostCount        float64 `json:"repost_count"`         // Number of reposts of the video
	AverageRating      float64 `json:"average_rating"`       // Average rating give by users, the scale used depends on the webpage
	CommentCount       float64 `json:"comment_count"`        // Number of comments on the video
	AgeLimit           float64 `json:"age_limit"`            // Age restriction for the video (years)
	IsLive             bool    `json:"is_live"`              // Whether this video is a live stream or a fixed-length video
	StartTime          float64 `json:"start_time"`           // Time in seconds where the reproduction should start, as specified in the URL
	EndTime            float64 `json:"end_time"`             // Time in seconds where the reproduction should end, as specified in the URL
	Extractor          string  `json:"extractor"`            // Name of the extractor
	ExtractorKey       string  `json:"extractor_key"`        // Key name of the extractor
	Epoch              float64 `json:"epoch"`                // Unix epoch when creating the file
	Autonumber         float64 `json:"autonumber"`           // Five-digit number that will be increased with each download, starting at zero
	Playlist           string  `json:"playlist"`             // Name or id of the playlist that contains the video
	PlaylistIndex      float64 `json:"playlist_index"`       // Index of the video in the playlist padded with leading zeros according to the total length of the playlist
	PlaylistID         string  `json:"playlist_id"`          // Playlist identifier
	PlaylistTitle      string  `json:"playlist_title"`       // Playlist title
	PlaylistUploader   string  `json:"playlist_uploader"`    // Full name of the playlist uploader
	PlaylistUploaderID string  `json:"playlist_uploader_id"` // Nickname or id of the playlist uploader

	// Available for the video that belongs to some logical chapter or section:
	Chapter       string  `json:"chapter"`        // Name or title of the chapter the video belongs to
	ChapterNumber float64 `json:"chapter_number"` // Number of the chapter the video belongs to
	ChapterID     string  `json:"chapter_id"`     // Id of the chapter the video belongs to

	// Available for the video that is an episode of some series or programme:
	Series        string  `json:"series"`         // Title of the series or programme the video episode belongs to
	Season        string  `json:"season"`         // Title of the season the video episode belongs to
	SeasonNumber  float64 `json:"season_number"`  // Number of the season the video episode belongs to
	SeasonID      string  `json:"season_id"`      // Id of the season the video episode belongs to
	Episode       string  `json:"episode"`        // Title of the video episode
	EpisodeNumber float64 `json:"episode_number"` // Number of the video episode within a season
	EpisodeID     string  `json:"episode_id"`     // Id of the video episode

	// Available for the media that is a track or a part of a music album:
	Track       string  `json:"track"`        // Title of the track
	TrackNumber float64 `json:"track_number"` // Number of the track within an album or a disc
	TrackID     string  `json:"track_id"`     // Id of the track
	Artist      string  `json:"artist"`       // Artist(s) of the track
	Genre       string  `json:"genre"`        // Genre(s) of the track
	Album       string  `json:"album"`        // Title of the album the track belongs to
	AlbumType   string  `json:"album_type"`   // Type of the album
	AlbumArtist string  `json:"album_artist"` // List of all artists appeared on the album
	DiscNumber  float64 `json:"disc_number"`  // Number of the disc or other physical medium the track belongs to
	ReleaseYear float64 `json:"release_year"` // Year (YYYY) when the album was released

	Type        string `json:"_type"`
	Direct      bool   `json:"direct"`
	WebpageURL  string `json:"webpage_url"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	// don't unmarshal, populated from image thumbnail file
	ThumbnailBytes []byte      `json:"-"`
	Thumbnails     []Thumbnail `json:"thumbnails"`

	Formats   []Format              `json:"formats"`
	Subtitles map[string][]Subtitle `json:"subtitles"`

	// Playlist entries if _type is playlist
	Entries []Info `json:"entries"`

	// Info can also be a mix of Info and one Format
	Format
}

func infoFromURL(
	ctx context.Context,
	rawURL string,
	options Options,
) (info Info, rawJSON []byte, err error) {
	cmd := exec.CommandContext(
		ctx,
		FindBin(),
		// see comment below about ignoring errors for playlists
		"--ignore-errors",
		// TODO: deprecated in yt-dlp?
		"--no-call-home",
		// use safer output filenmaes
		// TODO: needed?
		"--restrict-filenames",
		// use .netrc authentication data
		"--netrc",
		// provide url via stdin for security, youtube-dl has some run command args
		"--batch-file", "-",
		// dump info json
		"--dump-single-json",
	)

	if options.ProxyUrl != "" {
		cmd.Args = append(cmd.Args, "--proxy", options.ProxyUrl)
	}

	if options.UseIPV4 {
		cmd.Args = append(cmd.Args, "-4")
	}

	if options.Downloader != "" {
		cmd.Args = append(cmd.Args, "--downloader", options.Downloader)
	}

	if options.Impersonate != "" {
		cmd.Args = append(cmd.Args, "--impersonate", options.Impersonate)
	}

	if options.Cookies != "" {
		cmd.Args = append(cmd.Args, "--cookies", options.Cookies)
	}

	if options.CookiesFromBrowser != "" {
		cmd.Args = append(cmd.Args, "--cookies-from-browser", options.CookiesFromBrowser)
	}

	switch options.Type {
	case TypePlaylist, TypeChannel:
		cmd.Args = append(cmd.Args, "--yes-playlist")

		if options.PlaylistStart > 0 {
			cmd.Args = append(cmd.Args,
				"--playlist-start", strconv.Itoa(int(options.PlaylistStart)),
			)
		}
		if options.PlaylistEnd > 0 {
			cmd.Args = append(cmd.Args,
				"--playlist-end", strconv.Itoa(int(options.PlaylistEnd)),
			)
		}
		if options.FlatPlaylist {
			cmd.Args = append(cmd.Args, "--flat-playlist")
		}
	case TypeSingle:
		if options.DownloadSubtitles {
			cmd.Args = append(cmd.Args,
				"--all-subs",
			)
		}
		cmd.Args = append(cmd.Args,
			"--no-playlist",
		)
	case TypeAny:
		break
	default:
		return Info{}, nil, fmt.Errorf("unhandled options type value: %d", options.Type)
	}

	tempPath, _ := os.MkdirTemp("", "ydls")
	defer os.RemoveAll(tempPath)

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	stderrWriter := io.Discard
	if options.StderrFn != nil {
		stderrWriter = options.StderrFn(cmd)
	}

	cmd.Dir = tempPath
	cmd.Stdout = stdoutBuf
	cmd.Stderr = io.MultiWriter(stderrBuf, stderrWriter)
	cmd.Stdin = bytes.NewBufferString(rawURL + "\n")

	log.Trace("cmd", " ", cmd.Args)
	cmdErr := cmd.Run()

	stderrLineScanner := bufio.NewScanner(stderrBuf)
	errMessage := ""
	for stderrLineScanner.Scan() {
		const errorPrefix = "ERROR: "
		line := stderrLineScanner.Text()
		if strings.HasPrefix(line, errorPrefix) {
			errMessage = line[len(errorPrefix):]
		}
	}

	infoSeemsOk := false
	if len(stdoutBuf.Bytes()) > 0 {
		if infoErr := json.Unmarshal(stdoutBuf.Bytes(), &info); infoErr != nil {
			return Info{}, nil, infoErr
		}

		isPlaylist := info.Type == "playlist" || info.Type == "multi_video"
		switch {
		case options.Type == TypePlaylist && !isPlaylist:
			return Info{}, nil, ErrNotAPlaylist
		case options.Type == TypeSingle && isPlaylist:
			return Info{}, nil, ErrNotASingleEntry
		default:
			// any type
		}

		// HACK: --ignore-errors still return error message and exit code != 0
		// so workaround is to assume things went ok if we get some ok json on stdout
		infoSeemsOk = info.ID != ""
	}

	if !infoSeemsOk {
		if errMessage != "" {
			return Info{}, nil, YoutubedlError(errMessage)
		} else if cmdErr != nil {
			return Info{}, nil, cmdErr
		}

		return Info{}, nil, fmt.Errorf("unknown error")
	}

	get := func(url string) (*http.Response, error) {
		c := http.DefaultClient
		if options.HttpClient != nil {
			c = options.HttpClient
		}

		r, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		for k, v := range info.HTTPHeaders {
			r.Header.Set(k, v)
		}
		return c.Do(r)
	}

	if options.DownloadThumbnail && info.Thumbnail != "" {
		resp, respErr := get(info.Thumbnail)
		if respErr == nil {
			buf, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			info.ThumbnailBytes = buf
		}
	}

	for language, subtitles := range info.Subtitles {
		for i := range subtitles {
			subtitles[i].Language = language
		}
	}

	if options.DownloadSubtitles {
		for _, subtitles := range info.Subtitles {
			for i, subtitle := range subtitles {
				resp, respErr := get(subtitle.URL)
				if respErr == nil {
					buf, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					subtitles[i].Bytes = buf
				}
			}
		}
	}

	// as we ignore errors for playlists some entries might show up as null
	//
	// note: instead of doing full recursion, we assume entries in
	// playlists and channels are at most 2 levels deep, and we just
	// collect entries from both levels.
	//
	// the following cases have not been tested:
	//
	// - entries that are more than 2 levels deep (will be missed)
	// - the ability to restrict entries to a single level (we include both levels)
	if options.Type == TypePlaylist || options.Type == TypeChannel {
		var filteredEntries []Info
		for _, e := range info.Entries {
			if e.Type == "playlist" {
				for _, ee := range e.Entries {
					if ee.ID == "" {
						continue
					}
					filteredEntries = append(filteredEntries, ee)
				}
				continue
			} else if e.ID != "" {
				filteredEntries = append(filteredEntries, e)
			}
		}
		info.Entries = filteredEntries
	}

	return info, stdoutBuf.Bytes(), nil
}
