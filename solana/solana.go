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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type Hardhat struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
	Executor         effect.Executor
}

func NewHardhat(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) Hardhat {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Build:  true,
		Cache:  true,
		Launch: true,
	})
	return Hardhat{
		LayerContributor: contributor,
		Executor:         effect.NewExecutor(),
	}
}

func (r Hardhat) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	r.LayerContributor.Logger = r.Logger
	return r.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		bin := filepath.Join(layer.Path, "bin")

		r.Logger.Bodyf("Expanding %s to %s", artifact.Name(), bin)
		if err := crush.Extract(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand %s\n%w", artifact.Name(), err)
		}

		r.Logger.Bodyf("Setting %s in PATH", bin)
		if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", bin)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
		}

		// install hardhat
		r.Logger.Bodyf("Installing hardhat")
		if _, err := r.Execute("npm", []string{"install", "hardhat", "--save-dev"}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install hardhat\n%w", err)
		}

		// install node_modules
		r.Logger.Bodyf("Installing node_modules")
		if _, err := r.Execute("npm", []string{"install"}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install node_modules\n%w", err)
		}

		// get hardhat version
		buf, err := r.Execute("npx", []string{"hardhat", "--version"})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to get hardhat version\n%w", err)
		}
		version := strings.TrimSpace(buf.String())
		r.Logger.Bodyf("Checking hardhat version: %s", version)
		return layer, nil
	})
}

func (r Hardhat) Execute(command string, args []string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if err := r.Executor.Execute(effect.Execution{
		Command: command,
		Args:    args,
		Stdout:  buf,
		Stderr:  buf,
	}); err != nil {
		return buf, fmt.Errorf("%s: %w", buf.String(), err)
	}
	return buf, nil
}

func (r Hardhat) BuildProcessTypes(enableProcess string) ([]libcnb.Process, error) {
	processes := []libcnb.Process{}

	if enableProcess == "true" {
		processes = append(processes, libcnb.Process{
			Type:      "web",
			Command:   "npm",
			Arguments: []string{"start"},
			Default:   true,
		})
	}
	return processes, nil
}

func (r Hardhat) Name() string {
	return r.LayerContributor.LayerName()
}
