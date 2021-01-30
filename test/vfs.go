//
//  Copyright 2021 The AVFS authors
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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfsutils"
)

// Chdir tests Chdir and Getwd functions.
func (sfs *SuiteFS) Chdir(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.Chdir(rootDir)
		CheckPathError(t, "Chdir", "chdir", rootDir, avfs.ErrPermDenied, err)

		_, err = vfs.Getwd()
		CheckPathError(t, "Getwd", "getwd", "", avfs.ErrPermDenied, err)

		return
	}

	dirs := CreateDirs(t, vfs, rootDir)
	existingFile := CreateEmptyFile(t, vfs, rootDir)

	vfs = sfs.GetFsRead()

	t.Run("ChdirAbsolute", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			err := vfs.Chdir(path)
			if err != nil {
				t.Errorf("Chdir %s : want error to be nil, got %v", path, err)
			}

			curDir, err := vfs.Getwd()
			if err != nil {
				t.Errorf("Getwd %s : want error to be nil, got %v", path, err)
			}

			if curDir != path {
				t.Errorf("Getwd : want current directory to be %s, got %s", path, curDir)
			}
		}
	})

	t.Run("ChdirRelative", func(t *testing.T) {
		for _, dir := range dirs {
			err := vfs.Chdir(rootDir)
			if err != nil {
				t.Fatalf("Chdir %s : want error to be nil, got %v", rootDir, err)
			}

			relPath := dir.Path[1:]

			err = vfs.Chdir(relPath)
			if err != nil {
				t.Errorf("Chdir %s : want error to be nil, got %v", relPath, err)
			}

			curDir, err := vfs.Getwd()
			if err != nil {
				t.Errorf("Getwd : want error to be nil, got %v", err)
			}

			path := vfs.Join(rootDir, relPath)
			if curDir != path {
				t.Errorf("Getwd : want current directory to be %s, got %s", path, curDir)
			}
		}
	})

	t.Run("ChdirNonExisting", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path, "NonExistingDir")

			oldPath, err := vfs.Getwd()
			if err != nil {
				t.Errorf("Chdir : want error to be nil, got %v", err)
			}

			err = vfs.Chdir(path)
			CheckPathError(t, "Chdir", "chdir", path, avfs.ErrNoSuchFileOrDir, err)

			newPath, err := vfs.Getwd()
			if err != nil {
				t.Errorf("Getwd : want error to be nil, got %v", err)
			}

			if newPath != oldPath {
				t.Errorf("Getwd : want current dir to be %s, got %s", oldPath, newPath)
			}
		}
	})

	t.Run("ChdirOnFile", func(t *testing.T) {
		err := vfs.Chdir(existingFile)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Chdir", "chdir", existingFile, avfs.ErrWinDirNameInvalid, err)
		default:
			CheckPathError(t, "Chdir", "chdir", existingFile, avfs.ErrNotADirectory, err)
		}
	})
}

// Chtimes tests Chtimes function.
func (sfs *SuiteFS) Chtimes(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.Chtimes(rootDir, time.Now(), time.Now())
		CheckPathError(t, "Chtimes", "chtimes", rootDir, avfs.ErrPermDenied, err)

		return
	}

	t.Run("Chtimes", func(t *testing.T) {
		_ = CreateDirs(t, vfs, rootDir)
		files := CreateFiles(t, vfs, rootDir)
		tomorrow := time.Now().AddDate(0, 0, 1)

		for _, file := range files {
			path := vfs.Join(rootDir, file.Path)

			err := vfs.Chtimes(path, tomorrow, tomorrow)
			if err != nil {
				t.Errorf("Chtimes %s : want error to be nil, got %v", path, err)
			}

			infos, err := vfs.Stat(path)
			if err != nil {
				t.Errorf("Chtimes %s : want error to be nil, got %v", path, err)
			}

			if infos.ModTime() != tomorrow {
				t.Errorf("Chtimes %s : want modtime to bo %s, got %s", path, tomorrow, infos.ModTime())
			}
		}
	})

	t.Run("ChtimesNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		err := vfs.Chtimes(nonExistingFile, time.Now(), time.Now())
		CheckPathError(t, "Chtimes", "chtimes", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
	})
}

// Clone tests Clone function.
func (sfs *SuiteFS) Clone(t *testing.T) {
	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		if vfsClonable, ok := vfs.(avfs.Cloner); ok {
			vfsCloned := vfsClonable.Clone()

			if _, ok := vfsCloned.(avfs.Cloner); !ok {
				t.Errorf("Clone : want cloned vfs to be of type VFS, got type %v", reflect.TypeOf(vfsCloned))
			}
		}
	}
}

func (sfs *SuiteFS) Create(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.Create(rootDir)
		CheckPathError(t, "Create", "open", rootDir, avfs.ErrPermDenied, err)

		return
	}
}

// EvalSymlink tests EvalSymlink function.
func (sfs *SuiteFS) EvalSymlink(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()
	if !vfs.HasFeature(avfs.FeatSymlink) {
		_, err := vfs.EvalSymlinks(rootDir)
		CheckPathError(t, "EvalSymlinks", "lstat", rootDir, avfs.ErrPermDenied, err)

		return
	}

	_ = CreateDirs(t, vfs, rootDir)
	_ = CreateFiles(t, vfs, rootDir)
	_ = CreateSymlinks(t, vfs, rootDir)

	vfs = sfs.GetFsRead()

	t.Run("EvalSymlink", func(t *testing.T) {
		symlinks := GetSymlinksEval(vfs)
		for _, sl := range symlinks {
			wantOp := "lstat"
			wantPath := vfs.Join(rootDir, sl.OldName)
			slPath := vfs.Join(rootDir, sl.NewName)

			gotPath, err := vfs.EvalSymlinks(slPath)
			if sl.WantErr == nil && err == nil {
				if wantPath != gotPath {
					t.Errorf("EvalSymlinks %s : want Path to be %s, got %s", slPath, wantPath, gotPath)
				}

				continue
			}

			e, ok := err.(*os.PathError)
			if !ok && sl.WantErr != err {
				t.Errorf("EvalSymlinks %s : want error %v, got %v", slPath, sl.WantErr, err)
			}

			if wantOp != e.Op || wantPath != e.Path || sl.WantErr != e.Err {
				t.Errorf("EvalSymlinks %s : error"+
					"\nwant : Op: %s, Path: %s, Err: %v\ngot  : Op: %s, Path: %s, Err: %v",
					sl.NewName, wantOp, wantPath, sl.WantErr, e.Op, e.Path, e.Err)
			}
		}
	})
}

