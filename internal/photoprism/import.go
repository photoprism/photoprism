package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/karrick/godirwalk"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
)

// Import represents an importer that can copy/move MediaFiles to the originals directory.
type Import struct {
	conf     *config.Config
	index    *Index
	convert  *Convert
	AllowExt fs.ExtList
}

// NewImport returns a new importer and expects its dependencies as arguments.
func NewImport(conf *config.Config, index *Index, convert *Convert) *Import {
	instance := &Import{
		conf:     conf,
		index:    index,
		convert:  convert,
		AllowExt: conf.ImportAllow(),
	}

	return instance
}

// originalsPath returns the original media files path as string.
func (imp *Import) originalsPath() string {
	return imp.conf.OriginalsPath()
}

// thumbPath returns the thumbnails cache path as string.
func (imp *Import) thumbPath() string {
	return imp.conf.ThumbCachePath()
}

// insufficientStorage reports whether the import must abort due to quota or low free disk space.
func (imp *Import) insufficientStorage() bool {
	if !imp.conf.InsufficientStorage() {
		return false
	}

	log.Errorf("import: aborting due to insufficient storage")
	event.ErrorMsg(i18n.ErrInsufficientStorage)
	return true
}

// ErrImportBusy is returned by Import.Run when the import has to wait for another task to complete.
var ErrImportBusy = errors.New("waiting for another task to complete")

// ErrImportDest is returned by Import.Run when the destination folder or the Import.Dest pattern gives names that
// are not supported.
var ErrImportDest = errors.New("destination is not supported")

// ErrImportPath is returned by Import.Run when the import folder is the originals folder or a folder inside it.
var ErrImportPath = errors.New("import and originals folders must not overlap")

// importError is an error with its own message that matches a sentinel error such as ErrImportDest.
type importError struct {
	msg string
	err error
}

// Error returns the error message.
func (e importError) Error() string {
	return e.msg
}

// Unwrap returns the sentinel error.
func (e importError) Unwrap() error {
	return e.err
}

// ErrImportIncomplete is returned by Import.Run when files could not be moved or copied to the originals.
var ErrImportIncomplete = errors.New("some files could not be imported")

// Start imports media files from a directory and converts/indexes them as needed.
func (imp *Import) Start(opt ImportOptions) fs.Done {
	done, _ := imp.Run(opt)
	return done
}

