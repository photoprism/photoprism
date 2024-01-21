package config

import (
	"sync"
)

var once sync.Once
var LowMem = false
var TotalMem uint64
