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

package osfs

import (
	"github.com/avfs/avfs"
)

// New returns a new OsFS file system.
func New(opts ...Option) *OsFS {
	vfs := &OsFS{
		idm:      avfs.NotImplementedIdm,
		features: avfs.FeatBasicFs | avfs.FeatRealFS | avfs.FeatMainDirs | avfs.FeatSymlink | avfs.FeatHardlink,
	}

	vfs.InitUtils(avfs.CurrentOSType())

	if vfs.OSType() == avfs.OsLinux {
		vfs.features |= avfs.FeatChroot
	}

	for _, opt := range opts {
		opt(vfs)
	}

	vfs.setErrors()

	return vfs
}

// setErrors sets OsFS errors depending on the operating system.
func (vfs *OsFS) setErrors() {
	switch vfs.OSType() {
	case avfs.OsWindows:
		vfs.err.PermDenied = avfs.ErrWinAccessDenied
	default:
		vfs.err.PermDenied = avfs.ErrPermDenied
	}
}

// Features returns the set of features provided by the file system or identity manager.
func (vfs *OsFS) Features() avfs.Features {
	return vfs.features
}

// HasFeature returns true if the file system or identity manager provides a given features.
func (vfs *OsFS) HasFeature(feature avfs.Features) bool {
	return vfs.features&feature == feature
}

// Name returns the name of the fileSystem.
func (vfs *OsFS) Name() string {
	return ""
}

// Type returns the type of the fileSystem or Identity manager.
func (vfs *OsFS) Type() string {
	return "OsFS"
}

// Options

// WithIdm returns a function setting the identity manager for the file system.
func WithIdm(idm avfs.IdentityMgr) Option {
	return func(vfs *OsFS) {
		vfs.idm = idm
		vfs.features |= idm.Features()
	}
}

// Configuration functions.

// CreateSystemDirs creates the system directories of a file system.
func (vfs *OsFS) CreateSystemDirs(basePath string) error {
	return vfs.Utils.CreateSystemDirs(vfs, basePath)
}

// CreateHomeDir creates and returns the home directory of a user.
// If there is an error, it will be of type *PathError.
func (vfs *OsFS) CreateHomeDir(u avfs.UserReader) (string, error) {
	return vfs.Utils.CreateHomeDir(vfs, u)
}