// Run imports media files like Start and returns an error if the import did not run, refused files
// early, or could not move or copy some of them, such as ErrImportBusy, status.ErrInsufficientStorage,
// status.ErrCanceled, or ErrImportIncomplete (status.ErrInsufficientStorage if the storage was full).
func (imp *Import) Run(opt ImportOptions) (done fs.Done, result error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("import: %s (panic)\nstack: %s", r, debug.Stack())
			result = fmt.Errorf("import: %s", r)
		}
	}()

	var directories []string
	done = make(fs.Done)

	if imp.conf == nil {
		log.Errorf("import: config is not set")
		return done, errors.New("config is not set")
	}

	// Importing indexes what it moves, so it reaches the same marker writes indexing does.
	if held := imp.conf.FacesLocked(); held != "" {
		log.Infof("import: waiting for the %s to complete", held)
		return done, ErrImportBusy
	}

	// The same applies once it has completed: its target is recorded rather than loaded here.
	imp.conf.CheckFaceModelSuperseded()

	importPath := opt.Path

	// Folder records name the configured import namespace, so a walk of that path or of a
	// subfolder is recorded relative to it. A walk rooted elsewhere, such as a per-session upload
	// directory, has no place in that namespace and records none. Names are checked below the
	// configured import folder, or below the folder the walk starts in if it is elsewhere.
	folderBase, nameRoot := "", filepath.Clean(importPath)

	if base := imp.conf.ImportPath(); base != "" {
		if _, ok := relBelow(base, importPath); ok {
			folderBase, nameRoot = base, base
		}
	}

	// Check if the import folder exists.
	if !fs.PathExists(importPath) {
		event.Error(fmt.Sprintf("import: directory %s not found", clean.Log(displayFolder(importPath, folderBase))))
		return done, errors.New("directory not found")
	}

	if err := imp.Check(opt); err != nil {
		return done, err
	}

	// Reset the cached disk usage so a freshly freed disk is detected immediately.
	disk.FlushFree()

	if imp.insufficientStorage() {
		return done, status.ErrInsufficientStorage
	}

	// Make sure to run import only once, unless otherwise requested.
	if !opt.NonBlocking {
		if err := mutex.IndexWorker.Start(); err != nil {
			event.Warn(fmt.Sprintf("import: %s", err.Error()))
			return done, ErrImportBusy
		}

		defer mutex.IndexWorker.Stop()
	}

	// The import workers check related files at their destination in originals as the index does.
	defer flushUnportable("index")
	defer flushUnportable("import")
	defer flushHidden("import")

	jobs := make(chan ImportJob)

	// Start a fixed number of goroutines to import files.
	var wg sync.WaitGroup
	var numWorkers = imp.conf.IndexWorkers()
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			ImportWorker(jobs)
			wg.Done()
		}()
	}

	filesImported := 0

	settings := imp.conf.Settings()
	convert := settings.Index.Convert && imp.conf.SidecarWritable()
	indexOpt := NewIndexOptions("/", true, convert, true, false, false, imp.conf)
	indexOpt.UID = opt.UID
	indexOpt.Action = opt.Action
	skipRaw := imp.conf.DisableRaw()
	ignore := fs.NewIgnoreList(fs.PPIgnoreFilename, true, false)

	if err := ignore.Path(importPath); err != nil {
		log.Infof("import: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof(`import: ignored "%s"`, clean.Log(fs.RelName(fileName, importPath)))
	}

	var stopped error
	var failures ImportFailures

	err := godirwalk.Walk(importPath, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("import: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			// The error callback skips nodes on any error, so the first reason a node is refused is kept here.
			if mutex.IndexWorker.Canceled() {
				if stopped == nil {
					stopped = status.ErrCanceled
				}
				return status.ErrCanceled
			}

			// Refuse further nodes if storage drops below the threshold mid-import.
			if imp.insufficientStorage() {
				imp.Cancel()
				if stopped == nil {
					stopped = status.ErrInsufficientStorage
				}
				return status.ErrInsufficientStorage
			}

			isDir, _ := info.IsDirOrSymlinkToDir()
			isSymlink := info.IsSymlink()

			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
				if !isDir || errors.Is(result, filepath.SkipDir) {
					return result
				}

				// A folder whose name is not portable stays in the import folder with its contents. It is not recorded as
				// walked, so a folder reached under another name is still imported under that name.
				if p := skipUnportableDir(done, fileName, fs.RelName(fileName, nameRoot), isSymlink); p != "" {
					warnUnportable("import", nameRoot, p, "import it")
					return filepath.SkipDir
				}

				// A linked folder that leads to originals is not walked, since its files are already in the library.
				if isSymlink && imp.linksToOriginals(fileName) {
					log.Warnf("import: skipped %s because it links to the originals folder", clean.Log(fs.RelName(fileName, nameRoot)))
					return filepath.SkipDir
				}

				if fileName != importPath {
					directories = append(directories, fileName)
				}

				if folderBase != "" {
					folder := entity.NewFolder(entity.RootImport, fs.RelName(fileName, folderBase), fs.ModTime(fileName))

					if _, created, err := entity.FirstOrCreateFolder(&folder); err == nil && created && folder.Path != "" {
						log.Infof("import: added folder /%s", clean.Log(folder.Path))
					}
				}

				return result
			}

			// A file whose name is not portable stays in the import folder.
			if !clean.PortableName(filepath.Base(fileName)) {
				warnUnportable("import", nameRoot, fs.RelName(fileName, nameRoot), "import it")
				return nil
			}

			done[fileName] = fs.Found

			if !media.MainFile(fileName) {
				return nil
			}

			mf, err := NewMediaFile(fileName)

			// Check if file exists and is not empty.
			if err != nil {
				log.Warnf("import: %s", clean.Error(err))
				return nil
			} else if mf.Empty() {
				return nil
			}

			// Ignore RAW images?
			if mf.IsRaw() && skipRaw {
				log.Infof("import: skipped raw %s", clean.Log(mf.RootRelName()))
				return nil
			}

			// Report files that have a missing or invalid filename extension,
			// see https://github.com/photoprism/photoprism/issues/3518.
			typeErr := mf.CheckType()

			if typeErr != nil {
				if !opt.RemoveInvalidFiles {
					log.Warnf("import: %s %s and will not be indexed", clean.Log(mf.RootRelName()), typeErr)
				} else if removeErr := mf.Remove(); removeErr != nil {
					log.Errorf("import: %s %s and %s", clean.Log(mf.RootRelName()), typeErr, removeErr)
					return nil
				} else {
					log.Warnf("import: %s %s and was deleted", clean.Log(mf.RootRelName()), typeErr)
					return nil
				}
			}

			// Create JSON sidecar file, if needed. A file that failed the type check is still moved
			// with its stack, but not read by ExifTool.
			if typeErr == nil {
				if err = mf.CreateExifToolJson(imp.convert); err != nil {
					log.Warnf("import: %s", clean.Error(err))
				}
			}

			// Find related files to import.
			related, err := mf.RelatedFiles(imp.conf.Settings().StackSequences())

			if err != nil {
				event.Error(fmt.Sprintf("import: %s", err))
				return nil
			}

			// Related files whose names are not portable stay in the import folder.
			related = PortableRelated(related, mf, func(f *MediaFile) {
				warnUnportable("import", nameRoot, fs.RelName(f.FileName(), nameRoot), "import it")
			})

			if related.Main == nil {
				return nil
			}

			var files MediaFiles

			for _, f := range related.Files {
				if f.FileSize() == 0 || done[f.FileName()].Processed() {
					continue
				}

				files = append(files, f)
				filesImported++
				done[f.FileName()] = fs.Processed
			}

			done[fileName] = fs.Processed

			related.Files = files

			jobs <- ImportJob{
				FileName:  fileName,
				Related:   related,
				IndexOpt:  indexOpt,
				ImportOpt: opt,
				Imp:       imp,
				Failures:  &failures,
			}

			return nil
		},
		Unsorted:            false,
		FollowSymbolicLinks: true,
	})

	walkErr := err

	if walkErr == nil {
		walkErr = stopped
	}

	close(jobs)
	wg.Wait()

	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	if opt.RemoveEmptyDirectories {
		// Remove empty directories from import path.
		for _, directory := range directories {
			if fs.DirIsEmpty(directory) {
				if removeErr := os.Remove(directory); removeErr != nil {
					log.Errorf("import: failed to delete empty folder %s (%s)", clean.Log(fs.RelName(directory, importPath)), clean.Error(removeErr))
				} else {
					log.Infof("import: deleted empty folder %s", clean.Log(fs.RelName(directory, importPath)))
				}
			}
		}
	}

	if opt.RemoveHiddenFiles {
		// Hidden files are removed, except files of a known type whose names start with another character than a
		// dot, such as pictures and their sidecars, which are kept and reported if they are media files.
		for _, file := range ignore.Hidden() {
			if !fs.FileExists(file) {
				continue
			}

			if base := filepath.Base(file); base == ".keep" || base == ".gitkeep" {
				log.Debugf("import: %s file is preserved", clean.Log(fs.RelName(file, importPath)))
			} else if !strings.HasPrefix(base, ".") && fs.FileType(file) != fs.TypeUnknown {
				if media.MainFile(file) {
					warnHidden("import", nameRoot, fs.RelName(file, nameRoot), "import it")
				}
			} else if err = os.Remove(file); err != nil {
				log.Errorf("import: failed removing %s (%s)", clean.Log(fs.RelName(file, importPath)), clean.Error(err))
			}
		}
	}

	logWalkResult("import", walkErr)

	if filesImported > 0 {
		// Run face recognition if enabled.
		if w := NewFaces(imp.conf); w.Disabled() {
			log.Debugf("import: skipping face recognition")
		} else if err = w.Start(FacesOptionsDefault()); err != nil {
			log.Errorf("import: %s", err)
		}

		// Update photo counts and visibilities.
		if err = entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s (update counts)", err)
		}
	}

	config.FlushUsageCache()
	runtime.GC()

	// A stop reason takes precedence, since it applies to the files that were not attempted too.
	if walkErr == nil {
		walkErr = failures.Err()
	}

	return done, walkErr
}

