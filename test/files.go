//
//  Copyright 2020 The AVFS authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//  	http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

package test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/avfs/avfs"
)

// FileCloseRead tests file Close function for read only files.
func (sfs *SuiteFS) FileCloseRead(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestFileCloseRead.txt")

	err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	openInfo, err := vfs.Stat(path)
	if err != nil {
		t.Fatalf("Stat %s : want error to be nil, got %v", path, err)
	}

	t.Run("FileCloseReadOnly", func(t *testing.T) {
		vfs = sfs.GetFsRead()

		f, err := vfs.Open(path)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		err = f.Close()
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		closeInfo, err := vfs.Stat(path)
		if err != nil {
			t.Errorf("Stat %s : want error to be nil, got %v", path, err)
		}

		if !reflect.DeepEqual(openInfo, closeInfo) {
			t.Errorf("Stat %s : open info != close info\n%v\n%v", path, openInfo, closeInfo)
		}

		err = f.Close()
		CheckPathError(t, "Close", "close", path, os.ErrClosed, err)
	})
}

// FileCloseWrite tests file Close function for read/write files.
func (sfs *SuiteFS) FileCloseWrite(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestFileCloseWrite.txt")

	err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	openInfo, err := vfs.Stat(path)
	if err != nil {
		t.Fatalf("Stat %s : want error to be nil, got %v", path, err)
	}

	t.Run("FileCloseWrite", func(t *testing.T) {
		f, err := vfs.OpenFile(path, os.O_APPEND|os.O_WRONLY, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		n, err := f.Write(data)
		if err != nil {
			t.Fatalf("Write : want error to be nil, got %v", err)
		}

		if n != len(data) {
			t.Fatalf("Write : want bytes written to be %d, got %d", len(data), n)
		}

		err = f.Close()
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		closeInfo, err := vfs.Stat(path)
		if err != nil {
			t.Errorf("Stat %s : want error to be nil, got %v", path, err)
		}

		if reflect.DeepEqual(openInfo, closeInfo) {
			t.Errorf("Stat %s : open info != close info\n%v\n%v", path, openInfo, closeInfo)
		}

		err = f.Close()
		CheckPathError(t, "Close", "close", path, os.ErrClosed, err)
	})
}

// FileFuncOnClosedFile tests functions on closed files.
func (sfs *SuiteFS) FileFuncOnClosedFile(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	existingFile := vfs.Join(rootDir, "existingFile")

	err := vfs.WriteFile(existingFile, nil, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	vfs = sfs.GetFsRead()

	t.Run("FileFuncOnClosedFile", func(t *testing.T) {
		f, err := vfs.Open(existingFile)
		if err != nil {
			t.Fatalf("Create : want error to be nil, got %v", err)
		}

		err = f.Close()
		if err != nil {
			t.Fatalf("Close : want error to be nil, got %v", err)
		}

		b := make([]byte, 1)

		err = f.Close()
		CheckPathError(t, "Close", "close", existingFile, os.ErrClosed, err)

		fd := f.Fd()
		if fd != math.MaxUint64 {
			t.Errorf("Fd %s : want Fd to be %d, got %d", existingFile, uint64(math.MaxUint64), fd)
		}

		name := f.Name()
		if name != existingFile {
			t.Errorf("Name %s : want Name to be %s, got %s", existingFile, existingFile, name)
		}

		_, err = f.Read(b)
		CheckPathError(t, "Read", "read", existingFile, os.ErrClosed, err)

		_, err = f.ReadAt(b, 0)
		CheckPathError(t, "ReadAt", "read", existingFile, os.ErrClosed, err)

		_, err = f.Seek(0, io.SeekStart)
		CheckPathError(t, "Seek", "seek", existingFile, os.ErrClosed, err)

		_, err = f.Readdir(-1)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Readdir", "Readdir", existingFile, avfs.ErrWinPathNotFound, err)
		default:
			if err.Error() != avfs.ErrFileClosing.Error() {
				t.Errorf("Readdir %s : want error to be %v, got %v", existingFile, avfs.ErrFileClosing, err)
			}
		}

		_, err = f.Readdirnames(-1)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Readdirnames", "Readdir", existingFile, avfs.ErrWinPathNotFound, err)
		default:
			if err.Error() != avfs.ErrFileClosing.Error() {
				t.Errorf("Readdirnames %s : want error to be %v, got %v", existingFile, avfs.ErrFileClosing, err)
			}
		}

		_, err = f.Stat()
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Stat", "GetFileType", existingFile, avfs.ErrFileClosing, err)
		default:
			CheckPathError(t, "Stat", "stat", existingFile, avfs.ErrFileClosing, err)
		}

		err = f.Sync()
		CheckPathError(t, "Sync", "sync", existingFile, os.ErrClosed, err)

		if vfs.HasFeature(avfs.FeatReadOnly) {
			return
		}

		err = f.Chdir()
		CheckPathError(t, "Chdir", "chdir", existingFile, os.ErrClosed, err)

		err = f.Chmod(avfs.DefaultFilePerm)
		CheckPathError(t, "Chmod", "chmod", existingFile, os.ErrClosed, err)

		if vfs.HasFeature(avfs.FeatIdentityMgr) {
			err = f.Chown(0, 0)
			CheckPathError(t, "Chown", "chown", existingFile, os.ErrClosed, err)
		}

		err = f.Sync()
		CheckPathError(t, "Sync", "sync", existingFile, os.ErrClosed, err)

		err = f.Truncate(0)
		CheckPathError(t, "Truncate", "truncate", existingFile, os.ErrClosed, err)

		_, err = f.Write(b)
		CheckPathError(t, "Write", "write", existingFile, os.ErrClosed, err)

		_, err = f.WriteAt(b, 0)
		CheckPathError(t, "WriteAt", "write", existingFile, os.ErrClosed, err)
	})
}

