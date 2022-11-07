package env

import (
	"fmt"
	"runtime"

	"github.com/dustin/go-humanize"
	"github.com/pbnjay/memory"
)

// Resources represents runtime resource information.
type Resources struct {
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

// Update runtime resource information.
func (r *Resources) Update() {
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
