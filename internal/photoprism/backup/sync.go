package backup

import (
	"sync"
	"time"
)

var (
	backupDatabaseMutex = sync.Mutex{}
	backupAlbumsMutex   = sync.Mutex{}
	backupAlbumsTime    = time.Time{}
)
