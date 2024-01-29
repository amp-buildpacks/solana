package solana

import (
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
		it("$BP_DEPLOY_SOLANA_CONTRACT is set true", func() {
			result, err := build.Build(ctx)
			Except(err).NotTo(HaveOccurred())
			Expect(result.Layers).To(HaveLen(0))
		})
	})
}
