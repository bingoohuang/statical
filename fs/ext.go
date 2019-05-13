package fs

import "os"

func (fs statikFS) Files() map[string]file {
	return fs.files
}

func (fs statikFS) Dirs() map[string][]string {
	return fs.dirs
}

func (r file) Info() os.FileInfo {
	return r.FileInfo
}

func (r file) Data() []byte {
	return r.data
}

func (r file) Fs() *statikFS {
	return r.fs
}
