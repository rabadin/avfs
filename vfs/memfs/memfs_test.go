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

// +build !datarace

package memfs_test

import (
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/idm/dummyidm"
	"github.com/avfs/avfs/idm/memidm"
	"github.com/avfs/avfs/test"
	"github.com/avfs/avfs/vfs/memfs"
)

var (
	// memfs.MemFS struct implements avfs.VFS interface.
	_ avfs.VFS = &memfs.MemFS{}

	// memfs.MemFile struct implements avfs.File interface.
	_ avfs.File = &memfs.MemFile{}
)

func initTest(tb testing.TB) *test.SuiteFS {
	vfsRoot, err := memfs.New(memfs.WithIdm(memidm.New()), memfs.WithMainDirs())
	if err != nil {
		tb.Fatalf("New : want error to be nil, got %v", err)
	}

	sfs := test.NewSuiteFS(tb, vfsRoot)

	return sfs
}

func TestMemFS(t *testing.T) {
	sfs := initTest(t)
	sfs.All(t)
}

func TestMemFSPerm(t *testing.T) {
	sfs := initTest(t)
	sfs.Perm(t)
}

func TestMemFSOptionError(t *testing.T) {
	_, err := memfs.New(memfs.WithIdm(dummyidm.New()))
	if err != avfs.ErrPermDenied {
		t.Errorf("New : want error to be %v, got %v", avfs.ErrPermDenied, err)
	}
}

// TestMemFsOptionName tests MemFS initialization with or without option name (WithName()).
func TestMemFSOptionName(t *testing.T) {
	const wantName = "whatever"

	vfs, err := memfs.New()
	if err != nil {
		t.Fatalf("New : want error to be nil, got %v", err)
	}

	if vfs.Name() != "" {
		t.Errorf("New : want name to be '', got %s", vfs.Name())
	}

	vfs, err = memfs.New(memfs.WithName(wantName))
	if err != nil {
		t.Fatalf("New : want error to be nil, got %v", err)
	}

	name := vfs.Name()
	if name != wantName {
		t.Errorf("New : want name to be %s, got %s", wantName, vfs.Name())
	}
}

func TestMemFSNilPtrFile(t *testing.T) {
	f := (*memfs.MemFile)(nil)

	test.FileNilPtr(t, f)
}

func TestMemFSFeatures(t *testing.T) {
	vfs, err := memfs.New()
	if err != nil {
		t.Fatalf("memfs.New : want error to be nil, got %v", err)
	}

	if vfs.Features()&avfs.FeatIdentityMgr != 0 {
		t.Errorf("Features : want FeatIdentityMgr missing, got present")
	}

	vfs, err = memfs.New(memfs.WithIdm(memidm.New()))
	if err != nil {
		t.Fatalf("memfs.New : want error to be nil, got %v", err)
	}

	if vfs.Features()&avfs.FeatIdentityMgr == 0 {
		t.Errorf("Features : want FeatIdentityMgr present, got missing")
	}
}

func TestMemFSOSType(t *testing.T) {
	vfs, err := memfs.New()
	if err != nil {
		t.Fatalf("New : want error to be nil, got %v", err)
	}

	ost := vfs.OSType()
	if ost != avfs.OsLinux {
		t.Errorf("OSType : want os type to be %v, got %v", avfs.OsLinux, ost)
	}
}

func BenchmarkMemFSAll(b *testing.B) {
	sfs := initTest(b)
	sfs.BenchAll(b)
}