// GetTempDir tests GetTempDir function.
func (sfs *SuiteFS) GetTempDir(t *testing.T) {
	vfs := sfs.GetFsRead()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		tmp := vfs.GetTempDir()
		if tmp != avfs.TmpDir {
			t.Errorf("GetTempDir : want error to be %v, got %v", avfs.NotImplemented, tmp)
		}

		return
	}

	var wantTmp string

	switch vfs.OSType() {
	case avfs.OsDarwin:
		wantTmp, _ = filepath.EvalSymlinks(os.TempDir())
	case avfs.OsWindows:
		wantTmp = os.Getenv("TMP")
	default:
		wantTmp = avfs.TmpDir
	}

	gotTmp := vfs.GetTempDir()
	if gotTmp != wantTmp {
		t.Fatalf("GetTempDir : want temp dir to be %s, got %s", wantTmp, gotTmp)
	}
}

// Link tests Link function.
func (sfs *SuiteFS) Link(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatHardlink) {
		err := vfs.Link(rootDir, rootDir)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckLinkError(t, "Link", "link", rootDir, rootDir, avfs.ErrWinPathNotFound, err)
		default:
			CheckLinkError(t, "Link", "link", rootDir, rootDir, avfs.ErrPermDenied, err)
		}

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

// Lstat tests Lstat function.
func (sfs *SuiteFS) Lstat(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.Lstat(rootDir)
		CheckPathError(t, "Lstat", "lstat", rootDir, avfs.ErrPermDenied, err)

		return
	}

	dirs := CreateDirs(t, vfs, rootDir)
	files := CreateFiles(t, vfs, rootDir)
	CreateSymlinks(t, vfs, rootDir)

	vfs = sfs.GetFsRead()

	t.Run("LstatDir", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			info, err := vfs.Lstat(path)
			if err != nil {
				t.Errorf("Lstat %s : want error to be nil, got %v", path, err)

				continue
			}

			if vfs.Base(path) != info.Name() {
				t.Errorf("Lstat %s : want name to be %s, got %s", path, vfs.Base(path), info.Name())
			}

			wantMode := (dir.Mode | os.ModeDir) &^ vfs.GetUMask()
			if vfs.OSType() == avfs.OsWindows {
				wantMode = os.ModeDir | os.ModePerm
			}

			if wantMode != info.Mode() {
				t.Errorf("Lstat %s : want mode to be %s, got %s", path, wantMode, info.Mode())
			}
		}
	})

	t.Run("LstatFile", func(t *testing.T) {
		for _, file := range files {
			path := vfs.Join(rootDir, file.Path)

			info, err := vfs.Lstat(path)
			if err != nil {
				t.Errorf("Lstat %s : want error to be nil, got %v", path, err)

				continue
			}

			if info.Name() != vfs.Base(path) {
				t.Errorf("Lstat %s : want name to be %s, got %s", path, vfs.Base(path), info.Name())
			}

			wantMode := file.Mode &^ vfs.GetUMask()
			if vfs.OSType() == avfs.OsWindows {
				wantMode = 0o666
			}

			if wantMode != info.Mode() {
				t.Errorf("Lstat %s : want mode to be %s, got %s", path, wantMode, info.Mode())
			}

			wantSize := int64(len(file.Content))
			if wantSize != info.Size() {
				t.Errorf("Lstat %s : want size to be %d, got %d", path, wantSize, info.Size())
			}
		}
	})

	t.Run("LstatSymlink", func(t *testing.T) {
		for _, sl := range GetSymlinksEval(vfs) {
			newPath := vfs.Join(rootDir, sl.NewName)
			oldPath := vfs.Join(rootDir, sl.OldName)

			info, err := vfs.Lstat(newPath)
			if err != nil {
				if sl.WantErr == nil {
					t.Errorf("Lstat %s : want error to be nil, got %v", newPath, err)
				}

				CheckPathError(t, "Lstat", "stat", newPath, sl.WantErr, err)

				continue
			}

			var (
				wantName string
				wantMode os.FileMode
			)

			if sl.IsSymlink {
				wantName = vfs.Base(newPath)
				wantMode = os.ModeSymlink | os.ModePerm
			} else {
				wantName = vfs.Base(oldPath)
				wantMode = sl.Mode
			}

			if wantName != info.Name() {
				t.Errorf("Lstat %s : want name to be %s, got %s", newPath, wantName, info.Name())
			}

			if wantMode != info.Mode() {
				t.Errorf("Lstat %s : want mode to be %s, got %s", newPath, wantMode, info.Mode())
			}
		}
	})

	t.Run("LStatNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		_, err := vfs.Lstat(nonExistingFile)
		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Lstat", "CreateFile", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
		default:
			CheckPathError(t, "Lstat", "lstat", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
		}
	})

	t.Run("LStatSubDirOnFile", func(t *testing.T) {
		subDirOnFile := vfs.Join(rootDir, files[0].Path, "subDirOnFile")

		_, err := vfs.Lstat(subDirOnFile)
		CheckPathError(t, "Lstat", "lstat", subDirOnFile, avfs.ErrNotADirectory, err)
	})
}

