package brats_test

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet buildpack", func() {
	bratshelper.UnbuiltBuildpack("dotnet", CopyBrats)

	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)

	bratshelper.StagingWithADepThatIsNotTheLatestConstrained(
		"dotnet-sdk",
		FirstOfVersionLine("dotnet-sdk", "2.1.x"),
		func(v string) *cutlass.App { return CopyCSharpBratsWithRuntime(v, "2.1.x") },
	)

	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(CopyBrats)

	bratshelper.DeployAppWithExecutableProfileScript("dotnet-sdk", CopyBrats)

	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	compatible := func(sdkVersion, runtimeVersion string) bool {
		sdkSemver := semver.MustParse(sdkVersion)

		runtimeSemver := semver.MustParse(runtimeVersion)

		hasPreview := isPreview(sdkSemver) || isPreview(runtimeSemver)

		isCompatible := (sdkSemver.Major == runtimeSemver.Major) && (sdkSemver.Minor >= runtimeSemver.Minor) && !hasPreview

		return isCompatible
	}

	// Skip 1.0.X versions of the SDK when testing F# apps
	compatibleWithFSharp := func(sdkVersion, runtimeVersion string) bool {
		sdk := semver.MustParse(sdkVersion)
		if sdk.Major <= 1 && sdk.Minor < 1 {
			return false
		}
		return compatible(sdkVersion, runtimeVersion)
	}

	ensureAppWorks := func(sdkVersion, runtimeVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct version of .NET SDK + .NET Runtime", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet-sdk " + sdkVersion))
			Expect(app.Stdout.String()).To(MatchRegexp(
				"(Using dotnet runtime installed in .*\\Q/dotnet-sdk/shared/Microsoft.NETCore.App/%s\\E|\\QInstalling dotnet-runtime %s\\E)",
				runtimeVersion,
				runtimeVersion,
			))
		})

		By("runs a simple web server", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	}

	Context("for C# apps", func() {
		bratshelper.ForAllSupportedVersions2(
			"dotnet-sdk",
			"dotnet-runtime",
			compatible,
			"with .NET SDK version: %s and .NET Runtime version: %s",
			CopyCSharpBratsWithRuntime,
			ensureAppWorks,
		)
	})

	Context("for F# apps", func() {
		bratshelper.ForAllSupportedVersions2(
			"dotnet-sdk",
			"dotnet-runtime",
			compatibleWithFSharp,
			"with .NET SDK version: %s and .NET Runtime version: %s",
			CopyFSharpBratsWithRuntime,
			ensureAppWorks,
		)
	})
})

func isPreview(version semver.Version) bool {
	if len(version.Pre) == 0 {
		return false
	}
	for _, pre := range version.Pre {
		emptyPR := semver.PRVersion{}
		if pre != emptyPR {
			return true
		}
	}
	return false
}
