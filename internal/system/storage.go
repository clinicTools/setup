package system

import (
	"encoding/json"
	"strings"
	"syscall"
)

// Filesystem beschreibt einen gemounteten Datenträger.
type Filesystem struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Type       string  `json:"type"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	UsePercent float64 `json:"usePercent"`
}

// BlockDevice spiegelt die lsblk-Struktur (Datenträger + Partitionen) wider.
type BlockDevice struct {
	Name       string        `json:"name"`
	Size       string        `json:"size"`
	Type       string        `json:"type"`
	Mountpoint string        `json:"mountpoint,omitempty"`
	FSType     string        `json:"fstype,omitempty"`
	Model      string        `json:"model,omitempty"`
	Children   []BlockDevice `json:"children,omitempty"`
}

// StorageInfo bündelt Dateisysteme und Blockgeräte.
type StorageInfo struct {
	Filesystems []Filesystem  `json:"filesystems"`
	Devices     []BlockDevice `json:"devices"`
}

// Storage sammelt Dateisysteme aus /proc/mounts (+ statfs) und Blockgeräte via lsblk.
func Storage() (*StorageInfo, error) {
	info := &StorageInfo{
		Filesystems: filesystems(),
		Devices:     blockDevices(),
	}
	return info, nil
}

// realFSTypes filtert virtuelle Dateisysteme heraus.
var virtualFS = map[string]struct{}{
	"proc": {}, "sysfs": {}, "devtmpfs": {}, "devpts": {}, "tmpfs": {},
	"cgroup": {}, "cgroup2": {}, "pstore": {}, "securityfs": {}, "debugfs": {},
	"mqueue": {}, "hugetlbfs": {}, "fusectl": {}, "configfs": {}, "bpf": {},
	"tracefs": {}, "ramfs": {}, "autofs": {}, "binfmt_misc": {}, "overlay": {},
	"nsfs": {}, "squashfs": {},
}

func filesystems() []Filesystem {
	data, err := readFile("/proc/mounts")
	if err != nil {
		return nil
	}
	var out []Filesystem
	seen := map[string]struct{}{}
	for _, l := range lines(data) {
		fields := strings.Fields(l)
		if len(fields) < 3 {
			continue
		}
		device, mount, fstype := fields[0], fields[1], fields[2]
		if _, virtual := virtualFS[fstype]; virtual {
			continue
		}
		if _, ok := seen[mount]; ok {
			continue
		}
		seen[mount] = struct{}{}

		var st syscall.Statfs_t
		if err := syscall.Statfs(mount, &st); err != nil {
			continue
		}
		bsize := uint64(st.Bsize)
		total := st.Blocks * bsize
		free := st.Bavail * bsize
		used := total - st.Bfree*bsize
		fs := Filesystem{
			Device: device, Mountpoint: mount, Type: fstype,
			Total: total, Free: free, Used: used,
		}
		if total > 0 {
			fs.UsePercent = float64(used) / float64(total) * 100
		}
		out = append(out, fs)
	}
	return out
}

func blockDevices() []BlockDevice {
	if !commandExists("lsblk") {
		return nil
	}
	out, err := run("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL")
	if err != nil {
		return nil
	}
	var parsed struct {
		BlockDevices []BlockDevice `json:"blockdevices"`
	}
	if json.Unmarshal([]byte(out), &parsed) != nil {
		return nil
	}
	return parsed.BlockDevices
}