// Cancel stops the current import operation.
func (imp *Import) Cancel() {
	mutex.IndexWorker.Cancel()
}

// storedMatch describes how a file relates to the indexed file that holds the same content.
type storedMatch int

const (
	storedNone      storedMatch = iota // no indexed file holds the content
	storedCopy                         // another file holds the content
	storedSame                         // the file itself, e.g. a link or another name for it
	storedMaybeSame                    // possibly the file itself, reached through another mount
)

// storedFileOf reports how a file relates to the indexed file that holds the same content, and returns that file if
// it is another file. The answer comes from the files rather than from the row, since a caller removes its source
// on the strength of it.
func storedFileOf(mediaFile *MediaFile) (*entity.File, storedMatch) {
	if mediaFile == nil {
		return nil, storedNone
	}

	fileHash := mediaFile.Hash()

	if fileHash == "" {
		return nil, storedNone
	}

	found, err := entity.FirstFileByHash(fileHash)

	if err != nil {
		return nil, storedNone
	}

	// A name that holds nothing hashes to the empty string, so this covers an absent file too.
	storedName := FileName(found.FileRoot, found.FileName)

	if storedHash(storedName) != fileHash {
		return nil, storedNone
	}

	src, srcErr := os.Stat(mediaFile.FileName())

	if srcErr != nil {
		return nil, storedNone
	}

	dest, destErr := os.Stat(storedName)

	switch {
	case destErr != nil:
		return nil, storedNone
	case os.SameFile(src, dest):
		return nil, storedSame
	case maybeSameFile(src, dest):
		return nil, storedMaybeSame
	default:
		return &found, storedCopy
	}
}