// FileRead tests Read and ReadAt functions.
func (sfs *SuiteFS) FileRead(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestFileRead.txt")

	err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	vfs = sfs.GetFsRead()

	t.Run("FileRead", func(t *testing.T) {
		const bufSize = 5

		f, err := vfs.OpenFile(path, os.O_RDONLY, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("OpenFile : want error to be nil, got %v", err)
		}

		defer f.Close()

		buf := make([]byte, bufSize)
		for i := 0; ; i += bufSize {
			n, err1 := f.Read(buf)
			if err1 != nil {
				if err1 == io.EOF {
					break
				}

				t.Errorf("Read : want error to be %v, got %v", io.EOF, err1)
			}

			if !bytes.Equal(buf[:n], data[i:i+n]) {
				t.Errorf("Read : want content to be %s, got %s", buf[:n], data[i:i+n])
			}
		}
	})

	t.Run("FileReadAt", func(t *testing.T) {
		const bufSize = 3

		f, err := vfs.OpenFile(path, os.O_RDONLY, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("OpenFile : want error to be nil, got %v", err)
		}

		defer f.Close()

		var n int
		rb := make([]byte, bufSize)
		for i := len(data); i > 0; i -= bufSize {
			n, err = f.ReadAt(rb, int64(i-bufSize))
			if err != nil {
				t.Errorf("ReadAt : want error to be nil, got %v", err)
			}

			if n != bufSize {
				t.Errorf("ReadAt : want bytes read to be %d, got %d", bufSize, n)
			}

			if !bytes.Equal(rb, data[i-bufSize:i]) {
				t.Errorf("ReadAt : want bytes read to be %d, got %d", bufSize, n)
			}
		}
	})

	t.Run("FileReadAfterEndOfFile", func(t *testing.T) {
		f, err := vfs.Open(path)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		b := make([]byte, 1)

		off := int64(len(data) * 2)

		n, err := f.ReadAt(b, off)
		if err != io.EOF {
			t.Errorf("ReadAt : want error to be %v, got %v", io.EOF, err)
		}

		if n != 0 {
			t.Errorf("ReadAt : want bytes read to be 0, got %d", n)
		}

		n, err = f.ReadAt(b, -1)
		CheckPathError(t, "ReadAt", "readat", path, avfs.ErrNegativeOffset, err)

		if n != 0 {
			t.Errorf("ReadAt : want bytes read to be 0, got %d", n)
		}
	})

	t.Run("FileReadOnDir", func(t *testing.T) {
		f, err := vfs.Open(rootDir)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		b := make([]byte, 1)

		_, err = f.Read(b)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Read", "read", rootDir, avfs.ErrWinInvalidHandle, err)
		default:
			CheckPathError(t, "Read", "read", rootDir, avfs.ErrIsADirectory, err)
		}

		_, err = f.ReadAt(b, 0)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "ReadAt", "read", rootDir, avfs.ErrWinInvalidHandle, err)
		default:
			CheckPathError(t, "ReadAt", "read", rootDir, avfs.ErrIsADirectory, err)
		}
	})
}

// FileSeek tests Seek function.
func (sfs *SuiteFS) FileSeek(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestFileSeek.txt")

	err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	vfs = sfs.GetFsRead()

	f, err := vfs.Open(path)
	if err != nil {
		t.Fatalf("Open : want error to be nil, got %v", err)
	}

	defer f.Close()

	var pos int64

	lenData := int64(len(data))

	t.Run("FileSeek", func(t *testing.T) {
		for i := 0; i < len(data); i++ {
			pos, err = f.Seek(int64(i), io.SeekStart)
			if err != nil {
				t.Errorf("Seek : want error to be nil, got %v", err)
			}

			if int(pos) != i {
				t.Errorf("Seek : want position to be %d, got %d", i, pos)
			}
		}

		for i := 0; i < len(data); i++ {
			pos, err = f.Seek(-int64(i), io.SeekEnd)
			if err != nil {
				t.Errorf("Seek : want error to be nil, got %v", err)
			}

			if int(pos) != len(data)-i {
				t.Errorf("Seek : want position to be %d, got %d", i, pos)
			}
		}

		_, err = f.Seek(0, io.SeekEnd)
		if err != nil {
			t.Fatalf("Seek : want error to be nil, got %v", err)
		}

		for i := len(data) - 1; i >= 0; i-- {
			pos, err = f.Seek(-1, io.SeekCurrent)
			if err != nil {
				t.Errorf("Seek : want error to be nil, got %v", err)
			}

			if int(pos) != i {
				t.Errorf("Seek : want position to be %d, got %d", i, pos)
			}
		}
	})

	t.Run("FileSeekInvalidStart", func(t *testing.T) {
		pos, err = f.Seek(-1, io.SeekStart)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Seek", "seek", f.Name(), avfs.ErrWinNegativeSeek, err)
		default:
			CheckPathError(t, "Seek", "seek", f.Name(), os.ErrInvalid, err)
		}

		if pos != 0 {
			t.Errorf("Seek : want pos to be %d, got %d", 0, pos)
		}

		wantPos := lenData * 2

		pos, err = f.Seek(wantPos, io.SeekStart)
		if err != nil {
			t.Errorf("Seek : want error to be nil, got %v", err)
		}

		if pos != wantPos {
			t.Errorf("Seek : want pos to be %d, got %d", wantPos, pos)
		}
	})

	t.Run("FileSeekInvalidEnd", func(t *testing.T) {
		pos, err = f.Seek(1, io.SeekEnd)
		if err != nil {
			t.Errorf("Seek : want error to be nil, got %v", err)
		}

		wantPos := lenData + 1
		if pos != wantPos {
			t.Errorf("Seek : want pos to be %d, got %d", wantPos, pos)
		}

		pos, err = f.Seek(-lenData*2, io.SeekEnd)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Seek", "seek", f.Name(), avfs.ErrWinNegativeSeek, err)
		default:
			CheckPathError(t, "Seek", "seek", f.Name(), os.ErrInvalid, err)
		}

		if pos != 0 {
			t.Errorf("Seek : want pos to be %d, got %d", 0, pos)
		}
	})

	t.Run("FileSeekInvalidCur", func(t *testing.T) {
		wantPos := lenData / 2

		pos, err = f.Seek(wantPos, io.SeekStart)
		if err != nil || pos != wantPos {
			t.Fatalf("Seek : want  pos to be 0 and error to be nil, got %d, %v", pos, err)
		}

		pos, err = f.Seek(-lenData, io.SeekCurrent)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Seek", "seek", f.Name(), avfs.ErrWinNegativeSeek, err)
		default:
			CheckPathError(t, "Seek", "seek", f.Name(), os.ErrInvalid, err)
		}

		if pos != 0 {
			t.Errorf("Seek : want pos to be %d, got %d", 0, pos)
		}

		pos, err = f.Seek(lenData, io.SeekCurrent)
		if err != nil {
			t.Errorf("Seek : want error to be nil, got %v", err)
		}

		if pos != lenData/2+lenData {
			t.Errorf("Seek : want pos to be %d, got %d", wantPos, pos)
		}
	})

	t.Run("FileSeekInvalidWhence", func(t *testing.T) {
		pos, err = f.Seek(0, 10)

		switch vfs.OSType() {
		case avfs.OsWindows:
			if err != nil {
				t.Errorf("Seek : want error to be nil, got %v", err)
			}
		default:
			CheckPathError(t, "Seek", "seek", f.Name(), os.ErrInvalid, err)
		}

		if pos != 0 {
			t.Errorf("Seek : want pos to be %d, got %d", 0, pos)
		}
	})

	t.Run("FileSeekOnDir", func(t *testing.T) {
		f, err = vfs.Open(rootDir)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		_, err = f.Seek(0, io.SeekStart)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Seek", "seek", rootDir, avfs.ErrWinInvalidHandle, err)
		default:
			if err != nil {
				t.Errorf("Seek : want error to be nil, got %v", err)
			}
		}
	})
}

