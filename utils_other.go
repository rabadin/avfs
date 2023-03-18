//
//  Copyright 2022 The AVFS authors
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

//go:build !linux && !windows

package avfs

import (
	"io/fs"
	"sync/atomic"
)

// ShortPathName Retrieves the short path form of the specified path (Windows only).
func ShortPathName(path string) string {
	return path
}

// umask is the file mode creation mask.
var umask fs.FileMode = 0o022 //nolint:gochecknoglobals // Used by UMask and SetUMask.

// SetUMask sets the file mode creation mask.
func SetUMask(mask fs.FileMode) {
	m := uint32(mask & fs.ModePerm)
	atomic.StoreUint32((*uint32)(&umask), m)
}

// UMask returns the file mode creation mask.
func UMask() fs.FileMode {
	um := atomic.LoadUint32((*uint32)(&umask))

	return fs.FileMode(um)
}

// VolumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
//func (ut *Utils[_]) VolumeNameLen(path string) int {
//	return 0
//}
