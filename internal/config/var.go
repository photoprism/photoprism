package config

import (
	"sync"
)

var (
	once            sync.Once
	initThumbsMutex sync.Mutex
	// LowMem indicates the system has less RAM than the recommended minimum.
	LowMem = false
	// TotalMem stores the detected system memory in bytes.
	TotalMem uint64
)
