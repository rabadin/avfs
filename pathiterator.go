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

package avfs

import "strings"

// PathIterator iterates through an absolute path.
// It returns each part of the path in successive calls to Next.
// The volume name (for Windows) is not considered as part of the path
// it is returned by VolumeName.
//
//  ut := avfs.NewUtils(avfs.OsWindows)
//	pi := ut.NewPathIterator(path)
// 	for pi.Next() {
//		fmt.Println(pi.Part())
//	}
//
// The path below shows the different results of the PathIterator methods
// when thirdPart is the current Part :
//
// /firstPart/secondPart/thirdPart/fourthPart/fifthPart
//                      |- Part --|
//					  Start      End
// |------- Left -------|         |------ Right ------|
// |----- LeftPart ---------------|
//                      |----------- RightPart -------|
//
type PathIterator struct {
	path          string
	start         int
	end           int
	volumeNameLen int
	ut            Utils
}

// NewPathIterator creates a new path iterator from an absolute path.
func (ut *Utils) NewPathIterator(path string) *PathIterator {
	pi := PathIterator{
		path:          path,
		volumeNameLen: ut.VolumeNameLen(path),
		ut:            *ut,
	}

	pi.Reset()

	return &pi
}

// End returns the end position of the current Part.
func (pi *PathIterator) End() int {
	return pi.end
}

// IsLast returns true if the current Part is the last one.
func (pi *PathIterator) IsLast() bool {
	return pi.end == len(pi.path)
}

// Left returns the left path of the current Part.
func (pi *PathIterator) Left() string {
	return pi.path[:pi.start]
}

// LeftPart returns the left path and current Part.
func (pi *PathIterator) LeftPart() string {
	return pi.path[:pi.end]
}

// Next iterates through the next Part of the path.
// It returns false if there's no more parts.
func (pi *PathIterator) Next() bool {
	pi.start = pi.end + 1
	if pi.start >= len(pi.path) {
		pi.end = pi.start

		return false
	}

	pos := strings.IndexByte(pi.path[pi.start:], pi.ut.pathSeparator)
	if pos == -1 {
		pi.end = len(pi.path)
	} else {
		pi.end = pi.start + pos
	}

	return true
}

// Part returns the current Part.
func (pi *PathIterator) Part() string {
	return pi.path[pi.start:pi.end]
}

// Path returns the path to iterate.
func (pi *PathIterator) Path() string {
	return pi.path
}

// ReplacePart replaces the current Part of the path with the new path.
// If the path iterator has been reset it returns true.
// It can be used in symbolic link replacement.
func (pi *PathIterator) ReplacePart(newPath string) bool {
	ut := pi.ut
	oldPath := pi.path

	if ut.IsAbs(newPath) {
		pi.path = ut.Join(newPath, oldPath[pi.end:])
	} else {
		pi.path = ut.Join(oldPath[:pi.start], newPath, oldPath[pi.end:])
	}

	// If the old path before the current part is different, the iterator must be reset.
	if pi.start >= len(pi.path) || pi.path[:pi.start] != oldPath[:pi.start] {
		pi.Reset()

		return true
	}

	// restart from the part before the symbolic link part.
	pi.end = pi.start - 1

	return false
}

// Reset resets the iterator.
func (pi *PathIterator) Reset() {
	pi.end = pi.volumeNameLen
}

// Right returns the right path of the current Part.
func (pi *PathIterator) Right() string {
	return pi.path[pi.end:]
}

// RightPart returns the right path and the current Part.
func (pi *PathIterator) RightPart() string {
	return pi.path[pi.start:]
}

// Start returns the start position of the current Part.
func (pi *PathIterator) Start() int {
	return pi.start
}

// VolumeName returns leading volume name.
func (pi *PathIterator) VolumeName() string {
	return pi.path[:pi.volumeNameLen]
}

// VolumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
func (pi *PathIterator) VolumeNameLen() int {
	return pi.volumeNameLen
}