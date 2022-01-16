//go:build mage
// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.einride.tech/mage-tools/mgmake"
	"go.einride.tech/mage-tools/mgpath"
	"os"
	"regexp"
	"strings"

	// mage:import
	"go.einride.tech/mage-tools/targets/mgyamlfmt"

	// mage:import
	"go.einride.tech/mage-tools/targets/mgconvco"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggo"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggoreview"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggolangcilint"

	// mage:import
	"go.einride.tech/mage-tools/targets/mgmarkdownfmt"

	// mage:import
	"go.einride.tech/mage-tools/targets/mggitverifynodiff"
)

func init() {
	mgmake.GenerateMakefiles(
		mgmake.Makefile{
			Path:          mgpath.FromGitRoot("Makefile"),
			DefaultTarget: All,
		},
	)
}

func All() {
	mg.Deps(
		mg.F(mgconvco.ConvcoCheck, "origin/master..HEAD"),
		mggolangcilint.GolangciLint,
		mggoreview.Goreview,
		mggo.GoTest,
		mgmarkdownfmt.FormatMarkdown,
		mgyamlfmt.FormatYaml,
	)
	mg.Deps(
		ReadmeSnippet,
		mggo.GoModTidy,
	)
	mg.SerialDeps(
		mggitverifynodiff.GitVerifyNoDiff,
	)
}

func ReadmeSnippet() error {
	usage, err := sh.Output("go", "run", "./examples/cmd/grpc-server", "-help")
	if err != nil {
		return err
	}
	usage = strings.TrimSpace(usage)
	usage = "<!-- BEGIN usage -->\n\n```\n" + usage
	usage = usage + "\n```\n\n<!-- END usage -->"
	readme, err := os.ReadFile("README.md")
	if err != nil {
		return err
	}
	usageRegexp, err := regexp.Compile(`(?ms)<!-- BEGIN usage -->.*<!-- END usage -->`)
	if err != nil {
		return err
	}
	if !usageRegexp.Match(readme) {
		return fmt.Errorf("found no match for 'usage' snippet in README.md")
	}
	return os.WriteFile("README.md", usageRegexp.ReplaceAll(readme, []byte(usage)), 0o600)
}
