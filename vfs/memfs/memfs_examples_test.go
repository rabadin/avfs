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

//go:build avfs_setostype

package memfs_test

import (
	"errors"
	"fmt"
	"io/fs"
	"log"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/idm/dummyidm"
	"github.com/avfs/avfs/idm/memidm"
	"github.com/avfs/avfs/test"
	"github.com/avfs/avfs/vfs/memfs"
)

// ExampleNew should produce the same results independently of the host OS.
func ExampleNew() {
	idm := memidm.NewWithOptions(&memidm.Options{OSType: avfs.OsLinux})
	vfs := memfs.NewWithOptions(&memfs.Options{Idm: idm, OSType: avfs.OsLinux})

	fmt.Println(vfs.Features())
	fmt.Println(vfs.User().Name())
	fmt.Println(vfs.TempDir())

	homeDir := avfs.HomeDirUser(vfs, vfs.User())
	fmt.Println(homeDir)

	_, err := vfs.Stat(homeDir)
	if err == nil {
		fmt.Printf("%s exists", homeDir)
	}

	// Features(Hardlink|IdentityMgr|SetOSType|SubFS|Symlink|SystemDirs)
	// root
	// /tmp
	// /root
	// /root exists
}

func ExampleNewWithOptions_noSystemDirs() {
	vfs := memfs.NewWithOptions(&memfs.Options{OSType: avfs.OsLinux})
	tmpDir := vfs.TempDir()

	_, err := vfs.Stat(tmpDir)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("%s does not exist", tmpDir)
	}

	// Output: /tmp does not exist
}

func ExampleNewWithOptions_noIdm() {
	vfs := memfs.NewWithOptions(&memfs.Options{Idm: dummyidm.NotImplementedIdm, OSType: avfs.OsLinux})
	fmt.Println(vfs.User().Name())

	// Output: Default
}

func ExampleMemFS_Sub() {
	idm := memidm.NewWithOptions(&memidm.Options{OSType: avfs.OsLinux})
	vfsSrc := memfs.NewWithOptions(&memfs.Options{Idm: idm, OSType: avfs.OsLinux})

	_, err := vfsSrc.Idm().UserAdd(test.UsrTest, "root")
	if err != nil {
		log.Fatalf("UserAdd : want error to be nil, got %v", err)
	}

	vfsSub, err := vfsSrc.Sub("/")
	if err != nil {
		log.Fatalf("Sub : want error to be nil, got %v", err)
	}

	err = vfsSub.SetUserByName(test.UsrTest)
	if err != nil {
		log.Fatalf("SetUser : want error to be nil, got %v", err)
	}

	fmt.Println(vfsSrc.User().Name())
	fmt.Println(vfsSub.User().Name())

	// Output:
	// root
	// UsrTest
}

func ExampleMemFS_SetUserByName() {
	idm := memidm.NewWithOptions(&memidm.Options{OSType: avfs.OsLinux})
	vfs := memfs.NewWithOptions(&memfs.Options{Idm: idm, OSType: avfs.OsLinux})

	_, err := vfs.Idm().UserAdd(test.UsrTest, idm.AdminGroup().Name())
	if err != nil {
		log.Fatalf("UserAdd : want error to be nil, got %v", err)
	}

	fmt.Println(vfs.User().Name())

	err = vfs.SetUserByName(test.UsrTest)
	if err != nil {
		log.Fatalf("SetUser : want error to be nil, got %v", err)
	}

	fmt.Println(vfs.User().Name())

	// Output:
	// root
	// UsrTest
}
