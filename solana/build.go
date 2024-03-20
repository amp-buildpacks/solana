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
	"fmt"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()
	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}
	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache, err: %v\n", err)
	}
	dc.Logger = b.Logger

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver, err: %v\n", err)
	}
	// check solana cli version config
	v, _ := cr.Resolve("BP_SOLANA_ClI_VERSION")
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to find solana %s dependency, err:%v\n", v, err)
	}

	solanaDependency, err := dr.Resolve("solana-cli", v)
	solanaLayer := NewSolana(solanaDependency, dc, cr)
	solanaLayer.Logger = b.Logger

	deploySolanaContract, _ := cr.Resolve("BP_DEPLOY_SOLANA_CONTRACT")
	result.Processes, err = solanaLayer.BuildProcessTypes(deploySolanaContract)

	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to build list of process types\n%w", err)
	}
	result.Layers = append(result.Layers, solanaLayer)

	return result, nil
}
