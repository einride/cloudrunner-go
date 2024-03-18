package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"go.einride.tech/sage/sg"
	"go.einride.tech/sage/tools/sgconvco"
	"go.einride.tech/sage/tools/sggit"
	"go.einride.tech/sage/tools/sggo"
	"go.einride.tech/sage/tools/sggolangcilint"
	"go.einride.tech/sage/tools/sggolicenses"
	"go.einride.tech/sage/tools/sgmdformat"
	"go.einride.tech/sage/tools/sgyamlfmt"
)

func main() {
	sg.GenerateMakefiles(
		sg.Makefile{
			Path:          sg.FromGitRoot("Makefile"),
			DefaultTarget: All,
		},
	)
}

func All(ctx context.Context) error {
	sg.Deps(ctx, ConvcoCheck, GoLint, GoTest, FormatMarkdown, FormatYAML, ReadmeSnippet)
	sg.Deps(ctx, GoModTidy)
	sg.Deps(ctx, GitVerifyNoDiff)
	return nil
}

func FormatYAML(ctx context.Context) error {
	sg.Logger(ctx).Println("formatting YAML files...")
	return sgyamlfmt.Run(ctx)
}

func GoModTidy(ctx context.Context) error {
	sg.Logger(ctx).Println("tidying Go module files...")
	return sg.Command(ctx, "go", "mod", "tidy", "-v").Run()
}

func GoTest(ctx context.Context) error {
	sg.Logger(ctx).Println("running Go tests...")
	return sggo.TestCommand(ctx).Run()
}

func GoLint(ctx context.Context) error {
	sg.Logger(ctx).Println("linting Go files...")
	return sggolangcilint.Run(ctx)
}

// TODO: Add this to All target once it is working again.
func GoLicenses(ctx context.Context) error {
	sg.Logger(ctx).Println("checking Go licenses...")
	return sggolicenses.Check(ctx)
}

func FormatMarkdown(ctx context.Context) error {
	sg.Logger(ctx).Println("formatting Markdown files...")
	return sgmdformat.Command(ctx).Run()
}

func ConvcoCheck(ctx context.Context) error {
	sg.Logger(ctx).Println("checking git commits...")
	return sgconvco.Command(ctx, "check", "origin/master..HEAD").Run()
}

func GitVerifyNoDiff(ctx context.Context) error {
	sg.Logger(ctx).Println("verifying that git has no diff...")
	return sggit.VerifyNoDiff(ctx)
}

func ReadmeSnippet(ctx context.Context) error {
	usage := sg.Output(sg.Command(ctx, "go", "run", "./examples/cmd/grpc-server", "-help"))
	usage = strings.TrimSpace(usage)
	usage = "<!-- BEGIN usage -->\n\n```\n" + usage
	usage += "\n```\n\n<!-- END usage -->"
	readme, err := os.ReadFile("README.md")
	if err != nil {
		return err
	}
	usageRegexp := regexp.MustCompile(`(?ms)<!-- BEGIN usage -->.*<!-- END usage -->`)
	if !usageRegexp.Match(readme) {
		return fmt.Errorf("found no match for 'usage' snippet in README.md")
	}
	return os.WriteFile("README.md", usageRegexp.ReplaceAll(readme, []byte(usage)), 0o600)
}
