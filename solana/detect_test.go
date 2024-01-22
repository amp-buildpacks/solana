/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package solana

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.DetectContext
		detect Detect
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = os.MkdirTemp("", "solana")
		Expect(err).ToNot(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
	})

	context("missing requires file", func() {
		it("missing Cargo.toml", func() {
			Expect(os.Mkdir(filepath.Join(ctx.Application.Path, "src"), 0755)).ToNot(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "src/lib.rs"), []byte{}, 0644)).ToNot(HaveOccurred())

			plan, err := detect.Detect(ctx)
			Expect(err).To(MatchError(ErrCargoFileNotFound))
			Expect(plan).To(Equal(libcnb.DetectResult{
				Pass: false,
			}))
		})

		it("missing src/lib.rs file", func() {
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "Cargo.toml"), []byte{}, 0644)).ToNot(HaveOccurred())
			plan, err := detect.Detect(ctx)
			Expect(err).To(MatchError(ErrSrcLibFileNotFound))
			Expect(plan).To(Equal(libcnb.DetectResult{
				Pass: false,
			}))
		})
	})

	it("passes with both Cargo.toml and src/lib.rs", func() {
		Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "Cargo.toml"), []byte{}, 0644)).ToNot(HaveOccurred())
		Expect(os.Mkdir(filepath.Join(ctx.Application.Path, "src"), 0755)).ToNot(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "src/lib.rs"), []byte{}, 0644)).ToNot(HaveOccurred())

		Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
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
		}))
	})
}
