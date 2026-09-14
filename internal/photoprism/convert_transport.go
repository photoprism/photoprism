package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// transportConversions coordinates transport-stream output throughout the serving process.
var transportConversions transportGroup

// errTransportInterrupted reports an operation that ended without returning a result.
var errTransportInterrupted = errors.New("convert: transport conversion interrupted")

// transportRequest identifies callers that can share one conversion result.
type transportRequest struct {
	info    os.FileInfo
	source  string
	conf    *config.Config
	encoder encode.Encoder
	noMutex bool
	force   bool
}

// matches reports whether callers share both conversion settings and the observed source version.
func (r transportRequest) matches(other transportRequest) bool {
	if r.source != other.source || r.conf != other.conf || r.encoder != other.encoder || r.noMutex != other.noMutex || r.force != other.force {
		return false
	}
	if r.info == nil || other.info == nil {
		return r.info == nil && other.info == nil
	}
	return os.SameFile(r.info, other.info) && r.info.Size() == other.info.Size() && r.info.ModTime().Equal(other.info.ModTime())
}

// transportCall carries a completion shared by compatible callers.
type transportCall struct {
	request    transportRequest
	previous   *transportCall
	done       chan struct{}
	path       string
	err        error
	panicValue any
}

// transportGroup shares matching operations and queues other operations for the same destination.
type transportGroup struct {
	mutex sync.Mutex
	calls map[string]*transportCall
}

// start registers an operation before returning, retaining only the current tail for each destination.
func (g *transportGroup) start(key string, request transportRequest, run func() (string, error)) *transportCall {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	previous := g.calls[key]
	if previous != nil && previous.request.matches(request) {
		return previous
	}
	if g.calls == nil {
		g.calls = make(map[string]*transportCall)
	}
	call := &transportCall{request: request, previous: previous, done: make(chan struct{})}
	g.calls[key] = call
	go g.execute(key, call, run)
	return call
}

// execute publishes completion even when a callback panics or exits its goroutine.
func (g *transportGroup) execute(key string, call *transportCall, run func() (string, error)) {
	completed := false
	defer func() {
		if !completed {
			call.panicValue = recover()
			call.err = errTransportInterrupted
		}
		g.mutex.Lock()
		if g.calls[key] == call {
			delete(g.calls, key)
		}
		close(call.done)
		g.mutex.Unlock()
	}()
	if call.previous != nil {
		<-call.previous.done
		call.previous = nil
	}
	call.path, call.err = run()
	completed = true
}

// wait returns the operation's result or propagates its panic in the waiting caller.
func (call *transportCall) wait() (string, error) {
	<-call.done
	if call.panicValue != nil {
		panic(call.panicValue)
	}
	return call.path, call.err
}

// coordinatedTransport retains destination coordination through remuxing and any fallback transcode.
// Results carry paths rather than mutable MediaFile objects so each caller gets its own metadata view.
func (w *Convert) coordinatedTransport(f *MediaFile, encoder encode.Encoder, noMutex, force bool) (*MediaFile, error) {
	destination, err := fs.FileName(f.FileName(), w.conf.SidecarPath(), w.conf.OriginalsPath(), fs.ExtMp4)
	if err != nil {
		return nil, fmt.Errorf("convert: %w (remux)", err)
	}
	parent, err := fs.Resolve(filepath.Dir(destination))
	if err != nil {
		return nil, fmt.Errorf("convert: %w (remux)", err)
	}
	source, err := fs.Resolve(f.FileName())
	if err != nil {
		return nil, fmt.Errorf("convert: %w (remux)", err)
	}
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("convert: %w (remux)", err)
	}
	request := transportRequest{source: source, info: info, conf: w.conf, encoder: encoder, noMutex: noMutex, force: force}
	call := transportConversions.start(filepath.Join(parent, filepath.Base(destination)), request, func() (string, error) {
		file, runErr := w.toAvc(f, encoder, noMutex, force, false)
		if file == nil {
			return "", runErr
		}
		return file.FileName(), runErr
	})
	name, err := call.wait()
	if name == "" {
		return nil, err
	}
	file, readErr := NewMediaFile(name)
	if err != nil {
		return file, err
	}
	return file, readErr
}
