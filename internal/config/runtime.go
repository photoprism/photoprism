package config

import (
	"fmt"
	"runtime"

	"github.com/dustin/go-humanize"
	"github.com/pbnjay/memory"
)

// RuntimeInfo represents memory and cpu usage statistics.
type RuntimeInfo struct {
	Cores    int `json:"cores"`
	Routines int `json:"routines"`
	Memory   struct {
		Total    uint64 `json:"total"`
		Free     uint64 `json:"free"`
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
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	r.Cores = runtime.NumCPU()
	r.Routines = runtime.NumGoroutine()
	r.Memory.Total = memory.TotalMemory()
	r.Memory.Free = memory.FreeMemory()
	r.Memory.Used = mem.Alloc
	r.Memory.Reserved = mem.Sys
	r.Memory.Info = fmt.Sprintf("Used %s / Reserved %s", humanize.Bytes(r.Memory.Used), humanize.Bytes(r.Memory.Reserved))
}
