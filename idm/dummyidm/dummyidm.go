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

// Package dummyidm implements a dummy identity manager where all functions are not implemented.
package dummyidm

import (
	"github.com/avfs/avfs"
)

// CurrentUser returns the current user.
func (idm *DummyIdm) CurrentUser() avfs.UserReader {
	return NotImplementedUser
}

// GroupAdd adds a new group.
func (idm *DummyIdm) GroupAdd(name string) (avfs.GroupReader, error) {
	return nil, avfs.ErrPermDenied
}

// GroupDel deletes an existing group.
func (idm *DummyIdm) GroupDel(name string) error {
	return avfs.ErrPermDenied
}

// LookupGroup looks up a group by name.
// If the group cannot be found, the returned error is of type UnknownGroupError.
func (idm *DummyIdm) LookupGroup(name string) (avfs.GroupReader, error) {
	return nil, avfs.ErrPermDenied
}

// LookupGroupId looks up a group by groupid.
// If the group cannot be found, the returned error is of type UnknownGroupIdError.
func (idm *DummyIdm) LookupGroupId(gid int) (avfs.GroupReader, error) {
	return nil, avfs.ErrPermDenied
}

// LookupUser looks up a user by username.
// If the user cannot be found, the returned error is of type UnknownUserError.
func (idm *DummyIdm) LookupUser(name string) (avfs.UserReader, error) {
	return nil, avfs.ErrPermDenied
}

// LookupUserId looks up a user by userid.
// If the user cannot be found, the returned error is of type UnknownUserIdError.
func (idm *DummyIdm) LookupUserId(uid int) (avfs.UserReader, error) {
	return nil, avfs.ErrPermDenied
}

// User sets the current user of the file system to uid.
// If the current user has not root privileges avfs.errPermDenied is returned.
func (idm *DummyIdm) User(name string) (avfs.UserReader, error) {
	return nil, avfs.ErrPermDenied
}

// UserAdd adds a new user.
func (idm *DummyIdm) UserAdd(name, groupName string) (avfs.UserReader, error) {
	return nil, avfs.ErrPermDenied
}

// UserDel deletes an existing group.
func (idm *DummyIdm) UserDel(name string) error {
	return avfs.ErrPermDenied
}

// Group

// Gid returns the Group ID.
func (g *Group) Gid() int {
	return g.gid
}

// Name returns the Group name.
func (g *Group) Name() string {
	return g.name
}

// User

// Gid returns the primary Group ID of the User.
func (u *User) Gid() int {
	return u.gid
}

// IsRoot returns true if the User has root privileges.
func (u *User) IsRoot() bool {
	return u.uid == 0 || u.gid == 0
}

// Name returns the User name.
func (u *User) Name() string {
	return u.name
}

// Uid returns the User ID.
func (u *User) Uid() int {
	return u.uid
}
