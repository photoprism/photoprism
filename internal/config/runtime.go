package config

import (
	"fmt"
	"runtime"

	"github.com/dustin/go-humanize"
)

// RuntimeInfo represents memory and cpu usage statistics.
type RuntimeInfo struct {
	Cores    int `json:"cores"`
	Routines int `json:"routines"`
	Memory   struct {
		Used     uint64 `json:"used"`
		Reserved uint64 `json:"reserved"`
		Info     string `json:"info"`
	} `json:"memory"`
}

// NewRuntimeInfo returns a new RuntimeInfo instance.
func NewRuntimeInfo() (r RuntimeInfo) {
	r = RuntimeInfo{}
	r.Refresh()

	return r
}

// Refresh updates runtime info options like number of goroutines and memory usage.
func (r *RuntimeInfo) Refresh() {
	r.Cores = runtime.NumCPU()
	r.Routines = runtime.NumGoroutine()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	r.Memory.Used = mem.Alloc
	r.Memory.Reserved = mem.Sys
	r.Memory.Info = fmt.Sprintf("Used %s / Reserved %s", humanize.Bytes(r.Memory.Used), humanize.Bytes(r.Memory.Reserved))
}
