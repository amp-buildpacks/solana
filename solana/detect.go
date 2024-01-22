// Copyright (c) The Amphitheatre Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package solana

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
)

var (
	PlanEntrySolana                = "solana"
	ErrCargoFileNotFound           = errors.New("cargo.toml file not found")
	ErrUnableToDetermineCargoFile  = errors.New("unable to determine if Cargo.toml exists")
	ErrSrcLibFileNotFound          = errors.New("src/lib.rs file not found")
	ErrUnableToDetermineSrcLibFile = errors.New("unable to determine if src/lib.rs exists")
)

type Detect struct {
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	found, err := d.solanaProject(context.Application.Path)
	if err != nil {
		return libcnb.DetectResult{Pass: false}, err
	}

	if !found {
		return libcnb.DetectResult{Pass: false}, nil
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: PlanEntrySolana},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: PlanEntrySolana},
				},
			},
		},
	}, nil
}

func (d Detect) solanaProject(appDir string) (bool, error) {
	// check Cargo.toml file is exists
	_, err := os.Stat(filepath.Join(appDir, "Cargo.toml"))
	if os.IsNotExist(err) {
		return false, ErrCargoFileNotFound
	} else if err != nil {
		return false, ErrUnableToDetermineCargoFile
	}
	// check src/lib.ts file is exists
	_, err = os.Stat(filepath.Join(appDir, "src/lib.rs"))
	if os.IsNotExist(err) {
		return false, ErrSrcLibFileNotFound
	} else if err != nil {
		return false, ErrUnableToDetermineSrcLibFile
	}
	return true, nil
}