// FileTruncate tests Truncate function.
func (sfs *SuiteFS) FileTruncate(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestFileTruncate.txt")

	t.Run("FileTruncate", func(t *testing.T) {
		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		f, err := vfs.OpenFile(path, os.O_RDWR, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("OpenFile : want error to be nil, got %v", err)
		}

		defer f.Close()

		b := make([]byte, len(data))
		for i := len(data) - 1; i >= 0; i-- {
			err = f.Truncate(int64(i))
			if err != nil {
				t.Errorf("Truncate : want error to be nil, got %v", err)
			}

			_, err = f.ReadAt(b, 0)
			if err != io.EOF {
				t.Errorf("Read : want error to be nil, got %v", err)
			}

			if !bytes.Equal(data[:i], b[:i]) {
				t.Errorf("Truncate : want data to be %s, got %s", data[:i], b[:i])
			}
		}
	})

	t.Run("Truncate", func(t *testing.T) {
		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		for i := len(data); i >= 0; i-- {
			err = vfs.Truncate(path, int64(i))
			if err != nil {
				t.Errorf("Truncate : want error to be nil, got %v", err)
			}

			d, err := vfs.ReadFile(path)
			if err != nil {
				t.Errorf("Truncate : want error to be nil, got %v", err)
			}

			if len(d) != i {
				t.Errorf("Truncate : want length to be %d, got %d", i, len(d))
			}
		}
	})

	t.Run("TruncateOnDir", func(t *testing.T) {
		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		err = vfs.Truncate(rootDir, 0)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Truncate", "open", rootDir, avfs.ErrIsADirectory, err)
		default:
			CheckPathError(t, "Truncate", "truncate", rootDir, avfs.ErrIsADirectory, err)
		}
	})

	t.Run("FileTruncateOnDir", func(t *testing.T) {
		f, err := vfs.Open(rootDir)
		if err != nil {
			t.Errorf("Truncate : want error to be nil, got %v", err)
		}

		defer f.Close()

		err = f.Truncate(0)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Truncate", "truncate", rootDir, avfs.ErrWinInvalidHandle, err)
		default:
			CheckPathError(t, "Truncate", "truncate", rootDir, os.ErrInvalid, err)
		}
	})

	t.Run("TruncateSizeNegative", func(t *testing.T) {
		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		err = vfs.Truncate(path, -1)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Truncate", "truncate", path, avfs.ErrWinNegativeSeek, err)
		default:
			CheckPathError(t, "Truncate", "truncate", path, os.ErrInvalid, err)
		}

		f, err := vfs.OpenFile(path, os.O_RDWR, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("OpenFile : want error to be nil, got %v", err)
		}

		defer f.Close()

		err = f.Truncate(-1)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Truncate", "truncate", path, avfs.ErrWinNegativeSeek, err)
		default:
			CheckPathError(t, "Truncate", "truncate", path, os.ErrInvalid, err)
		}
	})

	t.Run("TruncateSizeBiggerFileSize", func(t *testing.T) {
		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		newSize := len(data) * 2

		err = vfs.Truncate(path, int64(newSize))
		if err != nil {
			t.Errorf("Truncate : want error to be nil, got %v", err)
		}

		info, err := vfs.Stat(path)
		if err != nil {
			t.Errorf("Stat : want error to be nil, got %v", err)
		}

		if newSize != int(info.Size()) {
			t.Errorf("Stat : want size to be %d, got %d", newSize, info.Size())
		}

		gotContent, err := vfs.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile : want error to be nil, got %v", err)
		}

		wantAdded := bytes.Repeat([]byte{0}, len(data))
		gotAdded := gotContent[len(data):]
		if !bytes.Equal(wantAdded, gotAdded) {
			t.Errorf("Bytes Added : want %v, got %v", wantAdded, gotAdded)
		}
	})

	t.Run("TruncateNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		err := vfs.Truncate(nonExistingFile, 0)
		switch vfs.OSType() {
		case avfs.OsWindows:
			if err != nil {
				t.Errorf("Truncate : want error to be nil, got %v", err)
			}
		default:
			CheckPathError(t, "Truncate", "truncate", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
		}
	})
}