// Mkdir tests Mkdir function.
func (sfs *SuiteFS) Mkdir(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.Mkdir(rootDir, avfs.DefaultDirPerm)
		CheckPathError(t, "Mkdir", "mkdir", rootDir, avfs.ErrPermDenied, err)

		return
	}

	existingFile := CreateEmptyFile(t, vfs, rootDir)

	vfs = sfs.GetFsRead()
	dirs := GetDirs()

	t.Run("MkdirNew", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			err := vfs.Mkdir(path, dir.Mode)
			if err != nil {
				t.Errorf("mkdir : want no error, got %v", err)
			}

			fi, err := vfs.Stat(path)
			if err != nil {
				t.Errorf("stat '%s' : want no error, got %v", path, err)

				continue
			}

			if !fi.IsDir() {
				t.Errorf("stat '%s' : want path to be a directory, got mode %s", path, fi.Mode())
			}

			if fi.Size() < 0 {
				t.Errorf("stat '%s': want directory size to be >= 0, got %d", path, fi.Size())
			}

			now := time.Now()
			if now.Sub(fi.ModTime()) > 2*time.Second {
				t.Errorf("stat '%s' : want time to be %s, got %s", path, time.Now(), fi.ModTime())
			}

			name := vfs.Base(dir.Path)
			if fi.Name() != name {
				t.Errorf("stat '%s' : want path to be %s, got %s", path, name, fi.Name())
			}

			curPath := rootDir
			for start, end, i, isLast := 1, 0, 0, false; !isLast; start, i = end+1, i+1 {
				end, isLast = vfsutils.SegmentPath(dir.Path, start)
				part := dir.Path[start:end]
				wantMode := dir.WantModes[i]

				curPath = vfs.Join(curPath, part)
				info, err := vfs.Stat(curPath)
				if err != nil {
					t.Fatalf("stat %s : want error to be nil, got %v", curPath, err)
				}

				wantMode &^= vfs.GetUMask()
				if vfs.OSType() == avfs.OsWindows {
					wantMode = os.ModePerm
				}

				mode := info.Mode() & os.ModePerm
				if wantMode != mode {
					t.Errorf("stat %s %s : want mode to be %s, got %s", path, curPath, wantMode, mode)
				}
			}
		}
	})

	t.Run("MkdirExisting", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			err := vfs.Mkdir(path, dir.Mode)
			if !vfs.IsExist(err) {
				t.Errorf("mkdir %s : want IsExist(err) to be true, got error %v", path, err)
			}
		}
	})

	t.Run("MkdirOnNonExistingDir", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path, "can't", "create", "this")

			err := vfs.Mkdir(path, avfs.DefaultDirPerm)

			switch vfs.OSType() {
			case avfs.OsWindows:
				CheckPathError(t, "Mkdir", "mkdir", path, avfs.ErrWinPathNotFound, err)
			default:
				CheckPathError(t, "Mkdir", "mkdir", path, avfs.ErrNoSuchFileOrDir, err)
			}
		}
	})

	t.Run("MkdirEmptyName", func(t *testing.T) {
		err := vfs.Mkdir("", avfs.DefaultFilePerm)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Mkdir", "mkdir", "", avfs.ErrWinPathNotFound, err)
		default:
			CheckPathError(t, "Mkdir", "mkdir", "", avfs.ErrNoSuchFileOrDir, err)
		}
	})

	t.Run("MkdirOnFile", func(t *testing.T) {
		subDirOnFile := vfs.Join(existingFile, "subDirOnFile")

		err := vfs.Mkdir(subDirOnFile, avfs.DefaultDirPerm)
		CheckPathError(t, "Mkdir", "mkdir", subDirOnFile, avfs.ErrNotADirectory, err)
	})
}

// MkdirAll tests MkdirAll function.
func (sfs *SuiteFS) MkdirAll(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.MkdirAll(rootDir, avfs.DefaultDirPerm)
		CheckPathError(t, "MkdirAll", "mkdir", rootDir, avfs.ErrPermDenied, err)

		return
	}

	existingFile := CreateEmptyFile(t, vfs, rootDir)

	vfs = sfs.GetFsWrite()
	dirs := GetDirsAll()

	t.Run("MkdirAll", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			err := vfs.MkdirAll(path, dir.Mode)
			if err != nil {
				t.Errorf("MkdirAll : want error to be nil, got %v", err)
			}

			fi, err := vfs.Stat(path)
			if err != nil {
				t.Fatalf("stat '%s' : want error to be nil, got %v", path, err)
			}

			if !fi.IsDir() {
				t.Errorf("stat '%s' : want path to be a directory, got mode %s", path, fi.Mode())
			}

			if fi.Size() < 0 {
				t.Errorf("stat '%s': want directory size to be >= 0, got %d", path, fi.Size())
			}

			now := time.Now()
			if now.Sub(fi.ModTime()) > 2*time.Second {
				t.Errorf("stat '%s' : want time to be %s, got %s", path, time.Now(), fi.ModTime())
			}

			name := vfs.Base(dir.Path)
			if fi.Name() != name {
				t.Errorf("stat '%s' : want path to be %s, got %s", path, name, fi.Name())
			}

			want := strings.Count(dir.Path, string(avfs.PathSeparator))
			got := len(dir.WantModes)
			if want != got {
				t.Fatalf("stat %s : want %d directories modes, got %d", path, want, got)
			}

			curPath := rootDir
			for start, end, i, isLast := 1, 0, 0, false; !isLast; start, i = end+1, i+1 {
				end, isLast = vfsutils.SegmentPath(dir.Path, start)
				part := dir.Path[start:end]
				wantMode := dir.WantModes[i]

				curPath = vfs.Join(curPath, part)
				info, err := vfs.Stat(curPath)
				if err != nil {
					t.Fatalf("stat %s : want error to be nil, got %v", curPath, err)
				}

				wantMode &^= vfs.GetUMask()
				if vfs.OSType() == avfs.OsWindows {
					wantMode = os.ModePerm
				}

				mode := info.Mode() & os.ModePerm
				if wantMode != mode {
					t.Errorf("stat %s %s : want mode to be %s, got %s", path, curPath, wantMode, mode)
				}
			}
		}
	})

	t.Run("MkdirAllExistingDir", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			err := vfs.MkdirAll(path, dir.Mode)
			if err != nil {
				t.Errorf("MkdirAll %s : want error to be nil, got error %v", path, err)
			}
		}
	})

	t.Run("MkdirAllOnFile", func(t *testing.T) {
		err := vfs.MkdirAll(existingFile, avfs.DefaultDirPerm)
		CheckPathError(t, "MkdirAll", "mkdir", existingFile, avfs.ErrNotADirectory, err)
	})

	t.Run("MkdirAllSubDirOnFile", func(t *testing.T) {
		subDirOnFile := vfs.Join(existingFile, "subDirOnFile")

		err := vfs.MkdirAll(subDirOnFile, avfs.DefaultDirPerm)
		CheckPathError(t, "MkdirAll", "mkdir", existingFile, avfs.ErrNotADirectory, err)
	})
}

