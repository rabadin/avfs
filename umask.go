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

package avfs

import (
	"io/fs"
	"sync"
)

var (
	umask  fs.FileMode  //nolint:gochecknoglobals // Used by UMask and SetUMask.
	umLock sync.RWMutex //nolint:gochecknoglobals // Used by UMask and SetUMask.
)

func init() { //nolint:gochecknoinits // To initialize umask.
	SetUMask(0)
}

// UMask returns the file mode creation mask.
func UMask() fs.FileMode {
	umLock.RLock()
	um := umask
	umLock.RUnlock()

	return um
}