// FileWrite tests Write and WriteAt functions.
func (sfs *SuiteFS) FileWrite(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")

	t.Run("FileWrite", func(t *testing.T) {
		path := vfs.Join(rootDir, "TestFileWrite.txt")

		f, err := vfs.Create(path)
		if err != nil {
			t.Fatalf("Create : want error to be nil, got %v", err)
		}

		defer f.Close()

		for i := 0; i < len(data); i += 3 {
			buf3 := data[i : i+3]
			var n int

			n, err = f.Write(buf3)
			if err != nil {
				t.Errorf("Write : want error to be nil, got %v", err)
			}

			if len(buf3) != n {
				t.Errorf("Write : want bytes written to be %d, got %d", len(buf3), n)
			}
		}

		rb, err := vfs.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile : want error to be nil, got %v", err)
		}

		if !bytes.Equal(rb, data) {
			t.Errorf("ReadFile : want content to be %s, got %s", data, rb)
		}
	})

	t.Run("FileWriteAt", func(t *testing.T) {
		path := vfs.Join(rootDir, "TestFileWriteAt.txt")

		f, err := vfs.OpenFile(path, os.O_CREATE|os.O_RDWR, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("OpenFile : want error to be nil, got %v", err)
		}

		defer f.Close()

		for i := len(data); i > 0; i -= 3 {
			var n int
			n, err = f.WriteAt(data[i-3:i], int64(i-3))
			if err != nil {
				t.Errorf("WriteAt : want error to be nil, got %v", err)
			}

			if n != 3 {
				t.Errorf("WriteAt : want bytes written to be %d, got %d", 3, n)
			}
		}

		err = f.Close()
		if err != nil {
			t.Errorf("Close : want error to be nil, got %v", err)
		}

		rb, err := vfs.ReadFile(path)
		if err != nil {
			t.Errorf("ReadFile : want error to be nil, got %v", err)
		}

		if !bytes.Equal(rb, data) {
			t.Errorf("ReadFile : want content to be %s, got %s", data, rb)
		}
	})

	t.Run("FileWriteNegativeOffset", func(t *testing.T) {
		path := vfs.Join(rootDir, "TestFileWriteNO.txt")

		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		f, err := vfs.OpenFile(path, os.O_RDWR, avfs.DefaultDirPerm)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		n, err := f.WriteAt(data, -1)
		CheckPathError(t, "WriteAt", "writeat", path, avfs.ErrNegativeOffset, err)

		if n != 0 {
			t.Errorf("WriteAt : want bytes written to be 0, got %d", n)
		}
	})

	t.Run("FileWriteAtAfterEndOfFile", func(t *testing.T) {
		path := vfs.Join(rootDir, "TestFileWriteAfterEOF.txt")

		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		f, err := vfs.OpenFile(path, os.O_RDWR, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		off := int64(len(data) * 3)

		n, err := f.WriteAt(data, off)
		if err != nil {
			t.Errorf("WriteAt : want error to be nil, got %v", err)
		}

		if n != len(data) {
			t.Errorf("WriteAt : want bytes written to be %d, got %d", len(data), n)
		}

		want := make([]byte, int(off)+len(data))
		_ = copy(want, data)
		_ = copy(want[off:], data)

		got, err := vfs.ReadFile(path)
		if err != nil {
			t.Errorf("ReadFile : want error to be nil, got %v", err)
		}

		if !bytes.Equal(want, got) {
			t.Errorf("want : %s\ngot  : %s", want, got)
		}
	})

	t.Run("FileReadOnly", func(t *testing.T) {
		path := vfs.Join(rootDir, "TestFileReadOnly.txt")

		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		f, err := vfs.Open(path)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		b := make([]byte, len(data)*2)
		n, err := f.Write(b)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Write", "write", path, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "Write", "write", path, avfs.ErrBadFileDesc, err)
		}

		if n != 0 {
			t.Errorf("Write : want bytes written to be 0, got %d", n)
		}

		n, err = f.Read(b)
		if err != nil {
			t.Errorf("Read : want error to be nil, got %v", err)
		}

		if !bytes.Equal(data, b[:n]) {
			t.Errorf("Read : want data to be %s, got %s", data, b[:n])
		}

		n, err = f.WriteAt(b, 0)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "WriteAt", "write", path, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "WriteAt", "write", path, avfs.ErrBadFileDesc, err)
		}

		if n != 0 {
			t.Errorf("WriteAt : want bytes read to be 0, got %d", n)
		}

		n, err = f.ReadAt(b, 0)
		if err != io.EOF {
			t.Errorf("ReadAt : want error to be nil, got %v", err)
		}

		if !bytes.Equal(data, b[:n]) {
			t.Errorf("ReadAt : want data to be %s, got %s", data, b[:n])
		}
	})

	t.Run("FileWriteOnDir", func(t *testing.T) {
		f, err := vfs.Open(rootDir)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		b := make([]byte, 1)

		_, err = f.Write(b)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Write", "write", rootDir, avfs.ErrWinInvalidHandle, err)
		default:
			CheckPathError(t, "Write", "write", rootDir, avfs.ErrBadFileDesc, err)
		}

		_, err = f.WriteAt(b, 0)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "WriteAt", "write", rootDir, avfs.ErrWinInvalidHandle, err)
		default:
			CheckPathError(t, "WriteAt", "write", rootDir, avfs.ErrBadFileDesc, err)
		}
	})
}