// Open tests Open function.
func (sfs *SuiteFS) Open(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.Open(rootDir)
		CheckPathError(t, "Open", "open", rootDir, avfs.ErrPermDenied, err)

		return
	}

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
		f, err := vfs.Open(existingDir)
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

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		return
	}

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

// ReadDir tests ReadDir function.
func (sfs *SuiteFS) ReadDir(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.ReadDir(rootDir)
		CheckPathError(t, "ReadDir", "open", rootDir, avfs.ErrPermDenied, err)

		return
	}

	rndTree := CreateRndDir(t, vfs, rootDir)
	wDirs := len(rndTree.Dirs)
	wFiles := len(rndTree.Files)
	wSymlinks := len(rndTree.SymLinks)

	existingFile := rndTree.Files[0]

	vfs = sfs.GetFsRead()

	t.Run("ReadDirAll", func(t *testing.T) {
		rdInfos, err := vfs.ReadDir(rootDir)
		if err != nil {
			t.Fatalf("ReadDir : want error to be nil, got %v", err)
		}

		var gDirs, gFiles, gSymlinks int
		for _, rdInfo := range rdInfos {
			mode := rdInfo.Mode()
			switch {
			case mode.IsDir():
				gDirs++
			case mode&os.ModeSymlink != 0:
				gSymlinks++
			default:
				gFiles++
			}
		}

		if wDirs != gDirs {
			t.Errorf("ReadDir : want number of dirs to be %d, got %d", wDirs, gDirs)
		}

		if wFiles != gFiles {
			t.Errorf("ReadDir : want number of files to be %d, got %d", wFiles, gFiles)
		}

		if wSymlinks != gSymlinks {
			t.Errorf("ReadDir : want number of symbolic links to be %d, got %d", wSymlinks, gSymlinks)
		}
	})

	t.Run("ReadDirEmptySubDirs", func(t *testing.T) {
		for _, dir := range rndTree.Dirs {
			dirInfos, err := vfs.ReadDir(dir)
			if err != nil {
				t.Errorf("ReadDir %s : want error to be nil, got %v", dir, err)
			}

			l := len(dirInfos)
			if l != 0 {
				t.Errorf("ReadDir %s : want count to be O, got %d", dir, l)
			}
		}
	})

	t.Run("ReadDirExistingFile", func(t *testing.T) {
		_, err := vfs.ReadDir(existingFile)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "ReadDir", "Readdir", existingFile, avfs.ErrNotADirectory, err)
		default:
			CheckSyscallError(t, "ReadDir", "readdirent", existingFile, avfs.ErrNotADirectory, err)
		}
	})
}

// ReadFile tests ReadFile function.
func (sfs *SuiteFS) ReadFile(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsRead()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.ReadFile(rootDir)
		CheckPathError(t, "ReadFile", "open", rootDir, avfs.ErrPermDenied, err)

		return
	}

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

// Readlink tests Readlink function.
func (sfs *SuiteFS) Readlink(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !sfs.vfsW.HasFeature(avfs.FeatSymlink) {
		_, err := vfs.Readlink(rootDir)
		CheckPathError(t, "Readlink", "readlink", rootDir, avfs.ErrPermDenied, err)

		return
	}

	dirs := CreateDirs(t, vfs, rootDir)
	files := CreateFiles(t, vfs, rootDir)
	symlinks := CreateSymlinks(t, vfs, rootDir)

	vfs = sfs.GetFsRead()

	t.Run("ReadlinkLink", func(t *testing.T) {
		for _, sl := range symlinks {
			oldPath := vfs.Join(rootDir, sl.OldName)
			newPath := vfs.Join(rootDir, sl.NewName)

			gotPath, err := vfs.Readlink(newPath)
			if err != nil {
				t.Errorf("ReadLink %s : want error to be nil, got %v", newPath, err)
			}

			if oldPath != gotPath {
				t.Errorf("ReadLink %s : want link to be %s, got %s", newPath, oldPath, gotPath)
			}
		}
	})

	t.Run("ReadlinkDir", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			_, err := vfs.Readlink(path)
			CheckPathError(t, "ReadLink", "readlink", path, os.ErrInvalid, err)
		}
	})

	t.Run("ReadlinkFile", func(t *testing.T) {
		for _, file := range files {
			path := vfs.Join(rootDir, file.Path)

			_, err := vfs.Readlink(path)
			CheckPathError(t, "ReadLink", "readlink", path, os.ErrInvalid, err)
		}
	})

	t.Run("ReadLinkNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		_, err := vfs.Readlink(nonExistingFile)
		CheckPathError(t, "Readlink", "readlink", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
	})
}

