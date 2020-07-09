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

package dummyfs_test

import (
	"math"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/fs/dummyfs"
	"github.com/avfs/avfs/test"
)

var (
	// dummyFs.DummyFs struct implements avfs.DummyFs interface
	_ avfs.Fs = &dummyfs.DummyFs{}

	// dummyfs.DummyFile struct implements avfs.DummyFile interface
	_ avfs.File = &dummyfs.DummyFile{}
)

//
func TestDummyFs(t *testing.T) {
	fs, err := dummyfs.New()
	if err != nil {
		t.Fatalf("New : want err to be nil, got %s", err)
	}

	if fs.HasFeature(avfs.Feature(math.MaxUint64)) {
		t.Error("HasFeature : want HasFeature(whatever) to be false, got true")
	}

	cf := test.NewConfigFs(t, fs)
	cf.SuiteNotImplemented()
}

func TestDummyFsFeatures(t *testing.T) {
	fs, err := dummyfs.New()
	if err != nil {
		t.Fatalf("dummy.New : want error to be nil, got %v", err)
	}

	if fs.Features() != 0 {
		t.Errorf("Features : want Features to be 0, got %d", fs.Features())
	}
}