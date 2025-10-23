package config

import (
	"sync"
)

var (
	once            sync.Once
	initThumbsMutex sync.Mutex
	LowMem          = false
	TotalMem        uint64
)
