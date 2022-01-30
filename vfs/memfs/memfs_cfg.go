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

package memfs

import (
	"io/fs"

	"github.com/avfs/avfs"
)

// New returns a new memory file system (MemFS).
func New(opts ...Option) *MemFS {
	ma := &memAttrs{
		idm:      avfs.NotImplementedIdm,
		dirMode:  fs.ModeDir,
		fileMode: 0,
		features: avfs.FeatBasicFs |
			avfs.FeatChroot |
			avfs.FeatHardlink |
			avfs.FeatSymlink,
	}

	vfs := &MemFS{
		user:     avfs.DefaultUser,
		curDir:   "/",
		rootNode: createRootNode(),
		memAttrs: ma,
		utils:    avfs.Cfg.Utils(),
	}

	for _, opt := range opts {
		opt(vfs)
	}

	volumeName := ""

	ut := vfs.utils
	if ut.OSType() == avfs.OsWindows {
		ma.features ^= avfs.FeatChroot
		ma.dirMode |= avfs.DefaultDirPerm
		ma.fileMode |= avfs.DefaultFilePerm

		vfs.volumes = make(volumes)
		volumeName = avfs.DefaultVolume
		vfs.volumes[volumeName] = vfs.rootNode
	}

	vfs.umask = avfs.Cfg.UMask()
	vfs.rootNode.mode = ma.dirMode &^ vfs.umask

	vfs.err.OSType(vfs.OSType())

	if vfs.HasFeature(avfs.FeatMainDirs) {
		u := vfs.user
		um := vfs.umask

		vfs.user = avfs.AdminUser
		vfs.umask = 0

		err := ut.CreateBaseDirs(vfs, volumeName)
		if err != nil {
			panic("CreateBaseDirs " + err.Error())
		}

		vfs.umask = um
		vfs.user = u
		vfs.curDir = ut.HomeDirUser(u.Name())
	}

	return vfs
}

// Features returns the set of features provided by the file system or identity manager.
func (vfs *MemFS) Features() avfs.Features {
	return vfs.memAttrs.features
}

// HasFeature returns true if the file system or identity manager provides a given features.
func (vfs *MemFS) HasFeature(feature avfs.Features) bool {
	return vfs.memAttrs.features&feature == feature
}

// Name returns the name of the fileSystem.
func (vfs *MemFS) Name() string {
	return vfs.memAttrs.name
}

// OSType returns the operating system type of the file system.
func (vfs *MemFS) OSType() avfs.OSType {
	return vfs.utils.OSType()
}

// Type returns the type of the fileSystem or Identity manager.
func (vfs *MemFS) Type() string {
	return "MemFS"
}

// Options

// WithMainDirs returns an option function to create main directories.
func WithMainDirs() Option {
	return func(vfs *MemFS) {
		vfs.memAttrs.features |= avfs.FeatMainDirs
	}
}

// WithIdm returns an option function which sets the identity manager.
func WithIdm(idm avfs.IdentityMgr) Option {
	return func(vfs *MemFS) {
		vfs.memAttrs.idm = idm
		vfs.memAttrs.features |= idm.Features()
		vfs.user = idm.AdminUser()
	}
}

// WithName returns an option function which sets the name of the file system.
func WithName(name string) Option {
	return func(vfs *MemFS) {
		vfs.memAttrs.name = name
	}
}

// WithOSType returns an option function which sets the OS type.
func WithOSType(osType avfs.OSType) Option {
	return func(vfs *MemFS) {
		vfs.utils = avfs.NewUtils(osType)
	}
}
