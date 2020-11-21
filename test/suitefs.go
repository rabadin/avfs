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
	"os"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/fsutil"
)

// SuiteFs is a test suite for file systems.
type SuiteFs struct {
	// t is passed to Test functions to manage test state and support formatted test logs.
	t *testing.T

	// vfsRoot is the file system as a root user.
	vfsRoot avfs.Fs

	// vfsW is the file system as test user with read and write permissions.
	vfsW avfs.Fs

	// vfsR is the file system as test user with read only permissions.
	vfsR avfs.Fs

	// rootDir is the root directory for tests, it can be generated automatically or specified with WithRootDir().
	rootDir string

	// maxRace is the maximum number of concurrent goroutines used in race tests.
	maxRace int

	// canTestPerm indicates if permissions can be tested.
	canTestPerm bool

	// osType is the operating system of the filesystem te test.
	osType avfs.OSType

	// Groups contains the test groups created with the identity manager.
	Groups []avfs.GroupReader

	// Users contains the test users created with the identity manager.
	Users []avfs.UserReader
}

// Option defines the option function used for initializing SuiteFs.
type Option func(*SuiteFs)

// NewSuiteFs creates a new test suite for a file system.
func NewSuiteFs(t *testing.T, vfsRoot avfs.Fs, opts ...Option) *SuiteFs {
	if vfsRoot == nil {
		t.Fatal("New : want FsRoot to be set, got nil")
	}

	currentUser := vfsRoot.CurrentUser()
	canTestPerm := vfsRoot.HasFeature(avfs.FeatIdentityMgr) && currentUser.IsRoot()

	info := "Info vfs : type = " + vfsRoot.Type()
	if vfsRoot.Name() != "" {
		info += ", name = " + vfsRoot.Name()
	}

	t.Log(info)

	sfs := &SuiteFs{
		t:           t,
		vfsRoot:     vfsRoot,
		vfsW:        vfsRoot,
		vfsR:        vfsRoot,
		rootDir:     "",
		maxRace:     1000,
		canTestPerm: canTestPerm,
		osType:      fsutil.RunTimeOS(),
	}

	if canTestPerm {
		sfs.Groups = CreateGroups(t, vfsRoot, "")
		sfs.Users = CreateUsers(t, vfsRoot, "")

		_, err := vfsRoot.User(UsrTest)
		if err != nil {
			t.Fatalf("User %s : want error to be nil, got %s", UsrTest, err)
		}
	}

	for _, opt := range opts {
		opt(sfs)
	}

	return sfs
}

// Options

// WithRootDir returns an option function which sets the root directory for the tests.
func WithRootDir(rootDir string) Option {
	return func(sfs *SuiteFs) {
		sfs.rootDir = rootDir
	}
}

// WithOs returns an option function which sets the operating system for the tests.
func WithOs(osType avfs.OSType) Option {
	return func(sfs *SuiteFs) {
		sfs.osType = osType
	}
}

// OS returns the operating system, real or simulated.
func (sfs *SuiteFs) OS() avfs.OSType {
	return sfs.osType
}

// GetFsAsUser sets the test user to userName.
func (sfs *SuiteFs) GetFsAsUser(name string) (avfs.Fs, avfs.UserReader) {
	vfs := sfs.vfsRoot

	u := vfs.CurrentUser()
	if !sfs.canTestPerm || u.Name() == name {
		return vfs, u
	}

	u, err := vfs.User(name)
	if err != nil {
		sfs.t.Fatalf("User %s : want error to be nil, got %s", name, err)
	}

	return vfs, u
}

// GetFsRead returns the file system for read only functions.
func (sfs *SuiteFs) GetFsRead() avfs.Fs {
	return sfs.vfsR
}

// GetFsRoot return the root file system.
func (sfs *SuiteFs) GetFsRoot() avfs.Fs {
	return sfs.vfsRoot
}

// GetFsWrite returns the file system for read and write functions.
func (sfs *SuiteFs) GetFsWrite() avfs.Fs {
	return sfs.vfsW
}