// FileWriteTime checks that modification time is updated on write operations.
func (sfs *SuiteFS) FileWriteTime(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	data := []byte("AAABBBCCCDDD")
	existingFile := vfs.Join(rootDir, "ExistingFile.txt")

	var start, end int64

	f, err := vfs.Create(existingFile)
	if err != nil {
		t.Fatalf("Create : want error to be nil, got %v", err)
	}

	// CompareTime tests if the modification time of the file has changed.
	CompareTime := func(mustChange bool) {
		time.Sleep(10 * time.Millisecond)

		info, err := f.Stat() //nolint:govet // Shadows previous declaration of err.
		if err != nil {
			if errors.Unwrap(err).Error() != avfs.ErrFileClosing.Error() {
				t.Fatalf("Stat : want error to be nil, got %v", err)
			}

			info, err = vfs.Stat(existingFile)
			if err != nil {
				t.Fatalf("Stat : want error to be nil, got %v", err)
			}
		}

		start = end
		end = info.ModTime().UnixNano()

		// dont compare for the first time.
		if start == 0 {
			return
		}

		if mustChange && (start >= end) {
			t.Errorf("Stat %s : want start time < end time\nstart : %v\nend : %v", existingFile, start, end)
		}

		if !mustChange && (start != end) {
			t.Errorf("Stat %s : want start time == end time\nstart : %v\nend : %v", existingFile, start, end)
		}
	}

	CompareTime(true)

	t.Run("TimeWrite", func(t *testing.T) {
		_, err = f.Write(data)
		if err != nil {
			t.Fatalf("Write : want error to be nil, got %v", err)
		}

		CompareTime(true)
	})

	t.Run("TimeWriteAt", func(t *testing.T) {
		_, err = f.WriteAt(data, 5)
		if err != nil {
			t.Fatalf("WriteAt : want error to be nil, got %v", err)
		}

		CompareTime(true)
	})

	t.Run("TimeTruncate", func(t *testing.T) {
		err = f.Truncate(5)
		if err != nil {
			t.Fatalf("Truncate : want error to be nil, got %v", err)
		}

		CompareTime(true)
	})

	t.Run("TimeClose", func(t *testing.T) {
		err = f.Close()
		if err != nil {
			t.Fatalf("Close : want error to be nil, got %v", err)
		}

		CompareTime(false)
	})
}

// Link tests Link function.
func (sfs *SuiteFS) Link(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	if !vfs.HasFeature(avfs.FeatHardlink) {
		return
	}

	dirs := CreateDirs(t, vfs, rootDir)
	files := CreateFiles(t, vfs, rootDir)

	pathLinks := vfs.Join(rootDir, "links")

	err := vfs.Mkdir(pathLinks, avfs.DefaultDirPerm)
	if err != nil {
		t.Fatalf("mkdir %s : want error to be nil, got %v", pathLinks, err)
	}

	t.Run("LinkCreate", func(t *testing.T) {
		for _, file := range files {
			oldPath := vfs.Join(rootDir, file.Path)
			newPath := vfs.Join(pathLinks, vfs.Base(file.Path))

			err := vfs.Link(oldPath, newPath)
			if err != nil {
				t.Errorf("Link %s %s : want error to be nil, got %v", oldPath, newPath, err)
			}

			newContent, err := vfs.ReadFile(newPath)
			if err != nil {
				t.Errorf("Readfile %s : want error to be nil, got %v", newPath, err)
			}

			if !bytes.Equal(file.Content, newContent) {
				t.Errorf("ReadFile %s : want content to be %s, got %s", newPath, file.Content, newContent)
			}
		}
	})

	t.Run("LinkExisting", func(t *testing.T) {
		for _, file := range files {
			oldPath := vfs.Join(rootDir, file.Path)
			newPath := vfs.Join(pathLinks, vfs.Base(file.Path))

			err := vfs.Link(oldPath, newPath)
			CheckLinkError(t, "Link", "link", oldPath, newPath, avfs.ErrFileExists, err)
		}
	})

	t.Run("LinkRemove", func(t *testing.T) {
		for _, file := range files {
			oldPath := vfs.Join(rootDir, file.Path)
			newPath := vfs.Join(pathLinks, vfs.Base(file.Path))

			err := vfs.Remove(oldPath)
			if err != nil {
				t.Errorf("Remove %s : want error to be nil, got %v", oldPath, err)
			}

			newContent, err := vfs.ReadFile(newPath)
			if err != nil {
				t.Errorf("Readfile %s : want error to be nil, got %v", newPath, err)
			}

			if !bytes.Equal(file.Content, newContent) {
				t.Errorf("ReadFile %s : want content to be %s, got %s", newPath, file.Content, newContent)
			}
		}
	})

	t.Run("LinkErrorDir", func(t *testing.T) {
		for _, dir := range dirs {
			oldPath := vfs.Join(rootDir, dir.Path)
			newPath := vfs.Join(rootDir, "WhateverDir")

			err := vfs.Link(oldPath, newPath)
			CheckLinkError(t, "Link", "link", oldPath, newPath, avfs.ErrOpNotPermitted, err)
		}
	})

	t.Run("LinkErrorFile", func(t *testing.T) {
		for _, file := range files {
			InvalidPath := vfs.Join(rootDir, file.Path, "OldInvalidPath")
			NewInvalidPath := vfs.Join(pathLinks, "WhateverFile")

			err := vfs.Link(InvalidPath, NewInvalidPath)
			CheckLinkError(t, "Link", "link", InvalidPath, NewInvalidPath, avfs.ErrNoSuchFileOrDir, err)
		}
	})

	t.Run("LinkNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")
		existingFile := vfs.Join(rootDir, "existingFile")

		err := vfs.WriteFile(existingFile, nil, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		err = vfs.Link(nonExistingFile, existingFile)
		CheckLinkError(t, "Link", "link", nonExistingFile, nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
	})
}