// Remove tests Remove function.
func (sfs *SuiteFS) Remove(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.Remove(rootDir)
		CheckPathError(t, "Remove", "remove", rootDir, avfs.ErrPermDenied, err)

		return
	}

	dirs := CreateDirs(t, vfs, rootDir)
	files := CreateFiles(t, vfs, rootDir)
	symlinks := CreateSymlinks(t, vfs, rootDir)

	if s, ok := vfs.(fmt.Stringer); ok {
		fmt.Println(s.String())
	}

	t.Run("RemoveFile", func(t *testing.T) {
		for _, file := range files {
			path := vfs.Join(rootDir, file.Path)

			_, err := vfs.Stat(path)
			if err != nil {
				t.Fatalf("Stat %s : want error to be nil, got %v", path, err)
			}

			err = vfs.Remove(path)
			if err != nil {
				t.Errorf("Remove %s : want error to be nil, got %v", path, err)
			}

			_, err = vfs.Stat(path)

			switch vfs.OSType() {
			case avfs.OsWindows:
				CheckPathError(t, "Stat", "CreateFile", path, avfs.ErrNoSuchFileOrDir, err)
			default:
				CheckPathError(t, "Stat", "stat", path, avfs.ErrNoSuchFileOrDir, err)
			}
		}
	})

	t.Run("RemoveDir", func(t *testing.T) {
		if s, ok := vfs.(fmt.Stringer); ok {
			fmt.Println(s.String())
		}

		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			dirInfos, err := vfs.ReadDir(path)
			if err != nil {
				t.Fatalf("ReadDir %s : want error to be nil, got %v", path, err)
			}

			err = vfs.Remove(path)

			isLeaf := len(dirInfos) == 0
			if isLeaf {
				if err != nil {
					t.Errorf("Remove %s : want error to be nil, got %v", path, err)
				}

				_, err = vfs.Stat(path)
				CheckPathError(t, "Stat", "stat", path, avfs.ErrNoSuchFileOrDir, err)
			} else {
				CheckPathError(t, "Remove", "remove", path, avfs.ErrDirNotEmpty, err)

				_, err = vfs.Stat(path)
				if err != nil {
					t.Errorf("Remove %s : want error to be nil, got %v", path, err)
				}
			}
		}
	})

	t.Run("RemoveSymlinks", func(t *testing.T) {
		for _, sl := range symlinks {
			newPath := vfs.Join(rootDir, sl.NewName)

			err := vfs.Remove(newPath)
			if err != nil {
				t.Errorf("Remove %s : want error to be nil, got %v", newPath, err)
			}

			_, err = vfs.Stat(newPath)
			CheckPathError(t, "Stat", "stat", newPath, avfs.ErrNoSuchFileOrDir, err)
		}
	})

	t.Run("RemoveNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		err := vfs.Remove(nonExistingFile)
		CheckPathError(t, "Remove", "remove", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
	})
}

// RemoveAll tests RemoveAll function.
func (sfs *SuiteFS) RemoveAll(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.RemoveAll(rootDir)
		CheckPathError(t, "RemoveAll", "removeall", rootDir, avfs.ErrPermDenied, err)

		return
	}

	baseDir := vfs.Join(rootDir, "RemoveAll")
	dirs := CreateDirs(t, vfs, baseDir)
	files := CreateFiles(t, vfs, baseDir)
	symlinks := CreateSymlinks(t, vfs, baseDir)

	t.Run("RemoveAll", func(t *testing.T) {
		err := vfs.RemoveAll(baseDir)
		if err != nil {
			t.Fatalf("RemoveAll %s : want error to be nil, got %v", baseDir, err)
		}

		for _, dir := range dirs {
			path := vfs.Join(baseDir, dir.Path)

			_, err = vfs.Stat(path)
			CheckPathError(t, "Stat", "stat", path, avfs.ErrNoSuchFileOrDir, err)
		}

		for _, file := range files {
			path := vfs.Join(baseDir, file.Path)

			_, err = vfs.Stat(path)
			CheckPathError(t, "Stat", "stat", path, avfs.ErrNoSuchFileOrDir, err)
		}

		for _, sl := range symlinks {
			path := vfs.Join(baseDir, sl.NewName)

			_, err = vfs.Stat(path)
			CheckPathError(t, "Stat", "stat", path, avfs.ErrNoSuchFileOrDir, err)
		}

		_, err = vfs.Stat(baseDir)
		CheckPathError(t, "Stat", "stat", baseDir, avfs.ErrNoSuchFileOrDir, err)
	})

	t.Run("RemoveAllOneFile", func(t *testing.T) {
		err := vfs.MkdirAll(baseDir, avfs.DefaultDirPerm)
		if err != nil {
			t.Fatalf("Mkdir %s : want error to be nil, got %v", baseDir, err)
		}

		existingFile := CreateEmptyFile(t, vfs, rootDir)

		err = vfs.RemoveAll(existingFile)
		if err != nil {
			t.Errorf("RemoveAll %s : want error to be nil, got %v", existingFile, err)
		}
	})

	t.Run("RemoveAllPathEmpty", func(t *testing.T) {
		CreateDirs(t, vfs, baseDir)

		err := vfs.Chdir(baseDir)
		if err != nil {
			t.Fatalf("Chdir %s : want error to be nil, got %v", baseDir, err)
		}

		err = vfs.RemoveAll("")
		if err != nil {
			t.Errorf("RemoveAll '' : want error to be nil, got %v", err)
		}

		// Verify that nothing was removed.
		for _, dir := range dirs {
			path := vfs.Join(baseDir, dir.Path)

			_, err = vfs.Stat(path)
			if err != nil {
				t.Fatalf("RemoveAll %s : want error to be nil, got %v", path, err)
			}
		}
	})

	t.Run("RemoveAllNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		err := vfs.RemoveAll(nonExistingFile)
		if err != nil {
			t.Errorf("RemoveAll %s : want error to be nil, got %v", nonExistingFile, err)
		}
	})
}

