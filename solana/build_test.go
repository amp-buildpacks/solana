package solana

import (
	"os"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Except = NewWithT(t).Expect
		build  Build
		ctx    libcnb.BuildContext
	)
	context("build and deploy", func() {
		it.Before(func() {
			ctx.Application.Path, _ = os.MkdirTemp("", "build")
			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "solana"})

			ctx.Buildpack.Metadata = map[string]interface{}{
				"dependencies": []map[string]interface{}{
					{
						"id":               "solana-cli",
						"name":             "Solana Cli",
						"version":          "1.17.17",
						"stacks":           []interface{}{"io.buildpacks.stacks.jammy"},
						"purl":             "pkg:generic/solana@1.17.17?download_url=https://github.com/solana-labs/solana/releases/download/v1.17.17/solana-release-x86_64-unknown-linux-gnu.tar.bz2",
						"sha256":           "1290babaf9a45034a78f04acf8f960d39c58fdf91b0eaa9d5a4ec138850d98ea",
						"uri":              "https://github.com/solana-labs/solana/releases/download/v1.17.17/solana-release-x86_64-unknown-linux-gnu.tar.bz2",
						"strip-components": 1,
						"licenses":         []string{"Apache-2.0"},
					},
				},
			}
		})
		it.After(func() {
			_ = os.RemoveAll(ctx.Application.Path)
		})
		it("$BP_DEPLOY_SOLANA_CONTRACT is set true", func() {
			_, err := build.Build(ctx)
			Except(err).NotTo(HaveOccurred())
		})
	})
}
