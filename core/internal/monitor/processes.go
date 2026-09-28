package monitor

import (
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessStat is one process's share of the CPU since the previous sample,
// in percent of one core, and its resident memory in bytes.
type ProcessStat struct {
	PID        int32
	PPID       int32
	Name       string
	CPUPercent float64
	RSS        uint64
}

// Processes samples every process on the host.
//
// It keeps a handle per PID between samples so that CPU shares are measured
// over the sampling interval; a PID taken over by a new process gets a fresh
// handle. Only the standalone system monitor uses it.
type Processes struct {
	handles map[int32]*process.Process
}

func NewProcesses() *Processes {
	return &Processes{handles: make(map[int32]*process.Process)}
}

// Sample returns the processes whose CPU and memory the current user can
// read. A process seen for the first time reports no CPU share yet.
func (p *Processes) Sample() ([]ProcessStat, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	handles := make(map[int32]*process.Process, len(procs))
	stats := make([]ProcessStat, 0, len(procs))
	for _, proc := range procs {
		handle := proc
		if cached, ok := p.handles[proc.Pid]; ok {
			cachedStart, _ := cached.CreateTime()
			start, _ := proc.CreateTime()
			if cachedStart == start {
				handle = cached
			}
		}
		handles[proc.Pid] = handle

		cpuPercent, err := handle.Percent(0)
		if err != nil {
			continue
		}
		mem, err := handle.MemoryInfo()
		if err != nil || mem.RSS == 0 {
			continue
		}
		name, _ := handle.Name()
		ppid, _ := handle.Ppid()
		stats = append(stats, ProcessStat{
			PID:        proc.Pid,
			PPID:       ppid,
			Name:       name,
			CPUPercent: cpuPercent,
			RSS:        mem.RSS,
		})
	}
	p.handles = handles
	return stats, nil
}