// Rename tests Rename function.
func (sfs *SuiteFS) Rename(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.Rename(rootDir, rootDir)
		CheckLinkError(t, "Rename", "rename", rootDir, rootDir, avfs.ErrPermDenied, err)

		return
	}

	t.Run("RenameDir", func(t *testing.T) {
		dirs := CreateDirs(t, vfs, rootDir)

		for i := len(dirs) - 1; i >= 0; i-- {
			oldPath := vfs.Join(rootDir, dirs[i].Path)
			newPath := oldPath + "New"

			err := vfs.Rename(oldPath, newPath)
			if err != nil {
				t.Errorf("Rename %s %s : want error to be nil, got %v", oldPath, newPath, err)
			}

			_, err = vfs.Stat(oldPath)

			switch vfs.OSType() {
			case avfs.OsWindows:
				CheckPathError(t, "Stat", "CreateFile", oldPath, avfs.ErrNoSuchFileOrDir, err)
			default:
				CheckPathError(t, "Stat", "stat", oldPath, avfs.ErrNoSuchFileOrDir, err)
			}

			_, err = vfs.Stat(newPath)
			if err != nil {
				t.Errorf("Stat %s : want error to be nil, got %v", newPath, err)
			}
		}
	})

	t.Run("RenameFile", func(t *testing.T) {
		CreateDirs(t, vfs, rootDir)
		files := CreateFiles(t, vfs, rootDir)

		for _, file := range files {
			oldPath := vfs.Join(rootDir, file.Path)
			newPath := vfs.Join(rootDir, vfs.Base(oldPath))

			err := vfs.Rename(oldPath, newPath)
			if err != nil {
				t.Errorf("Rename %s %s : want error to be nil, got %v", oldPath, newPath, err)
			}

			_, err = vfs.Stat(oldPath)

			switch {
			case oldPath == newPath:
				if err != nil {
					t.Errorf("Stat %s : want error to be nil, got %v", oldPath, err)
				}
			default:

				switch vfs.OSType() {
				case avfs.OsWindows:
					CheckPathError(t, "Stat", "CreateFile", oldPath, avfs.ErrNoSuchFileOrDir, err)
				default:
					CheckPathError(t, "Stat", "stat", oldPath, avfs.ErrNoSuchFileOrDir, err)
				}
			}

			_, err = vfs.Stat(newPath)
			if err != nil {
				t.Errorf("Stat %s : want error to be nil, got %v", newPath, err)
			}
		}
	})

	t.Run("RenameNonExistingFile", func(t *testing.T) {
		srcNonExistingFile := vfs.Join(rootDir, "srcNonExistingFile1")
		dstNonExistingFile := vfs.Join(rootDir, "dstNonExistingFile1")

		err := vfs.Rename(srcNonExistingFile, dstNonExistingFile)
		CheckLinkError(t, "Rename", "rename", srcNonExistingFile, dstNonExistingFile, avfs.ErrNoSuchFileOrDir, err)
	})

	t.Run("RenameDirToExistingDir", func(t *testing.T) {
		srcExistingDir := vfs.Join(rootDir, "srcExistingDir2")
		dstExistingDir := vfs.Join(rootDir, "dstExistingDir2")

		err := vfs.Mkdir(srcExistingDir, avfs.DefaultDirPerm)
		if err != nil {
			t.Fatalf("Mkdir : want error to be nil, got %v", err)
		}

		err = vfs.Mkdir(dstExistingDir, avfs.DefaultDirPerm)
		if err != nil {
			t.Fatalf("Mkdir : want error to be nil, got %v", err)
		}

		err = vfs.Rename(srcExistingDir, dstExistingDir)
		CheckLinkError(t, "Rename", "rename", srcExistingDir, dstExistingDir, avfs.ErrFileExists, err)
	})

	t.Run("RenameFileToExistingFile", func(t *testing.T) {
		srcExistingFile := vfs.Join(rootDir, "srcExistingFile3")
		dstExistingFile := vfs.Join(rootDir, "dstExistingFile3")
		data := []byte("data")

		err := vfs.WriteFile(srcExistingFile, data, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		err = vfs.WriteFile(dstExistingFile, nil, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		err = vfs.Rename(srcExistingFile, dstExistingFile)
		if err != nil {
			t.Errorf("Rename : want error to be nil, got %v", err)
		}

		_, err = vfs.Stat(srcExistingFile)
		CheckPathError(t, "Stat", "stat", srcExistingFile, avfs.ErrNoSuchFileOrDir, err)

		info, err := vfs.Stat(dstExistingFile)
		if err != nil {
			t.Errorf("Stat : want error to be nil, got %v", err)
		}

		if int(info.Size()) != len(data) {
			t.Errorf("Stat : want size to be %d, got %d", len(data), info.Size())
		}
	})

	t.Run("RenameFileToExistingDir", func(t *testing.T) {
		srcExistingFile := vfs.Join(rootDir, "srcExistingFile4")
		dstExistingDir := vfs.Join(rootDir, "dstExistingDir4")

		err := vfs.WriteFile(srcExistingFile, nil, avfs.DefaultFilePerm)
		if err != nil {
			t.Fatalf("WriteFile : want error to be nil, got %v", err)
		}

		err = vfs.Mkdir(dstExistingDir, avfs.DefaultDirPerm)
		if err != nil {
			t.Fatalf("Mkdir : want error to be nil, got %v", err)
		}

		err = vfs.Rename(srcExistingFile, dstExistingDir)
		CheckLinkError(t, "Rename", "rename", srcExistingFile, dstExistingDir, avfs.ErrFileExists, err)
	})
}

// SameFile tests SameFile function.
func (sfs *SuiteFS) SameFile(t *testing.T) {
	rootDir1, removeDir1 := sfs.CreateRootDir(t, UsrTest)
	defer removeDir1()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		if vfs.SameFile(nil, nil) {
			t.Errorf("SameFile : want SameFile to be false, got true")
		}

		return
	}

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

// Stat tests Stat function.
func (sfs *SuiteFS) Stat(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.Stat(rootDir)
		CheckPathError(t, "Stat", "stat", rootDir, avfs.ErrPermDenied, err)

		return
	}

	dirs := CreateDirs(t, vfs, rootDir)
	files := CreateFiles(t, vfs, rootDir)
	_ = CreateSymlinks(t, vfs, rootDir)

	vfs = sfs.GetFsRead()

	t.Run("StatDir", func(t *testing.T) {
		for _, dir := range dirs {
			path := vfs.Join(rootDir, dir.Path)

			info, err := vfs.Stat(path)
			if err != nil {
				t.Errorf("Stat %s : want error to be nil, got %v", path, err)

				continue
			}

			if vfs.Base(path) != info.Name() {
				t.Errorf("Stat %s : want name to be %s, got %s", path, vfs.Base(path), info.Name())
			}

			wantMode := (dir.Mode | os.ModeDir) &^ vfs.GetUMask()
			if vfs.OSType() == avfs.OsWindows {
				wantMode = os.ModeDir | os.ModePerm
			}

			if wantMode != info.Mode() {
				t.Errorf("Stat %s : want mode to be %s, got %s", path, wantMode, info.Mode())
			}
		}
	})

	t.Run("StatFile", func(t *testing.T) {
		for _, file := range files {
			path := vfs.Join(rootDir, file.Path)

			info, err := vfs.Stat(path)
			if err != nil {
				t.Errorf("Stat %s : want error to be nil, got %v", path, err)

				continue
			}

			if info.Name() != vfs.Base(path) {
				t.Errorf("Stat %s : want name to be %s, got %s", path, vfs.Base(path), info.Name())
			}

			wantMode := file.Mode &^ vfs.GetUMask()
			if vfs.OSType() == avfs.OsWindows {
				wantMode = 0o666
			}

			if wantMode != info.Mode() {
				t.Errorf("Stat %s : want mode to be %s, got %s", path, wantMode, info.Mode())
			}

			wantSize := int64(len(file.Content))
			if wantSize != info.Size() {
				t.Errorf("Lstat %s : want size to be %d, got %d", path, wantSize, info.Size())
			}
		}
	})

	t.Run("StatSymlink", func(t *testing.T) {
		for _, sl := range GetSymlinksEval(vfs) {
			newPath := vfs.Join(rootDir, sl.NewName)
			oldPath := vfs.Join(rootDir, sl.OldName)

			info, err := vfs.Stat(newPath)
			if err != nil {
				if sl.WantErr == nil {
					t.Errorf("Stat %s : want error to be nil, got %v", newPath, err)
				}
				CheckPathError(t, "Lstat", "stat", newPath, sl.WantErr, err)

				continue
			}

			var (
				wantName string
				wantMode os.FileMode
			)

			if sl.IsSymlink {
				wantName = vfs.Base(newPath)
			} else {
				wantName = vfs.Base(oldPath)
			}

			wantMode = sl.Mode
			if wantName != info.Name() {
				t.Errorf("Stat %s : want name to be %s, got %s", newPath, wantName, info.Name())
			}

			if wantMode != info.Mode() {
				t.Errorf("Stat %s : want mode to be %s, got %s", newPath, wantMode, info.Mode())
			}
		}
	})

	t.Run("StatNonExistingFile", func(t *testing.T) {
		nonExistingFile := vfs.Join(rootDir, "nonExistingFile")

		_, err := vfs.Stat(nonExistingFile)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Stat", "CreateFile", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
		default:
			CheckPathError(t, "Stat", "stat", nonExistingFile, avfs.ErrNoSuchFileOrDir, err)
		}
	})

	t.Run("StatsubDirOnFile", func(t *testing.T) {
		subDirOnFile := vfs.Join(rootDir, files[0].Path, "subDirOnFile")

		_, err := vfs.Stat(subDirOnFile)

		switch vfs.OSType() {
		case avfs.OsWindows:
			CheckPathError(t, "Stat", "CreateFile", subDirOnFile, avfs.ErrNotADirectory, err)
		default:
			CheckPathError(t, "Stat", "stat", subDirOnFile, avfs.ErrNotADirectory, err)
		}
	})
}

