// +build linux

package stats

import (
	"syscall"
)

func (mem *MemStatus) fillInStatus() {
	//system memory usage
	sysInfo := new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(sysInfo)
	if err == nil {
		mem.All = uint64(sysInfo.Totalram)
		mem.Free = uint64(sysInfo.Freeram)
		mem.Used = mem.All - mem.Free
	}
}