// storedHash returns the hash of an indexed file, which reads the whole file.
var storedHash = fs.Hash

// maybeSameFile reports whether two files may be the same file on another mount, and can be replaced in tests.
var maybeSameFile = fs.MaybeSameFile

// Check returns ErrImportPath or ErrImportDest if the import folder or the destination of the options is not
// supported, see CheckPath and CheckDest. The details name folders, so only the system log has them.
func (imp *Import) Check(opt ImportOptions) error {
	// Files are moved or copied to originals, so the import folder must not overlap with it.
	if err := imp.CheckPath(opt.Path); err != nil {
		event.SystemError([]string{"import", "%s"}, err.Error())
		event.Error("import: import and originals folders must not overlap")
		return ErrImportPath
	}

	// Imported files must get names that the index accepts.
	if err := imp.CheckDest(opt.DestFolder); err != nil {
		event.SystemError([]string{"import", "%s"}, err.Error())
		event.Error("import: destination is not supported")
		return ErrImportDest
	}

	return nil
}

// CheckPath returns ErrImportPath if the import folder is the originals folder, a folder inside it, or a folder
// that contains it, compared after resolving symbolic links. Walks skip hidden folders and the storage folder, so a
// folder below one, such as originals/.photoprism/storage or originals/storage, may contain or be in the other.
func (imp *Import) CheckPath(importPath string) error {
	originals := imp.originalsPath()

	if importPath == "" || originals == "" {
		return nil
	}

	src, dest := resolvedPath(importPath), resolvedPath(originals)

	// walked reports whether folder is dir or a folder below it that a walk of dir reaches.
	walked := func(dir, folder string) bool {
		rel, ok := relBelow(dir, folder)
		return ok && !walkSkips(dir, rel)
	}

	if walked(dest, src) || walked(src, dest) {
		return fmt.Errorf("%w (import %s, originals %s)", ErrImportPath, clean.Log(importPath), clean.Log(originals))
	}

	return nil
}

// linksToOriginals reports whether a linked folder leads to the originals folder, a folder in it, or a folder
// that contains it.
func (imp *Import) linksToOriginals(dir string) bool {
	return overlappingPaths(dir, imp.originalsPath())
}

// overlappingPaths reports whether two folders are the same, or one is inside the other, after resolving symbolic
// links. An empty path overlaps with nothing.
func overlappingPaths(a, b string) bool {
	if a == "" || b == "" {
		return false
	}

	a, b = resolvedPath(a), resolvedPath(b)

	if _, ok := relBelow(b, a); ok {
		return true
	}

	_, ok := relBelow(a, b)

	return ok
}

// relBelow returns the path of folder relative to dir, and whether folder is dir itself or a folder below it.
func relBelow(dir, folder string) (string, bool) {
	rel, err := filepath.Rel(dir, folder)

	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}

	return rel, true
}