// NilPtrFile test calls to File methods when f is a nil File.
func NilPtrFile(t *testing.T, f avfs.File) {
	err := f.Chdir()
	CheckInvalid(t, "Chdir", err)

	err = f.Chmod(0)
	CheckInvalid(t, "Chmod", err)

	err = f.Chown(0, 0)
	CheckInvalid(t, "Chown", err)

	err = f.Close()
	CheckInvalid(t, "Close", err)

	CheckPanic(t, "f.Name()", func() { _ = f.Name() })

	fd := f.Fd()
	if fd != math.MaxUint64 {
		t.Errorf("Fd : want fd to be %d, got %d", 0, fd)
	}

	_, err = f.Read([]byte{})
	CheckInvalid(t, "Read", err)

	_, err = f.ReadAt([]byte{}, 0)
	CheckInvalid(t, "ReadAt", err)

	_, err = f.Readdir(0)
	CheckInvalid(t, "Readdir", err)

	_, err = f.Readdirnames(0)
	CheckInvalid(t, "Readdirnames", err)

	_, err = f.Seek(0, io.SeekStart)
	CheckInvalid(t, "Seek", err)

	_, err = f.Stat()
	CheckInvalid(t, "Stat", err)

	err = f.Sync()
	CheckInvalid(t, "Sync", err)

	err = f.Truncate(0)
	CheckInvalid(t, "Truncate", err)

	_, err = f.Write([]byte{})
	CheckInvalid(t, "Write", err)

	_, err = f.WriteAt([]byte{}, 0)
	CheckInvalid(t, "WriteAt", err)

	_, err = f.WriteString("")
	CheckInvalid(t, "WriteString", err)
}