// Symlink tests Symlink function.
func (sfs *SuiteFS) Symlink(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatSymlink) {
		err := vfs.Symlink(rootDir, rootDir)
		CheckLinkError(t, "Symlink", "symlink", rootDir, rootDir, avfs.ErrPermDenied, err)

		return
	}

	_ = CreateDirs(t, vfs, rootDir)
	_ = CreateFiles(t, vfs, rootDir)

	t.Run("Symlink", func(t *testing.T) {
		symlinks := GetSymlinks(vfs)
		for _, sl := range symlinks {
			oldPath := vfs.Join(rootDir, sl.OldName)
			newPath := vfs.Join(rootDir, sl.NewName)

			err := vfs.Symlink(oldPath, newPath)
			if err != nil {
				t.Errorf("Symlink %s %s : want error to be nil, got %v", oldPath, newPath, err)
			}

			gotPath, err := vfs.Readlink(newPath)
			if err != nil {
				t.Errorf("ReadLink %s : want error to be nil, got %v", newPath, err)
			}

			if oldPath != gotPath {
				t.Errorf("ReadLink %s : want link to be %s, got %s", newPath, oldPath, gotPath)
			}
		}
	})
}

// TempDir tests TempDir function.
func (sfs *SuiteFS) TempDir(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.TempDir(rootDir, "")
		if err.(*os.PathError).Err != avfs.ErrPermDenied {
			t.Errorf("TempDir : want error to be %v, got %v", avfs.ErrPermDenied, err)
		}

		return
	}

	existingFile := CreateEmptyFile(t, vfs, rootDir)

	t.Run("TempDirOnFile", func(t *testing.T) {
		_, err := vfs.TempDir(existingFile, "")

		e, ok := err.(*os.PathError)
		if !ok {
			t.Fatalf("TempDir : want error type *os.PathError, got %v", reflect.TypeOf(err))
		}

		const op = "mkdir"
		wantErr := avfs.ErrNotADirectory
		if e.Op != op || vfs.Dir(e.Path) != existingFile || e.Err != wantErr {
			wantPathErr := &os.PathError{Op: op, Path: existingFile + "/<random number>", Err: wantErr}
			t.Errorf("TempDir : want error to be %v, got %v", wantPathErr, err)
		}
	})
}

// TempFile tests TempFile function.
func (sfs *SuiteFS) TempFile(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		_, err := vfs.TempFile(rootDir, "")
		if err.(*os.PathError).Err != avfs.ErrPermDenied {
			t.Errorf("TempFile : want error to be %v, got %v", avfs.ErrPermDenied, err)
		}

		return
	}

	existingFile := CreateEmptyFile(t, vfs, rootDir)

	t.Run("TempFileOnFile", func(t *testing.T) {
		_, err := vfs.TempFile(existingFile, "")

		e, ok := err.(*os.PathError)
		if !ok {
			t.Fatalf("TempFile : want error type *os.PathError, got %v", reflect.TypeOf(err))
		}

		const op = "open"
		wantErr := avfs.ErrNotADirectory
		if e.Op != op || vfs.Dir(e.Path) != existingFile || e.Err != wantErr {
			wantPathErr := &os.PathError{Op: op, Path: existingFile + "/<random number>", Err: wantErr}
			t.Errorf("TempDir : want error to be %v, got %v", wantPathErr, err)
		}
	})
}

// Truncate tests Truncate function.
func (sfs *SuiteFS) Truncate(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.Truncate(rootDir, 0)
		CheckPathError(t, "Truncate", "truncate", rootDir, avfs.ErrPermDenied, err)

		return
	}
}

