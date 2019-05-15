// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fs contains an HTTP file system that works with zip contents.
package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var ZipData string

// file holds unzipped read-only file contents and file metadata.
type File struct {
	os.FileInfo
	Data []byte
	Fs   *StatiqFS
}

type StatiqFS struct {
	Files map[string]File
	Dirs  map[string][]string
}

// Register registers zip contents data, later used to initialize
// the statiq file system.
func Register(data string) { ZipData = data }

// New creates a new file system with the registered zip contents data.
// It unzips all files and stores them in an in-memory map.
func New() (*StatiqFS, error) {
	if ZipData == "" {
		return nil, errors.New("statiq/fs: no zip data registered")
	}
	zipReader, err := zip.NewReader(strings.NewReader(ZipData), int64(len(ZipData)))
	if err != nil {
		return nil, err
	}

	files := make(map[string]File, len(zipReader.File))
	dirs := make(map[string][]string)
	fs := &StatiqFS{Files: files, Dirs: dirs}
	for _, zipFile := range zipReader.File {
		fi := zipFile.FileInfo()
		f := File{FileInfo: fi, Fs: fs}
		if f.Data, err = Unzip(zipFile); err != nil {
			return nil, fmt.Errorf("statiq/fs: error unzipping file %q: %s", zipFile.Name, err)
		}
		files["/"+zipFile.Name] = f
	}
	for fn := range files {
		// go up directories recursively in order to care deep directory
		for dn := path.Dir(fn); dn != fn; {
			if _, ok := files[dn]; ok {
				break
			}

			files[dn] = File{FileInfo: DirInfo{dn}, Fs: fs}
			fn, dn = dn, path.Dir(dn)
		}
	}
	for fn := range files {
		if dn := path.Dir(fn); fn != dn {
			fs.Dirs[dn] = append(fs.Dirs[dn], path.Base(fn))
		}
	}
	for _, s := range fs.Dirs {
		sort.Strings(s)
	}
	return fs, nil
}

var _ os.FileInfo = (*DirInfo)(nil)

type DirInfo struct{ name string }

func (di DirInfo) RawName() string    { return di.name }
func (di DirInfo) Name() string       { return path.Base(di.name) }
func (di DirInfo) Size() int64        { return 0 }
func (di DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (di DirInfo) ModTime() time.Time { return time.Time{} }
func (di DirInfo) IsDir() bool        { return true }
func (di DirInfo) Sys() interface{}   { return nil }

func Unzip(zf *zip.File) ([]byte, error) {
	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}

// Open returns a file matching the given file name, or os.ErrNotExists if
// no file matching the given file name is found in the archive.
// If a directory is requested, Open returns the file named "index.html"
// in the requested directory, if that file exists.
func (fs *StatiqFS) Open(name string) (http.File, error) {
	name = strings.Replace(name, "//", "/", -1)
	if f, ok := fs.Files[name]; ok {
		return NewHTTPFile(f), nil
	}

	return nil, os.ErrNotExist
}

func NewHTTPFile(file File) *HttpFile {
	if file.IsDir() {
		return &HttpFile{File: file, isDir: true}
	}
	return &HttpFile{File: file, reader: bytes.NewReader(file.Data)}
}

// HttpFile represents an HTTP file and acts as a bridge
// between file and http.File.
type HttpFile struct {
	File

	reader *bytes.Reader
	isDir  bool
	dirIdx int
}

// Read reads bytes into p, returns the number of read bytes.
func (f *HttpFile) Read(p []byte) (n int, err error) {
	if f.reader == nil && f.isDir {
		return 0, io.EOF
	}
	return f.reader.Read(p)
}

// Seek seeks to the offset.
func (f *HttpFile) Seek(offset int64, whence int) (ret int64, err error) {
	return f.reader.Seek(offset, whence)
}

func (f *HttpFile) Stat() (os.FileInfo, error) { return f, nil }  // Stat stats the file.
func (f *HttpFile) Close() error               { return nil }     // Close closes the file.
func (f *HttpFile) IsDir() bool                { return f.isDir } // IsDir returns true if the file location represents a directory.

// Readdir returns an empty slice of files, directory listing is disabled.
func (f *HttpFile) Readdir(count int) ([]os.FileInfo, error) {
	var fis []os.FileInfo
	if !f.isDir {
		return fis, nil
	}
	di, ok := f.FileInfo.(DirInfo)
	if !ok {
		return nil, fmt.Errorf("failed to read directory: %q", f.Name())
	}

	// If count is positive, the specified number of files will be returned,
	// and if negative, all remaining files will be returned.
	// The reading position of which file is returned is held in dirIndex.
	fnames := f.File.Fs.Dirs[di.name]
	flen := len(fnames)

	// If dirIdx reaches the end and the count is a positive value,
	// an io.EOF error is returned.
	// In other cases, no error will be returned even if, for example,
	// you specified more counts than the number of remaining files.
	start := f.dirIdx
	if start >= flen && count > 0 {
		return fis, io.EOF
	}

	var end int
	if count < 0 {
		end = flen
	} else {
		end = start + count
	}
	if end > flen {
		end = flen
	}

	for i := start; i < end; i++ {
		fis = append(fis, f.File.Fs.Files[path.Join(di.name, fnames[i])].FileInfo)
	}
	f.dirIdx += len(fis)
	return fis, nil
}
