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

type Solana struct {
	LayerContributor libpak.DependencyLayerContributor
	configResolver   libpak.ConfigurationResolver
	Logger           bard.Logger
	Executor         effect.Executor
	EndPoint         string
	WalletAddress    string
}

func NewSolana(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, configResolver libpak.ConfigurationResolver) Solana {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Build:  true,
		Cache:  true,
		Launch: true,
	})
	return Solana{
		LayerContributor: contributor,
		Executor:         effect.NewExecutor(),
		configResolver:   configResolver,
	}
}

func (r Solana) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
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

		// get solana version
		buf, err := r.Execute("solana", []string{"--version"})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to get solana version\n%w", err)
		}
		version := strings.TrimSpace(buf.String())
		r.Logger.Bodyf("Checking solana-cli version: %s", version)

		// compile contract
		var args []string
		r.Logger.Bodyf("Compiling contracts")
		if _, err := r.Execute("cargo-build-bpf", args); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to compile contract\n%w", err)
		}
		// 1. config deploy endpoint
		if _, err := r.InitializeEnv(); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to initialize solana endpoint\n%w", err)
		}
		// 2. import wallet and valid wallet
		if err = r.ImportWalletAndValid(); err != nil {
			return libcnb.Layer{}, fmt.Errorf("import wallet and valid fail\n%w", err)
		}
		return layer, nil
	})
}

// InitializeEnv config deploy endpoint
func (r Solana) InitializeEnv() (*bytes.Buffer, error) {
	r.Logger.Bodyf("Initializing solana endpoint")
	deployNet, ok := r.configResolver.Resolve("BP_SOLANA_DEPLOY_NETWORK")
	if !ok {
		return nil, fmt.Errorf("unable to resolve deploy network")
	}
	// get endpoint config
	deployEndPoint, ok := r.configResolver.Resolve(fmt.Sprintf("BP_%s_ENDPOINT", strings.ToUpper(deployNet)))
	if !ok {
		return nil, fmt.Errorf("unable to resolve deploy deploy endpoint")
	}
	r.EndPoint = deployEndPoint
	args := []string{
		"config",
		"set",
		"--url",
		deployEndPoint,
	}
	return r.Execute(PlanEntrySolana, args)
}

// ImportWalletAndValid write to id.json and valid
func (s Solana) ImportWalletAndValid() error {
	keyPair, ok := s.configResolver.Resolve("BP_WALLET_KEYPAIR")
	if !ok {
		return fmt.Errorf("resolve BP_WALLET_KEYPAIR err")
	}
	keyPairPath := "/tmp/id.json"
	// write to /tmp/id.json
	err := os.WriteFile(keyPairPath, []byte(keyPair), 0644)
	if err != nil {
		s.Logger.Logger.Infof("write id.json err", err)
		return err
	}
	// update config keypair
	args := []string{
		"config",
		"set",
		"--keypair",
		keyPairPath,
	}
	_, err = s.Execute(PlanEntrySolana, args)
	if err != nil {
		s.Logger.Logger.Infof("solana config set keypair err", err)
		return err
	}
	// args := []string{
	//     "config",
	//     "set",
	//     "--keypair",
	//     keyPairPath,
	// }
	// _, err = s.Execute(PlanEntrySolana, args)
	// if err != nil {
	//     s.Logger.Logger.Infof("solana config set keypair err", err)
	//     return err
	// }
	return nil
}

func (r Solana) Execute(command string, args []string) (*bytes.Buffer, error) {
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

func (r Solana) BuildProcessTypes(enableProcess string) ([]libcnb.Process, error) {
	var processes []libcnb.Process
	if enableProcess == "true" {
		r.Logger.Bodyf(" solana deploy  %s for %s endpoint", r.WalletAddress, r.WalletAddress)
		processes = append(processes, libcnb.Process{
			Type:      "cli",
			Command:   "solana ",
			Arguments: []string{PlanEntrySolana, "deploy", "./target/deploy/*.so"},
			Default:   true,
		})
	}
	return processes, nil
}

func (r Solana) Name() string {
	return r.LayerContributor.LayerName()
}