// Umask tests UMask and GetUMask functions.
func (sfs *SuiteFS) Umask(t *testing.T) {
	const umaskTest = 0o077

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		vfs.UMask(0)

		if um := vfs.GetUMask(); um != 0 {
			t.Errorf("GetUMask : want umask to be 0, got %d", um)
		}

		return
	}

	umaskStart := vfs.GetUMask()
	vfs.UMask(umaskTest)

	u := vfs.GetUMask()
	if u != umaskTest {
		t.Errorf("umaskTest : want umask to be %o, got %o", umaskTest, u)
	}

	vfs.UMask(umaskStart)

	u = vfs.GetUMask()
	if u != umaskStart {
		t.Errorf("umaskTest : want umask to be %o, got %o", umaskStart, u)
	}
}

// WriteFile tests WriteFile function.
func (sfs *SuiteFS) WriteFile(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		err := vfs.WriteFile(rootDir, []byte{0}, avfs.DefaultFilePerm)
		CheckPathError(t, "WriteFile", "open", rootDir, avfs.ErrPermDenied, err)

		return
	}

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

// WriteOnReadOnly tests all write functions of a read only file system.
func (sfs *SuiteFS) WriteOnReadOnly(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	existingFile := vfs.Join(rootDir, "existingFile")

	err := vfs.WriteFile(existingFile, nil, avfs.DefaultFilePerm)
	if err != nil {
		t.Fatalf("WriteFile : want error to be nil, got %v", err)
	}

	newFile := vfs.Join(existingFile, "newFile")

	vfs = sfs.GetFsRead()
	if !vfs.HasFeature(avfs.FeatReadOnly) {
		t.Errorf("HasFeature : want read only file system")
	}

	t.Run("ReadOnlyFs", func(t *testing.T) {
		err := vfs.Chmod(existingFile, avfs.DefaultFilePerm)
		CheckPathError(t, "Chmod", "chmod", existingFile, avfs.ErrPermDenied, err)

		err = vfs.Chown(existingFile, 0, 0)
		CheckPathError(t, "Chown", "chown", existingFile, avfs.ErrPermDenied, err)

		err = vfs.Chroot(rootDir)
		CheckPathError(t, "Chroot", "chroot", rootDir, avfs.ErrPermDenied, err)

		err = vfs.Chtimes(existingFile, time.Now(), time.Now())
		CheckPathError(t, "Chtimes", "chtimes", existingFile, avfs.ErrPermDenied, err)

		_, err = vfs.Create(newFile)
		CheckPathError(t, "Create", "open", newFile, avfs.ErrPermDenied, err)

		err = vfs.Lchown(existingFile, 0, 0)
		CheckPathError(t, "Lchown", "lchown", existingFile, avfs.ErrPermDenied, err)

		err = vfs.Link(existingFile, newFile)
		CheckLinkError(t, "Link", "link", existingFile, newFile, avfs.ErrPermDenied, err)

		err = vfs.Mkdir(newFile, avfs.DefaultDirPerm)
		CheckPathError(t, "Mkdir", "mkdir", newFile, avfs.ErrPermDenied, err)

		err = vfs.MkdirAll(newFile, avfs.DefaultDirPerm)
		CheckPathError(t, "MkdirAll", "mkdir", newFile, avfs.ErrPermDenied, err)

		_, err = vfs.OpenFile(newFile, os.O_RDWR, avfs.DefaultFilePerm)
		CheckPathError(t, "OpenFile", "open", newFile, avfs.ErrPermDenied, err)

		err = vfs.Remove(existingFile)
		CheckPathError(t, "Remove", "remove", existingFile, avfs.ErrPermDenied, err)

		err = vfs.RemoveAll(existingFile)
		CheckPathError(t, "RemoveAll", "removeall", existingFile, avfs.ErrPermDenied, err)

		err = vfs.Rename(existingFile, newFile)
		CheckLinkError(t, "Rename", "rename", existingFile, newFile, avfs.ErrPermDenied, err)

		err = vfs.Symlink(existingFile, newFile)
		CheckLinkError(t, "Symlink", "symlink", existingFile, newFile, avfs.ErrPermDenied, err)

		_, err = vfs.TempDir(rootDir, "")
		if err.(*os.PathError).Err != avfs.ErrPermDenied {
			t.Errorf("TempDir : want error to be %v, got %v", avfs.ErrPermDenied, err)
		}

		_, err = vfs.TempFile(rootDir, "")
		if err.(*os.PathError).Err != avfs.ErrPermDenied {
			t.Errorf("TempFile : want error to be %v, got %v", avfs.ErrPermDenied, err)
		}

		err = vfs.Truncate(existingFile, 0)
		CheckPathError(t, "Truncate", "truncate", existingFile, avfs.ErrPermDenied, err)

		err = vfs.WriteFile(newFile, []byte{0}, avfs.DefaultFilePerm)
		CheckPathError(t, "WriteFile", "open", newFile, avfs.ErrPermDenied, err)
	})

	t.Run("ReadOnlyFile", func(t *testing.T) {
		f, err := vfs.Open(existingFile)
		if err != nil {
			t.Fatalf("Open : want error to be nil, got %v", err)
		}

		err = f.Chmod(0o777)
		CheckPathError(t, "Chmod", "chmod", f.Name(), avfs.ErrPermDenied, err)

		err = f.Chown(0, 0)
		CheckPathError(t, "Chown", "chown", f.Name(), avfs.ErrPermDenied, err)

		err = f.Truncate(0)
		CheckPathError(t, "Truncate", "truncate", f.Name(), avfs.ErrPermDenied, err)

		_, err = f.Write([]byte{})
		CheckPathError(t, "Write", "write", f.Name(), avfs.ErrPermDenied, err)

		_, err = f.WriteAt([]byte{}, 0)
		CheckPathError(t, "WriteAt", "write", f.Name(), avfs.ErrPermDenied, err)

		_, err = f.WriteString("")
		CheckPathError(t, "WriteString", "write", f.Name(), avfs.ErrPermDenied, err)
	})
}

// WriteString tests WriteString function.
func (sfs *SuiteFS) WriteString(t *testing.T) {
	rootDir, removeDir := sfs.CreateRootDir(t, UsrTest)
	defer removeDir()

	vfs := sfs.GetFsWrite()

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		return
	}

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