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

package orefafs

import (
	"io/fs"

	"github.com/avfs/avfs"
)

// New returns a new memory file system (OrefaFS) with the default Options.
func New() *OrefaFS {
	return NewWithOptions(nil)
}

// NewWithOptions returns a new memory file system (OrefaFS) with the selected Options.
func NewWithOptions(opts *Options) *OrefaFS {
	if opts == nil {
		opts = &Options{
			OSType:     avfs.CurrentOSType(),
			SystemDirs: true,
		}
	}

	features := avfs.FeatHardlink
	if opts.SystemDirs {
		features |= avfs.FeatSystemDirs
	}

	vfs := &OrefaFS{
		nodes:    make(nodes),
		user:     avfs.NotImplementedIdm.AdminUser(),
		curDir:   "",
		features: features,
		umask:    avfs.UMask(),
		dirMode:  fs.ModeDir,
		fileMode: 0,
	}

	vfs.SetOSType(opts.OSType)

	var volumeName string

	if vfs.OSType() == avfs.OsWindows {
		volumeName = avfs.DefaultVolume
		vfs.dirMode |= avfs.DefaultDirPerm
		vfs.fileMode |= avfs.DefaultFilePerm
	}

	vfs.nodes[volumeName] = createRootNode()
	vfs.curDir = volumeName

	vfs.err.SetOSType(vfs.OSType())

	if vfs.HasFeature(avfs.FeatSystemDirs) {
		// Save the current user and umask.
		u := vfs.user
		um := vfs.umask

		// Create system directories as administrator user without umask.
		vfs.user = avfs.NotImplementedIdm.AdminUser()
		vfs.umask = 0

		err := vfs.CreateSystemDirs(volumeName)
		if err != nil {
			panic("CreateSystemDirs " + err.Error())
		}

		// Restore the previous user and umask.
		vfs.umask = um
		vfs.user = u
		vfs.curDir = vfs.Utils.HomeDirUser(u)
	}

	return vfs
}

// Features returns the set of features provided by the file system or identity manager.
func (vfs *OrefaFS) Features() avfs.Features {
	return vfs.features
}

// HasFeature returns true if the file system or identity manager provides a given features.
func (vfs *OrefaFS) HasFeature(feature avfs.Features) bool {
	return vfs.features&feature == feature
}

// Name returns the name of the fileSystem.
func (vfs *OrefaFS) Name() string {
	return vfs.name
}

// Type returns the type of the fileSystem or Identity manager.
func (vfs *OrefaFS) Type() string {
	return "OrefaFS"
}

// Configuration functions.

// CreateSystemDirs creates system directories of a file system.
func (vfs *OrefaFS) CreateSystemDirs(basePath string) error {
	return vfs.Utils.CreateSystemDirs(vfs, basePath)
}

// CreateHomeDir creates and returns the home directory of a user.
// If there is an error, it will be of type *PathError.
func (vfs *OrefaFS) CreateHomeDir(u avfs.UserReader) (string, error) {
	return vfs.Utils.CreateHomeDir(vfs, u)
}