// walkSkips reports whether a walk of dir skips the folder at a relative path, because a folder on the way has a
// hidden name or contains the storage marker file, as fs.SkipWalk checks.
func walkSkips(dir, rel string) bool {
	if rel == "." {
		return false
	}

	for _, name := range strings.Split(rel, string(filepath.Separator)) {
		if dir = filepath.Join(dir, name); fs.FileNameHidden(name) || fs.FileExists(filepath.Join(dir, fs.PPStorageFilename)) {
			return true
		}
	}

	return false
}

// resolvedPath returns the absolute path of a folder with symbolic links resolved, or cleaned if it cannot be resolved.
func resolvedPath(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}

	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}

	return filepath.Clean(dir)
}

// CheckDest returns ErrImportDest if the destination folder or the Import.Dest pattern gives names that are not
// portable or hidden, which the index would skip.
func (imp *Import) CheckDest(folder string) error {
	// Checked as cleaned, since the destination is joined below the originals folder; ".." is refused as hidden.
	if rel := path.Clean(clean.SlashPath(folder)); rel != "." && clean.UnsupportedPath(rel) != "" {
		return importError{fmt.Sprintf("destination folder %s is not supported", clean.Log(folder)), ErrImportDest}
	}

	// Literal characters of the pattern appear in every name, so one sample time shows them.
	pathPattern, namePattern := imp.conf.Settings().Import.GetDestName()
	sample := time.Date(2026, time.October, 15, 13, 4, 5, 0, time.UTC)

	if clean.UnsupportedPath(path.Join(sample.Format(pathPattern), fs.CanonicalName(sample, "", namePattern)+".jpg")) != "" {
		return importError{fmt.Sprintf("destination pattern %s is not supported", clean.Log(imp.conf.Settings().Import.GetDest())), ErrImportDest}
	}

	return nil
}

// DestinationFilename returns the destination filename of a MediaFile to be imported.
// Format: 2006/01/20060102_150405_CHECKSUM.ext
func (imp *Import) DestinationFilename(mainFile *MediaFile, mediaFile *MediaFile, folder string) (string, error) {
	// Get the import destination path and file name patterns.
	pathPattern, namePattern := imp.conf.Settings().Import.GetDestName()

	fileName := mainFile.CanonicalName(namePattern)
	fileExtension := mediaFile.Extension()
	dateCreated := mainFile.DateCreated()

	if !mediaFile.IsSidecar() {
		if f, err := entity.FirstFileByHash(mediaFile.Hash()); err == nil {
			existingFilename := FileName(f.FileRoot, f.FileName)

			// The recorded name has three states, not two: the file is there and this is a duplicate,
			// the name is free and the file can be restored to it, or an entry holds the name without
			// the file being there, which is neither and takes the search below, as does a recorded
			// name that is not portable.
			if fs.FileExists(existingFilename) {
				// The index records this hash under that name, which is not the same as the file
				// still holding it, so the message says what was matched rather than claiming the
				// two files are identical. storedFileOf is what establishes that, for the caller
				// that removes its source.
				return existingFilename, fmt.Errorf("%s is already indexed as %s", clean.Log(filepath.Base(mediaFile.FileName())), clean.Log(f.FileName))
			} else if !fs.IsSymlink(existingFilename) && clean.UnportablePath(f.FileName) == "" {
				return existingFilename, nil
			}
		}
	}

	// Find and return the next available file name if the default name is already being used by another
	// file. A symbolic link holds the name whether or not it resolves, so the search steps over one.
	i := 0
	pathName := filepath.Join(imp.originalsPath(), folder, dateCreated.Format(pathPattern))
	filePath := filepath.Join(pathName, fileName+fileExtension)

	for fs.FileExists(filePath) || fs.IsSymlink(filePath) {
		// Both sides hash to the empty string when they cannot be read, so a hash is compared only
		// once there is one.
		if h := mediaFile.Hash(); h != "" && h == fs.Hash(filePath) {
			return filePath, fmt.Errorf("%s already exists", clean.Log(fs.RelName(filePath, imp.originalsPath())))
		}

		i++

		filePath = filepath.Join(pathName, fileName+"."+fmt.Sprintf("%05d", i)+fileExtension)
	}

	return filePath, nil
}