// OpenFileRead tests OpenFile function for read.
func (sfs *SuiteFS) OpenFileRead(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")
	existingFile := vfs.Join(rootDir, "ExistingFile.txt")

	err := vfs.WriteFile(existingFile, data, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	existingDir := vfs.Join(rootDir, "existingDir")

	err = vfs.Mkdir(existingDir, avfs.DefaultDirPerm)
	if err != nil {
		t.Fatalf("Mkdir : want error to be nil, got %v", err)
	}

	vfs = sfs.GetFsRead()

	t.Run("OpenFileReadOnly", func(t *testing.T) {
		f, err := vfs.Open(existingFile)
		if err != nil {
			t.Errorf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		gotData, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("ReadAll : want error to be nil, got %v", err)
		}

		if !bytes.Equal(gotData, data) {
			t.Errorf("ReadAll : want error data to be %v, got %v", data, gotData)
		}
	})

	t.Run("OpenFileDirReadOnly", func(t *testing.T) {
		f, err := vfs.OpenFile(existingDir, os.O_RDONLY, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("OpenFile : want error to be nil, got %v", err)
		}

		defer f.Close()

		dirs, err := f.Readdir(-1)
		if err != nil {
			t.Errorf("Readdir : want error to be nil, got %v", err)
		}

		if len(dirs) != 0 {
			t.Errorf("Readdir : want number of directories to be 0, got %d", len(dirs))
		}
	})

	t.Run("OpenNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")
		buf := make([]byte, 1)

		f, err := vfs.Open(nonExistingFile)
		switch vfs.OSType() {
		case avfs.OsWindows:
			if err != nil {
				t.Errorf("Truncate : want error to be nil, got %v", err)
			}
		default:
			CheckPathError(t, "Open", "open", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
		}

		if f == nil {
			t.Fatal("Open : want f to be != nil, got nil")
		}

		err = f.Chdir()
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Chdir", "chdir", nonExistingFile, avfs.ErrWinNotSupported, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Chdir : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		err = f.Chmod(0)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Chmod", "chmod", nonExistingFile, avfs.ErrWinNotSupported, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Chmod : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		if vfs.HasFeature(avfs.FeatIdentityMgr) {
			err = f.Chown(0, 0)
			if err != os.ErrInvalid {
				t.Errorf("Chown : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		err = f.Close()
		switch vfs.OSType() {
		case avfs.OsWindows:
			if err != nil {
				t.Errorf("Truncate : want error to be nil, got %v", err)
			}
		default:
			if err != os.ErrInvalid {
				t.Errorf("Close : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.Read(buf)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Read", "read", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Read : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.ReadAt(buf, 0)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "ReadAt", "read", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("ReadAt : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.Readdir(-1)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Readdir", "Readdir", nonExistingFile, avfs.ErrWinPathNotFound, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Readdir : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.Readdirnames(-1)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Readdirnames", "Readdir", nonExistingFile, avfs.ErrWinPathNotFound, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Readdirnames : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.Seek(0, io.SeekStart)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Seek", "seek", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Seek : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.Stat()
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Stat", "GetFileType", nonExistingFile, avfs.ErrFileClosing, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Stat : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		err = f.Sync()
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Sync", "sync", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Sync : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		err = f.Truncate(0)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Truncate", "truncate", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Truncate : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.Write(buf)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Write", "write", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Write : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.WriteAt(buf, 0)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "WriteAt", "write", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("WriteAt : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		_, err = f.WriteString("")
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "WriteString", "write", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("WriteString : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}

		err = f.Close()
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Close", "close", nonExistingFile, os.ErrClosed, err)
		default:
			if err != os.ErrInvalid {
				t.Errorf("Close : want error to be %v, got %v", os.ErrInvalid, err)
			}
		}
	})
}

// OpenFileWrite tests OpenFile function for write.
func (sfs *SuiteFS) OpenFileWrite(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")
	whateverData := []byte("whatever")
	existingFile := vfs.Join(rootDir, "ExistingFile.txt")
	buf3 := make([]byte, 3)

	err := vfs.WriteFile(existingFile, data, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	t.Run("OpenFileWriteOnly", func(t *testing.T) {
		f, err := vfs.OpenFile(existingFile, os.O_WRONLY, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		n, err := f.Write(whateverData)
		if err != nil {
			t.Errorf("Write : want error to be nil, got %v", err)
		}

		if n != len(whateverData) {
			t.Errorf("Write : want bytes written to be %d, got %d", len(whateverData), n)
		}

		n, err = f.Read(buf3)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Read", "read", existingFile, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "Read", "read", existingFile, avfs.ErrBadFileDesc, err)
		}

		if n != 0 {
			t.Errorf("Read : want bytes written to be 0, got %d", n)
		}

		n, err = f.ReadAt(buf3, 3)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "ReadAt", "read", existingFile, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "ReadAt", "read", existingFile, avfs.ErrBadFileDesc, err)
		}

		if n != 0 {
			t.Errorf("ReadAt : want bytes read to be 0, got %d", n)
		}

		err = f.Chmod(0o777)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Chmod", "chmod", existingFile, avfs.ErrWinNotSupported, err)
		default:
			if err != nil {
				t.Errorf("Chmod : want error to be nil, got %v", err)
			}
		}

		if vfs.HasFeature(avfs.FeatIdentityMgr) {
			u := vfs.CurrentUser()
			err = f.Chown(u.Uid(), u.Gid())
			if err != nil {
				t.Errorf("Chown : want error to be nil, got %v", err)
			}
		}

		fst, err := f.Stat()
		if err != nil {
			t.Errorf("Stat : want error to be nil, got %v", err)
		}

		wantName := vfs.Base(f.Name())
		if wantName != fst.Name() {
			t.Errorf("Stat : want name to be %s, got %s", wantName, fst.Name())
		}

		err = f.Truncate(0)
		if err != nil {
			t.Errorf("Chmod : want error to be nil, got %v", err)
		}

		err = f.Sync()
		if err != nil {
			t.Errorf("Sync : want error to be nil, got %v", err)
		}
	})

	t.Run("OpenFileAppend", func(t *testing.T) {
		err := vfs.WriteFile(existingFile, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("Chmod : want error to be nil, got %v", err)
		}

		f, err := vfs.OpenFile(existingFile, os.O_WRONLY|os.O_APPEND, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("OpenFile : want error to be nil, got %v", err)
		}

		n, err := f.Write(whateverData)
		if err != nil {
			t.Errorf("Write : want error to be nil, got %v", err)
		}

		if n != len(whateverData) {
			t.Errorf("Write : want error to be %d, got %d", len(whateverData), n)
		}

		_ = f.Close()

		gotContent, err := vfs.ReadFile(existingFile)
		if err != nil {
			t.Errorf("ReadFile : want error to be nil, got %v", err)
		}

		wantContent := append(data, whateverData...)
		if !bytes.Equal(wantContent, gotContent) {
			t.Errorf("ReadAll : want content to be %s, got %s", wantContent, gotContent)
		}
	})

	t.Run("OpenFileReadOnly", func(t *testing.T) {
		f, err := vfs.Open(existingFile)
		if err != nil {
			t.Errorf("Open : want error to be nil, got %v", err)
		}

		defer f.Close()

		n, err := f.Write(whateverData)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Write", "write", existingFile, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "Write", "write", existingFile, avfs.ErrBadFileDesc, err)
		}

		if n != 0 {
			t.Errorf("Write : want bytes written to be 0, got %d", n)
		}

		n, err = f.WriteAt(whateverData, 3)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "WriteAt", "write", existingFile, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "WriteAt", "write", existingFile, avfs.ErrBadFileDesc, err)
		}

		if n != 0 {
			t.Errorf("WriteAt : want bytes written to be 0, got %d", n)
		}

		err = f.Chmod(0o777)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Chmod", "chmod", existingFile, avfs.ErrWinNotSupported, err)
		default:
			if err != nil {
				t.Errorf("Chmod : want error to be nil, got %v", err)
			}
		}

		if vfs.HasFeature(avfs.FeatIdentityMgr) {
			u := vfs.CurrentUser()
			err = f.Chown(u.Uid(), u.Gid())
			if err != nil {
				t.Errorf("Chown : want error to be nil, got %v", err)
			}
		}

		fst, err := f.Stat()
		if err != nil {
			t.Errorf("Stat : want error to be nil, got %v", err)
		}

		wantName := vfs.Base(f.Name())
		if wantName != fst.Name() {
			t.Errorf("Stat : want name to be %s, got %s", wantName, fst.Name())
		}

		err = f.Truncate(0)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Truncate", "truncate", existingFile, avfs.ErrWinAccessDenied, err)
		default:
			CheckPathError(t, "Truncate", "truncate", existingFile, os.ErrInvalid, err)
		}
	})

	t.Run("OpenFileDir", func(t *testing.T) {
		existingDir := vfs.Join(rootDir, "existingDir")

		err := vfs.Mkdir(existingDir, avfs.DefaultDirPerm)
		if err != nil {
			t.Fatalf("Mkdir : want error to be nil, got %v", err)
		}

		f, err := vfs.OpenFile(existingDir, os.O_WRONLY, avfs.DefaultFilePerm)
		CheckPathError(t, "OpenFile", "open", existingDir, avfs.ErrIsADirectory, err)

		if !reflect.ValueOf(f).IsNil() {
			t.Errorf("OpenFile : want file to be nil, got %v", f)
		}
	})

	t.Run("OpenFileExcl", func(t *testing.T) {
		fileExcl := vfs.Join(rootDir, "fileExcl")

		f, err := vfs.OpenFile(fileExcl, os.O_CREATE|os.O_EXCL, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("OpenFile : want error to be nil, got %v", err)
		}

		f.Close()

		_, err = vfs.OpenFile(fileExcl, os.O_CREATE|os.O_EXCL, avfs.DefaultFilePerm)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "OpenFile", "open", fileExcl, avfs.ErrWinFileExists, err)
		default:
			CheckPathError(t, "OpenFile", "open", fileExcl, avfs.ErrFileExists, err)
		}
	})

	t.Run("OpenFileNonExistingPath", func(t *testing.T) {
		nonExistingPath := vfs.Join(rootDir, "non/existing/path")
		_, err := vfs.OpenFile(nonExistingPath, os.O_CREATE, avfs.DefaultFilePerm)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "OpenFile", "open", nonExistingPath, avfs.ErrWinPathNotFound, err)
		default:
			CheckPathError(t, "OpenFile", "open", nonExistingPath, avfs.ErrNoSuchFileOrDir, err)
		}
	})
}