// FsRead sets the file system for read functions.
func (sfs *SuiteFs) FsRead(fsR avfs.Fs) {
	if fsR == nil {
		sfs.t.Fatal("SuiteFs : want FsR to be set, got nil")
	}

	sfs.vfsR = fsR
}

// FsWrite sets the file system for write functions.
func (sfs *SuiteFs) FsWrite(fsW avfs.Fs) {
	if fsW == nil {
		sfs.t.Fatal("SuiteFs : want FsW to be set, got nil")
	}

	sfs.vfsW = fsW
}

// CreateRootDir creates the root directory for the tests.
// Each test have its own directory in /tmp/avfs.../
// this directory and its descendants are removed by removeDir() function.
func (sfs *SuiteFs) CreateRootDir(userName string) (t *testing.T, rootDir string, removeDir func()) {
	t = sfs.t

	vfs, _ := sfs.GetFsAsUser(userName)

	if !vfs.HasFeature(avfs.FeatBasicFs) {
		return t, avfs.TmpDir, func() {}
	}

	var err error
	if sfs.rootDir == "" {
		rootDir, err = vfs.TempDir("", avfs.Avfs)
		if err != nil {
			t.Fatalf("TempDir : want error to be nil, got %s", err)
		}
	} else {
		rootDir = sfs.rootDir

		err = vfs.MkdirAll(rootDir, 0o700)
		if err != nil {
			t.Fatalf("MkdirAll : want error to be nil, got %s", err)
		}
	}

	if fsutil.RunTimeOS() == avfs.OsDarwin {
		rootDir, _ = vfs.EvalSymlinks(rootDir)
	}

	_, err = vfs.Stat(rootDir)
	if err != nil {
		t.Fatalf("Stat : want error to be nil, got %s", err)
	}

	// default permissions on rootDir are 0o700 (generated by TempDir())
	// make rootDir accessible by anyone.
	err = vfs.Chmod(rootDir, 0o777)
	if err != nil {
		t.Fatalf("Chmod : want error to be nil, got %s", err)
	}

	err = vfs.Chdir(rootDir)
	if err != nil {
		t.Fatalf("Chdir : want error to be nil, got %s", err)
	}

	removeDir = func() {
		vfs, _ := sfs.GetFsAsUser(avfs.UsrRoot)

		err = vfs.RemoveAll(rootDir)
		if err != nil && sfs.osType != avfs.OsWindows {
			t.Logf("RemoveAll : want error to be nil, got %s", err)
		}
	}

	return t, rootDir, removeDir
}

// All runs all file systems tests.
func (sfs *SuiteFs) All() {
	sfs.Read()
	sfs.Write()
	sfs.Path()
}

// Write runs all file systems tests with write access.
func (sfs *SuiteFs) Write() {
	sfs.Chtimes()
	sfs.DirFuncOnFile()
	sfs.FileReadEdgeCases()
	sfs.FileWrite()
	sfs.FileWriteEdgeCases()
	sfs.FileWriteTime()
	sfs.FileCloseWrite()
	sfs.FuncNonExistingFile()
	sfs.Link()
	sfs.Mkdir()
	sfs.MkdirAll()
	sfs.OpenFileWrite()
	sfs.Remove()
	sfs.RemoveAll()
	sfs.RemoveAllEdgeCases()
	sfs.Rename()
	sfs.SameFile()
	sfs.Symlink()
	sfs.WriteFile()
	sfs.WriteString()
	sfs.Umask()
}

// Read runs all file systems tests with read access.
func (sfs *SuiteFs) Read() {
	sfs.Chdir()
	sfs.EvalSymlink()
	sfs.FileFuncOnClosed()
	sfs.FileFuncOnDir()
	sfs.FileRead()
	sfs.FileSeek()
	sfs.FileCloseRead()
	sfs.FileTruncate()
	sfs.GetTempDir()
	sfs.Glob()
	sfs.Lstat()
	sfs.NotImplemented()
	sfs.OpenFileRead()
	sfs.ReadDir()
	sfs.ReadDirNames()
	sfs.ReadFile()
	sfs.Readlink()
	sfs.Stat()
}

// Dir contains the data to test directories.
type Dir struct {
	Path      string
	Mode      os.FileMode
	WantModes []os.FileMode
}

