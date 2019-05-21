package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple Integration Test", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	// TODO This test is pending because it currently fails. It is just an example
	It("app deploys", func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "rubyapp"))
		app.Manifest = filepath.Join(bpDir, "fixtures", "rubyapp", "manifest.cfdev.yml")
		V3PushAppAndConfirm(app)
		Expect(app.GetBody("/")).To(ContainSubstring("Hi, I'm an app with a sidecar!"))
		Expect(app.GetBody("/config")).To(ContainSubstring(`{"Scope":"some-service.admin","Password":"not-a-real-p4$$w0rd"}`))
	})
})