// ReadFile tests ReadFile function.
func (sfs *SuiteFS) ReadFile(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsRead()

	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestReadFile.txt")

	t.Run("ReadFile", func(t *testing.T) {
		rb, err := vfs.ReadFile(path)
		if err == nil {
			t.Errorf("ReadFile : want error to be %v, got nil", avfs.ErrNoSuchFileOrDir)
		}

		if len(rb) != 0 {
			t.Errorf("ReadFile : want read bytes to be 0, got %d", len(rb))
		}

		vfs = sfs.GetFsWrite()

		err = vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		vfs = sfs.GetFsRead()

		rb, err = vfs.ReadFile(path)
		if err != nil {
			t.Errorf("ReadFile : want error to be nil, got %v", err)
		}

		if !bytes.Equal(rb, data) {
			t.Errorf("ReadFile : want content to be %s, got %s", data, rb)
		}
	})
}

// SameFile tests SameFile function.
func (sfs *SuiteFS) SameFile(t *testing.T) {
	rootDir1, removeDir1 := sfs.CreateRootDir(t, UsrTest)
	defer removeDir1()

	vfs := sfs.GetFsWrite()
	CreateDirs(t, vfs, rootDir1)
	files := CreateFiles(t, vfs, rootDir1)

	rootDir2, removeDir2 := sfs.CreateRootDir(t, UsrTest)
	defer removeDir2()
	CreateDirs(t, vfs, rootDir2)

	t.Run("SameFileLink", func(t *testing.T) {
		if !vfs.HasFeature(avfs.FeatHardlink) {
			return
		}

		for _, file := range files {
			path1 := vfs.Join(rootDir1, file.Path)
			path2 := vfs.Join(rootDir2, file.Path)

			info1, err := vfs.Stat(path1)
			if err != nil {
				t.Fatalf("Stat %s : want error to be nil, got %v", path1, err)
			}

			err = vfs.Link(path1, path2)
			if err != nil {
				t.Fatalf("Link %s : want error to be nil, got %v", path1, err)
			}

			info2, err := vfs.Stat(path2)
			if err != nil {
				t.Fatalf("Stat %s : want error to be nil, got %v", path1, err)
			}

			if !vfs.SameFile(info1, info2) {
				t.Fatalf("SameFile %s, %s : not same files\n%v\n%v", path1, path2, info1, info2)
			}

			err = vfs.Remove(path2)
			if err != nil {
				t.Fatalf("Remove %s : want error to be nil, got %v", path2, err)
			}
		}
	})

	t.Run("SameFileSymlink", func(t *testing.T) {
		if !vfs.HasFeature(avfs.FeatSymlink) {
			return
		}

		for _, file := range files {
			path1 := vfs.Join(rootDir1, file.Path)
			path2 := vfs.Join(rootDir2, file.Path)

			info1, err := vfs.Stat(path1)
			if err != nil {
				t.Fatalf("Stat %s : want error to be nil, got %v", path1, err)
			}

			err = vfs.Symlink(path1, path2)
			if err != nil {
				t.Fatalf("Symlink %s : want error to be nil, got %v", path1, err)
			}

			info2, err := vfs.Stat(path2)
			if err != nil {
				t.Fatalf("Stat %s : want error to be nil, got %v", path1, err)
			}

			if !vfs.SameFile(info1, info2) {
				t.Fatalf("SameFile %s, %s : not same files\n%v\n%v", path1, path2, info1, info2)
			}

			info3, err := vfs.Lstat(path2)
			if err != nil {
				t.Fatalf("Stat %s : want error to be nil, got %v", path1, err)
			}

			if vfs.SameFile(info1, info3) {
				t.Fatalf("SameFile %s, %s : not the same file\n%v\n%v", path1, path2, info1, info3)
			}

			err = vfs.Remove(path2)
			if err != nil {
				t.Fatalf("Remove %s : want error to be nil, got %v", path2, err)
			}
		}
	})
}

// WriteFile tests WriteFile function.
func (sfs *SuiteFS) WriteFile(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")

	t.Run("WriteFile", func(t *testing.T) {
		path := vfs.Join(rootDir, "WriteFile.txt")

		err := vfs.WriteFile(path, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Errorf("WriteFile : want error to be nil, got %v", err)
		}

		rb, err := vfs.ReadFile(path)
		if err != nil {
			t.Errorf("ReadFile : want error to be nil, got %v", err)
		}

		if !bytes.Equal(rb, data) {
			t.Errorf("ReadFile : want content to be %s, got %s", data, rb)
		}
	})
}

// WriteString tests WriteString function.
func (sfs *SuiteFS) WriteString(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	data := []byte("AAABBBCCCDDD")
	path := vfs.Join(rootDir, "TestWriteString.txt")

	t.Run("WriteString", func(t *testing.T) {
		f, err := vfs.Create(path)
		if err != nil {
			t.Errorf("Create %s : want error to be nil, got %v", path, err)
		}

		n, err := f.WriteString(string(data))
		if err != nil {
			t.Errorf("WriteString : want error to be nil, got %v", err)
		}

		if len(data) != n {
			t.Errorf("WriteString : want written bytes to be %d, got %d", len(data), n)
		}

		f.Close()

		rb, err := vfs.ReadFile(path)
		if err != nil {
			t.Errorf("ReadFile : want error to be nil, got %v", err)
		}

		if !bytes.Equal(rb, data) {
			t.Errorf("ReadFile : want content to be %s, got %s", data, rb)
		}
	})
}