// GetDirs returns the test directories used by Mkdir function.
func GetDirs() []*Dir {
	dirs := []*Dir{
		{Path: "/A", Mode: 0o777, WantModes: []os.FileMode{0o777}},
		{Path: "/B", Mode: 0o755, WantModes: []os.FileMode{0o755}},
		{Path: "/B/1", Mode: 0o755, WantModes: []os.FileMode{0o755, 0o755}},
		{Path: "/B/1/D", Mode: 0o700, WantModes: []os.FileMode{0o755, 0o755, 0o700}},
		{Path: "/B/1/E", Mode: 0o755, WantModes: []os.FileMode{0o755, 0o755, 0o755}},
		{Path: "/B/2", Mode: 0o750, WantModes: []os.FileMode{0o755, 0o750}},
		{Path: "/B/2/F", Mode: 0o755, WantModes: []os.FileMode{0o755, 0o750, 0o755}},
		{Path: "/B/2/F/3", Mode: 0o755, WantModes: []os.FileMode{0o755, 0o750, 0o755, 0o755}},
		{Path: "/B/2/F/3/G", Mode: 0o777, WantModes: []os.FileMode{0o755, 0o750, 0o755, 0o755, 0o777}},
		{Path: "/B/2/F/3/G/4", Mode: 0o777, WantModes: []os.FileMode{0o755, 0o750, 0o755, 0o755, 0o777, 0o777}},
		{Path: "/C", Mode: 0o750, WantModes: []os.FileMode{0o750}},
		{Path: "/C/5", Mode: 0o750, WantModes: []os.FileMode{0o750, 0o750}},
	}

	return dirs
}

// GetDirsAll returns the test directories used by MkdirAll function.
func GetDirsAll() []*Dir {
	dirs := []*Dir{
		{Path: "/H/6", Mode: 0o750, WantModes: []os.FileMode{0o750, 0o750}},
		{Path: "/H/6/I/7", Mode: 0o755, WantModes: []os.FileMode{0o750, 0o750, 0o755, 0o755}},
		{Path: "/H/6/I/7/J/8", Mode: 0o777, WantModes: []os.FileMode{0o750, 0o750, 0o755, 0o755, 0o777, 0o777}},
	}

	return dirs
}

// CreateDirs create test directories.
func CreateDirs(t *testing.T, vfs avfs.Fs, rootDir string) []*Dir {
	dirs := GetDirs()
	for _, dir := range dirs {
		path := vfs.Join(rootDir, dir.Path)

		err := vfs.Mkdir(path, dir.Mode)
		if err != nil {
			t.Fatalf("Mkdir %s : want error to be nil, got %v", path, err)
		}
	}

	return dirs
}

// File contains the data to test files.
type File struct {
	Path    string
	Mode    os.FileMode
	Content []byte
}

// GetFiles returns the test files.
func GetFiles() []*File {
	files := []*File{
		{Path: "/file.txt", Mode: avfs.DefaultFilePerm, Content: []byte("file")},
		{Path: "/A/afile1.txt", Mode: 0o777, Content: []byte("afile1")},
		{Path: "/A/afile2.txt", Mode: avfs.DefaultFilePerm, Content: []byte("afile2")},
		{Path: "/A/afile3.txt", Mode: 0o600, Content: []byte("afile3")},
		{Path: "/B/1/1file.txt", Mode: avfs.DefaultFilePerm, Content: []byte("1file")},
		{Path: "/B/1/E/efile.txt", Mode: avfs.DefaultFilePerm, Content: []byte("efile")},
		{Path: "/B/2/F/3/3file1.txt", Mode: 0o640, Content: []byte("3file1")},
		{Path: "/B/2/F/3/3file2.txt", Mode: avfs.DefaultFilePerm, Content: []byte("3file2")},
		{Path: "/B/2/F/3/G/4/4file.txt", Mode: avfs.DefaultFilePerm, Content: []byte("4file")},
		{Path: "/C/cfile.txt", Mode: avfs.DefaultFilePerm, Content: []byte("cfile")},
	}

	return files
}

