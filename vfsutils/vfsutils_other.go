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

// +build !linux

package vfsutils

import (
	"math"
	"os"

	"github.com/avfs/avfs"
)

// UMaskType is the file mode creation mask.
// it must be set to be read, so it must be protected with a mutex.
type UMaskType struct{}

// GetUMask returns the file mode creation mask.
func (um *UMaskType) Get() os.FileMode {
	return 0o022
}

// UMask sets the file mode creation mask.
// umask must be set to 0 using umask(2) system call to be read,
// so its value is cached and protected by a mutex.
func (um *UMaskType) Set(mask os.FileMode) {
}

// AsStatT converts a value as an avfs.StatT.
func AsStatT(value interface{}) *avfs.StatT {
	switch s := value.(type) {
	case *avfs.StatT:
		return s
	default:
		return &avfs.StatT{Uid: math.MaxUint32, Gid: math.MaxUint32}
	}
}
