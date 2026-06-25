package system

import (
	"io/fs"
	"syscall"
)

// statUID extrahiert die Besitzer-UID aus FileInfo (Linux-spezifisch).
func statUID(fi fs.FileInfo) (uint32, bool) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return st.Uid, true
}