// CreateFiles create test files.
func CreateFiles(t *testing.T, vfs avfs.Fs, rootDir string) []*File {
	files := GetFiles()
	for _, file := range files {
		path := vfs.Join(rootDir, file.Path)

		err := vfs.WriteFile(path, file.Content, file.Mode)
		if err != nil {
			t.Fatalf("WriteFile %s : want error to be nil, got %v", path, err)
		}
	}

	return files
}

// Symlink contains the data to create symbolic links.
type Symlink struct {
	NewName string
	OldName string
}

// GetSymlinks returns the symbolic links to create.
func GetSymlinks(vfs avfs.Featurer) []*Symlink {
	if !vfs.HasFeature(avfs.FeatSymlink) {
		return nil
	}

	symlinks := []*Symlink{
		{NewName: "/A/lroot", OldName: "/"},
		{NewName: "/lC", OldName: "/C"},
		{NewName: "/B/1/lafile2.txt", OldName: "/A/afile2.txt"},
		{NewName: "/B/2/lf", OldName: "/B/2/F"},
		{NewName: "/B/2/F/3/llf", OldName: "/B/2/lf"},
		{NewName: "/C/lllf", OldName: "/B/2/F/3/llf"},
		{NewName: "/A/l3file2.txt", OldName: "/C/lllf/3/3file2.txt"},
		{NewName: "/C/lNonExist", OldName: "/A/path/to/a/non/existing/file"},
	}

	return symlinks
}

// SymlinkEval contains the data to evaluate symbolic links.
type SymlinkEval struct {
	NewName   string
	OldName   string
	WantErr   error
	IsSymlink bool
	Mode      os.FileMode
}

// GetSymlinksEval return the symbolic links to evaluate.
func GetSymlinksEval(vfs avfs.Featurer) []*SymlinkEval {
	if !vfs.HasFeature(avfs.FeatSymlink) {
		return nil
	}

	sl := []*SymlinkEval{
		{NewName: "/A/lroot/", OldName: "/", WantErr: nil, IsSymlink: true, Mode: os.ModeDir | 0o777},
		{NewName: "/A/lroot/B", OldName: "/B", WantErr: nil, IsSymlink: false, Mode: os.ModeDir | 0o755},
		{NewName: "/C/lNonExist", OldName: "A/path", WantErr: avfs.ErrNoSuchFileOrDir, IsSymlink: true},
		{NewName: "/", OldName: "/", WantErr: nil, IsSymlink: false, Mode: os.ModeDir | 0o777},
		{NewName: "/lC", OldName: "/C", WantErr: nil, IsSymlink: true, Mode: os.ModeDir | 0o750},
		{NewName: "/B/1/lafile2.txt", OldName: "/A/afile2.txt", WantErr: nil, IsSymlink: true, Mode: 0o644},
		{NewName: "/B/2/lf", OldName: "/B/2/F", WantErr: nil, IsSymlink: true, Mode: os.ModeDir | 0o755},
		{NewName: "/B/2/F/3/llf", OldName: "/B/2/F", WantErr: nil, IsSymlink: true, Mode: os.ModeDir | 0o755},
		{NewName: "/C/lllf", OldName: "/B/2/F", WantErr: nil, IsSymlink: true, Mode: os.ModeDir | 0o755},
		{NewName: "/B/2/F/3/llf", OldName: "/B/2/F", WantErr: nil, IsSymlink: true, Mode: os.ModeDir | 0o755},
		{NewName: "/C/lllf/3/3file1.txt", OldName: "/B/2/F/3/3file1.txt", WantErr: nil, IsSymlink: false, Mode: 0o640},
	}

	return sl
}

// CreateSymlinks creates test symbolic links.
func CreateSymlinks(t *testing.T, vfs avfs.Fs, rootDir string) []*Symlink {
	symlinks := GetSymlinks(vfs)
	for _, sl := range symlinks {
		oldPath := vfs.Join(rootDir, sl.OldName)
		newPath := vfs.Join(rootDir, sl.NewName)

		err := vfs.Symlink(oldPath, newPath)
		if err != nil {
			t.Fatalf("Symlink %s %s : want error to be nil, got %v", oldPath, newPath, err)
		}
	}

	return symlinks
}
