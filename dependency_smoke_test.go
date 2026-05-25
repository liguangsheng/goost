package goost

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const modulePath = "github.com/liguangsheng/goost"

func TestCorePackagesDoNotImportOptionalIntegrations(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()
	imports := corePackageImports(t, root)

	writeFile(t, filepath.Join(dir, "go.mod"), fmt.Sprintf(`module goostdepcheck

go 1.25

require `+modulePath+` v0.0.0

replace `+modulePath+` => %s
`, filepath.ToSlash(root)))

	writeFile(t, filepath.Join(dir, "main.go"), consumerMain(imports))

	cmd := exec.Command("go", "list", "-mod=mod", "-buildvcs=false", "-deps", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}

	for _, unwanted := range []string{
		"github.com/gin-gonic/gin",
		"google.golang.org/grpc",
		"go.opentelemetry.io/otel",
	} {
		if importListContainsPrefix(string(out), unwanted) {
			t.Fatalf("core package dependency list unexpectedly includes %s\n%s", unwanted, out)
		}
	}
}

func TestBasePackagesDoNotImportIntegrationDependencies(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()
	imports := basePackageImports(t, root)

	writeFile(t, filepath.Join(dir, "go.mod"), fmt.Sprintf(`module goostbasecheck

go 1.25

require `+modulePath+` v0.0.0

replace `+modulePath+` => %s
`, filepath.ToSlash(root)))

	writeFile(t, filepath.Join(dir, "main.go"), consumerMain(imports))

	cmd := exec.Command("go", "list", "-mod=mod", "-buildvcs=false", "-deps", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}

	for _, unwanted := range []string{
		"github.com/gin-gonic/gin",
		"go.uber.org/zap",
		"google.golang.org/grpc",
		"go.opentelemetry.io/otel",
	} {
		if importListContainsPrefix(string(out), unwanted) {
			t.Fatalf("base package dependency list unexpectedly includes %s\n%s", unwanted, out)
		}
	}
}

func TestReadmePackageListMatchesPublicPackages(t *testing.T) {
	root := repoRoot(t)
	want := publicPackageNames(t, root)
	got := readmePackageNames(t, filepath.Join(root, "README.md"))
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("README package list mismatch\nREADME:\n%s\npackages:\n%s",
			strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	gotZH := readmePackageNames(t, filepath.Join(root, "README.zh.md"))
	if strings.Join(gotZH, "\n") != strings.Join(want, "\n") {
		t.Fatalf("README.zh package list mismatch\nREADME.zh:\n%s\npackages:\n%s",
			strings.Join(gotZH, "\n"), strings.Join(want, "\n"))
	}
}

func TestReadmeOptionalModuleListMatchesNestedModules(t *testing.T) {
	root := repoRoot(t)
	want := optionalModuleNames(t, root)
	got := readmeSectionNames(t, filepath.Join(root, "README.md"), "## Optional Integration Modules")
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("README optional module list mismatch\nREADME:\n%s\nmodules:\n%s",
			strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	gotZH := readmeSectionNames(t, filepath.Join(root, "README.zh.md"), "## 可选集成 Modules")
	if strings.Join(gotZH, "\n") != strings.Join(want, "\n") {
		t.Fatalf("README.zh optional module list mismatch\nREADME.zh:\n%s\nmodules:\n%s",
			strings.Join(gotZH, "\n"), strings.Join(want, "\n"))
	}
}

func TestNestedModulesHaveDocsAndExamples(t *testing.T) {
	root := repoRoot(t)
	for _, name := range nestedModuleNames(t, root) {
		dir := filepath.Join(root, filepath.FromSlash(name))
		if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
			t.Fatalf("%s: missing README.md: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(dir, "README.zh.md")); err != nil {
			t.Fatalf("%s: missing README.zh.md: %v", name, err)
		}
		if !isOptionalIntegrationModule(name) {
			continue
		}
		if !packageHasCompiledExample(t, dir) {
			t.Fatalf("%s: optional integration module has no compiled Example test", name)
		}
	}
}

func TestExamplesReadmesMatchRunnablePrograms(t *testing.T) {
	root := repoRoot(t)
	want := runnableExampleNames(t, filepath.Join(root, "examples"))
	got := readmeSectionNames(t, filepath.Join(root, "examples/README.md"), "# examples")
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("examples/README.md example list mismatch\nREADME:\n%s\nprograms:\n%s",
			strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	gotZH := readmeSectionNames(t, filepath.Join(root, "examples/README.zh.md"), "# examples")
	if strings.Join(gotZH, "\n") != strings.Join(want, "\n") {
		t.Fatalf("examples/README.zh.md example list mismatch\nREADME.zh:\n%s\nprograms:\n%s",
			strings.Join(gotZH, "\n"), strings.Join(want, "\n"))
	}

	readme, err := os.ReadFile(filepath.Join(root, "examples/README.md"))
	if err != nil {
		t.Fatalf("read examples/README.md: %v", err)
	}
	for _, want := range []string{"own module", "demo-only dependencies"} {
		if !strings.Contains(string(readme), want) {
			t.Fatalf("examples/README.md does not document %s", want)
		}
	}

	readmeZH, err := os.ReadFile(filepath.Join(root, "examples/README.zh.md"))
	if err != nil {
		t.Fatalf("read examples/README.zh.md: %v", err)
	}
	if !strings.Contains(string(readmeZH), "独立 module") {
		t.Fatal("examples/README.zh.md does not document its independent module boundary")
	}
}

func TestRunnableExamplesHaveStableSmokeOutput(t *testing.T) {
	root := repoRoot(t)
	examplesDir := filepath.Join(root, "examples")
	wants := map[string][]string{
		"cache": {
			"loads after herd: 1 (expected 1)",
			"warm hit reused cached value",
		},
		"concurrent": {
			"processed=20 first=item-0 last=item-9",
		},
		"configlayers": {
			"alpha=eu-west-1/standard admin=alice timeout=500ms",
			"preview admin=\"\" loaded=false total=3",
		},
		"eventbus": {
			"[audit] config changed: app.yaml",
			"[audit] config changed: db.yaml",
			"reloads=2 audits=2 metrics=2",
		},
		"resilientclient": {
			"status=200 body=ok calls=3",
		},
	}
	for _, name := range runnableExampleNames(t, examplesDir) {
		if name == "httpserver" {
			continue
		}
		want, ok := wants[name]
		if !ok {
			t.Fatalf("%s: no stable smoke output expectation", name)
		}
		cmd := exec.Command("go", "run", "./"+name)
		cmd.Dir = examplesDir
		cmd.Env = append(os.Environ(), "GOWORK=off")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("go run ./examples/%s failed: %v\n%s", name, err, out)
		}
		for _, substr := range want {
			if !strings.Contains(string(out), substr) {
				t.Fatalf("go run ./examples/%s output missing %q\n%s", name, substr, out)
			}
		}
	}
}

func TestReadmesDocumentReleaseGate(t *testing.T) {
	root := repoRoot(t)
	for _, name := range []string{"README.md", "README.zh.md"} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(content), "./scripts/check-release.sh") {
			t.Fatalf("%s does not document ./scripts/check-release.sh", name)
		}
	}
}

func TestScriptSelfCheckIsPartOfReleaseGate(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "scripts/check-scripts.sh")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing scripts/check-scripts.sh: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatal("scripts/check-scripts.sh is not executable")
	}

	markdownMustContain(t, path, []string{
		"bash -n",
		"check-ci-cache-paths.sh",
		"list-nested-modules.sh",
		"--help",
		".agents",
	})
	markdownMustContain(t, filepath.Join(root, "scripts/check-ci-cache-paths.sh"), []string{
		"cache-dependency-path: [^|]",
		"cache-dependency-path: \\|",
		".agents",
		"go.sum",
	})
	markdownMustContain(t, filepath.Join(root, "scripts/check-release.sh"), []string{
		"./scripts/check-scripts.sh",
		"./scripts/check-root.sh --full",
		"./scripts/check-split-modules.sh --full",
	})
}

func TestStressGateIsDocumentedAndScripted(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "scripts/check-stress.sh")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing scripts/check-stress.sh: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatal("scripts/check-stress.sh is not executable")
	}
	markdownMustContain(t, path, []string{
		"--quick",
		"--race",
		"go test -race -run 'Stress'",
		"./batcher",
		"./fanout",
		"./keyedmutex",
		"./pool",
		"./ttlmap",
	})
	markdownMustContain(t, filepath.Join(root, "scripts/check-scripts.sh"), []string{
		"scripts/check-stress.sh",
		"./scripts/check-stress.sh --help",
	})

	packages := stressScriptPackages(t, path)
	if strings.Join(packages, "\n") != strings.Join([]string{"batcher", "fanout", "keyedmutex", "pool", "ttlmap"}, "\n") {
		t.Fatalf("unexpected stress package list: %v", packages)
	}
}

func TestGitHubTemplatesCaptureContributionBoundaries(t *testing.T) {
	root := repoRoot(t)
	markdownMustContain(t, filepath.Join(root, ".github/pull_request_template.md"), []string{
		"Change Surface",
		"Public API changed",
		"Root module dependency graph changed",
		"English and Chinese docs updated",
		"./scripts/check-root.sh --quick",
		"./scripts/check-split-modules.sh --quick --module <path>",
	})
	markdownMustContain(t, filepath.Join(root, ".github/ISSUE_TEMPLATE/bug_report.md"), []string{
		"Package or Module",
		"Version or commit",
		"Go version",
		"OS/architecture",
		"Commands run",
	})
	markdownMustContain(t, filepath.Join(root, ".github/ISSUE_TEMPLATE/feature_request.md"), []string{
		"Use Case",
		"Root package, nested module, example, benchmark, or docs",
		"Expected dependency impact",
		"Standard library option",
		"English and Chinese docs affected",
	})
}

func TestReadmesLinkLocalizedSecurityDocs(t *testing.T) {
	root := repoRoot(t)
	for _, path := range []string{
		"zapctx/zapctxgin/README.md",
		"zapctx/zapctxgrpc/README.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(content), "Payload logging") {
			t.Fatalf("%s does not document payload logging risk", path)
		}
	}

	for _, path := range []string{
		"zapctx/zapctxgin/README.zh.md",
		"zapctx/zapctxgrpc/README.zh.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(content), "Payload logging") {
			t.Fatalf("%s does not document payload logging risk", path)
		}
	}
}

func TestMigrationExampleFixturesCompile(t *testing.T) {
	root := repoRoot(t)
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), fmt.Sprintf(`module goostmigrationcheck

go 1.25

require (
	github.com/liguangsheng/goost v0.0.0
	github.com/liguangsheng/goost/zapctx/zapctxgin v0.0.0
	github.com/liguangsheng/goost/zapctx/zapctxgrpc v0.0.0
)

replace github.com/liguangsheng/goost => %s
replace github.com/liguangsheng/goost/zapctx/zapctxgin => %s
replace github.com/liguangsheng/goost/zapctx/zapctxgrpc => %s
`, filepath.ToSlash(root), filepath.ToSlash(filepath.Join(root, "zapctx/zapctxgin")), filepath.ToSlash(filepath.Join(root, "zapctx/zapctxgrpc"))))

	for _, name := range []string{"zapctxgin", "zapctxgrpc"} {
		fixture := filepath.Join(root, "testdata/migration", name, "main.go.txt")
		content, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("read %s: %v", fixture, err)
		}
		pkgDir := filepath.Join(dir, name)
		if err := os.Mkdir(pkgDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", pkgDir, err)
		}
		writeFile(t, filepath.Join(pkgDir, "main.go"), string(content))
	}

	cmd := exec.Command("go", "test", "-mod=mod", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("migration example fixtures failed to compile: %v\n%s", err, out)
	}
}

func TestRootDocReferencesPublicGovernanceDocs(t *testing.T) {
	root := repoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "doc.go"))
	if err != nil {
		t.Fatalf("read doc.go: %v", err)
	}
	for _, want := range []string{"README.md", "examples/"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("doc.go does not reference %s", want)
		}
	}
}

func TestGoVersionPolicyStaysAligned(t *testing.T) {
	root := repoRoot(t)
	want := ciGoVersion(t, filepath.Join(root, ".github/workflows/ci.yml"))
	for _, path := range goModFiles(t, root) {
		got := goModVersion(t, path)
		if got != want {
			rel, _ := filepath.Rel(root, path)
			t.Fatalf("%s declares Go %s, CI declares %s", rel, got, want)
		}
	}

	for _, name := range []string{"README.md", "README.zh.md"} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(content), want) {
			t.Fatalf("%s does not document Go version %s", name, want)
		}
	}
}

func TestCIIncludesWindowsRootSmoke(t *testing.T) {
	root := repoRoot(t)
	markdownMustContain(t, filepath.Join(root, ".github/workflows/ci.yml"), []string{
		"windows-root-smoke",
		"windows-latest",
		"go test ./...",
	})
}

func TestChineseMarkdownLinksLocalizedReleaseDocs(t *testing.T) {
	root := repoRoot(t)
	for _, path := range chineseMarkdownFiles(t, root) {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, stale := range localizedRootDocNames(t, root) {
			if hasMarkdownLinkTo(string(content), stale) {
				rel, _ := filepath.Rel(root, path)
				t.Fatalf("%s links English root doc %s despite localized docs", rel, stale)
			}
		}
	}

}

func TestRootDocsStayLocalizedInPairs(t *testing.T) {
	root := repoRoot(t)
	for _, english := range localizedRootDocNames(t, root) {
		chinese := strings.TrimSuffix(english, ".md") + ".zh.md"
		if _, err := os.Stat(filepath.Join(root, chinese)); err != nil {
			t.Fatalf("%s has no localized pair %s: %v", english, chinese, err)
		}
	}
}

func TestPublicPackageReadmesHaveCompiledExamples(t *testing.T) {
	root := repoRoot(t)
	for _, name := range publicPackageNames(t, root) {
		dir := filepath.Join(root, filepath.FromSlash(name))
		if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
			t.Fatalf("%s: missing README.md: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(dir, "README.zh.md")); err != nil {
			t.Fatalf("%s: missing README.zh.md: %v", name, err)
		}
		if !packageHasCompiledExample(t, dir) {
			t.Fatalf("%s: README-covered public package has no compiled Example test", name)
		}
	}
}

func TestTTLMapReadmesDocumentLifecycle(t *testing.T) {
	root := repoRoot(t)
	markdownMustContain(t, filepath.Join(root, "ttlmap/README.md"), []string{
		"Close",
		"background sweep goroutine",
		"safe to call",
		"still work after `Close`",
		"lazily by `Get`",
		"PurgeExpired",
	})
	markdownMustContain(t, filepath.Join(root, "ttlmap/README.zh.md"), []string{
		"Close",
		"后台 sweep goroutine",
		"可以重复调用",
		"Close` 后仍可使用",
		"懒删除",
		"PurgeExpired",
	})
}

func TestRotatingWriterReadmesDocumentPortability(t *testing.T) {
	root := repoRoot(t)
	markdownMustContain(t, filepath.Join(root, "rotatingwriter/README.md"), []string{
		"Portability",
		"filepath",
		"Windows",
		"permission-bit semantics",
		"pre-create directories or files",
		"ACL policy",
	})
	markdownMustContain(t, filepath.Join(root, "rotatingwriter/README.zh.md"), []string{
		"可移植性",
		"filepath",
		"Windows",
		"permission-bit 语义",
		"预先创建目录或文件",
		"ACL policy",
	})
}

func TestShutdownReadmesDocumentPortability(t *testing.T) {
	root := repoRoot(t)
	markdownMustContain(t, filepath.Join(root, "shutdown/README.md"), []string{
		"Portability",
		"SIGINT",
		"SIGTERM",
		"platform-specific",
		"SIGUSR1",
		"Windows",
		"Cleanup",
		"returns `nil`",
	})
	markdownMustContain(t, filepath.Join(root, "shutdown/README.zh.md"), []string{
		"可移植性",
		"SIGINT",
		"SIGTERM",
		"Unix-only signal",
		"SIGUSR1",
		"Windows",
		"Cleanup",
		"返回 `nil`",
	})
}

func TestZapAndSlogContextReadmesShareConcepts(t *testing.T) {
	root := repoRoot(t)
	for _, path := range []string{
		"zapctx/README.md",
		"slogctx/README.md",
	} {
		markdownMustContain(t, filepath.Join(root, path), []string{
			"Shared Model",
			"ToContext",
			"Extract",
			"AddFields",
			"AddAttrs",
			"Sampled",
			"nested modules",
		})
	}
	for _, path := range []string{
		"zapctx/README.zh.md",
		"slogctx/README.zh.md",
	} {
		markdownMustContain(t, filepath.Join(root, path), []string{
			"共享模型",
			"ToContext",
			"Extract",
			"AddFields",
			"AddAttrs",
			"Sampled",
			"nested modules",
		})
	}
}

func TestObservabilityReadmesDocumentSnapshotSemantics(t *testing.T) {
	root := repoRoot(t)
	for _, path := range []string{
		"batcher/README.md",
		"fanout/README.md",
		"pool/README.md",
		"ratelimit/README.md",
	} {
		markdownMustContain(t, filepath.Join(root, path), []string{
			"point-in-time",
			"read-only",
			"current values",
			"configuration values",
			"metrics labels",
			"logs",
		})
	}
	for _, path := range []string{
		"batcher/README.zh.md",
		"fanout/README.zh.md",
		"pool/README.zh.md",
		"ratelimit/README.zh.md",
	} {
		markdownMustContain(t, filepath.Join(root, path), []string{
			"调用时刻",
			"只读 snapshot",
			"当前值",
			"配置值",
			"metrics label",
			"日志字段",
		})
	}
}

func TestTaskgroupReadmesDocumentPanicBehavior(t *testing.T) {
	root := repoRoot(t)
	markdownMustContain(t, filepath.Join(root, "taskgroup/README.md"), []string{
		"taskgroup: panic:",
		"cancels siblings",
		"returned by `Wait`",
		"Cause()` reports",
		"Results[T]",
		"returns values that completed before cancellation",
	})
	markdownMustContain(t, filepath.Join(root, "taskgroup/README.zh.md"), []string{
		"taskgroup: panic:",
		"取消兄弟任务",
		"由 `Wait`",
		"Cause()`",
		"Results[T]",
		"已经完成的 values",
	})
}

func TestRemovedPackagesStayOutOfActiveDocs(t *testing.T) {
	root := repoRoot(t)
	files := []string{
		"README.md",
		"README.zh.md",
		"doc.go",
		"examples/README.md",
		"examples/README.zh.md",
		"slogctx/README.md",
		"slogctx/README.zh.md",
		"zapctx/README.md",
		"zapctx/README.zh.md",
	}
	removed := []string{"bytesconv", "itertools", "redact", "slogctxotel", "zapctxotel"}
	for _, name := range files {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, pkg := range removed {
			if strings.Contains(string(content), pkg) {
				t.Fatalf("%s still mentions removed package %s", name, pkg)
			}
		}
	}
}

func corePackageImports(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list packages failed: %v\n%s", err, out)
	}

	var imports []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		path := scanner.Text()
		if !strings.HasPrefix(path, modulePath) || isOptionalOrNonCorePackage(path) || isIntegrationPackage(path) {
			continue
		}
		imports = append(imports, path)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan package list: %v", err)
	}
	if len(imports) == 0 {
		t.Fatal("no core packages discovered")
	}
	return imports
}

func basePackageImports(t *testing.T, root string) []string {
	t.Helper()
	var imports []string
	for _, path := range publicPackageImports(t, root) {
		if isIntegrationPackage(path) {
			continue
		}
		imports = append(imports, path)
	}
	if len(imports) == 0 {
		t.Fatal("no base packages discovered")
	}
	return imports
}

func publicPackageImports(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list packages failed: %v\n%s", err, out)
	}

	var imports []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		path := scanner.Text()
		if path == modulePath || !strings.HasPrefix(path, modulePath) || isOptionalOrNonCorePackage(path) {
			continue
		}
		imports = append(imports, path)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan package list: %v", err)
	}
	if len(imports) == 0 {
		t.Fatal("no public packages discovered")
	}
	return imports
}

func publicPackageNames(t *testing.T, root string) []string {
	t.Helper()
	imports := publicPackageImports(t, root)
	names := make([]string, 0, len(imports))
	for _, path := range imports {
		names = append(names, strings.TrimPrefix(path, modulePath+"/"))
	}
	sort.Strings(names)
	return names
}

func optionalModuleNames(t *testing.T, root string) []string {
	t.Helper()
	all := nestedModuleNames(t, root)
	names := make([]string, 0, len(all))
	for _, name := range all {
		if isOptionalIntegrationModule(name) {
			names = append(names, name)
		}
	}
	return names
}

func goModFiles(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	var walk func(string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == ".agents" {
					continue
				}
				walk(path)
				continue
			}
			if entry.Name() == "go.mod" {
				paths = append(paths, path)
			}
		}
	}
	walk(root)
	sort.Strings(paths)
	return paths
}

func goModVersion(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close %s: %v", path, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	t.Fatalf("%s has no go directive", path)
	return ""
}

func ciGoVersion(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close %s: %v", path, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "GO_VERSION:") {
			continue
		}
		_, value, ok := strings.Cut(line, ":")
		if !ok {
			break
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	t.Fatalf("%s has no GO_VERSION", path)
	return ""
}

func nestedModuleNames(t *testing.T, root string) []string {
	t.Helper()
	var names []string
	var walk func(string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if !entry.IsDir() {
				continue
			}
			if entry.Name() == ".git" || entry.Name() == ".agents" {
				continue
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				rel, err := filepath.Rel(root, path)
				if err != nil {
					t.Fatalf("rel %s: %v", path, err)
				}
				name := filepath.ToSlash(rel)
				if name != "." {
					names = append(names, name)
				}
				continue
			}
			walk(path)
		}
	}
	walk(root)
	sort.Strings(names)
	return names
}

func isOptionalIntegrationModule(name string) bool {
	return strings.HasPrefix(name, "zapctx/")
}

func readmePackageNames(t *testing.T, path string) []string {
	t.Helper()
	return readmeSectionNames(t, path, "## Packages", "## 包")
}

func runnableExampleNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name(), "main.go")
		if _, err := os.Stat(path); err == nil {
			names = append(names, entry.Name())
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
	sort.Strings(names)
	return names
}

func markdownMustContain(t *testing.T, path string, wants []string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, want := range wants {
		if !strings.Contains(string(content), want) {
			t.Fatalf("%s does not document %s", path, want)
		}
	}
}

func readmeSectionNames(t *testing.T, path string, headings ...string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close %s: %v", path, err)
		}
	}()

	headingSet := make(map[string]bool, len(headings))
	for _, heading := range headings {
		headingSet[heading] = true
	}

	var names []string
	inPackages := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			inPackages = headingSet[line]
			continue
		}
		if !inPackages {
			continue
		}
		var rest string
		switch {
		case strings.HasPrefix(line, "| [`"):
			rest = strings.TrimPrefix(line, "| [`")
		case strings.HasPrefix(line, "| `"):
			rest = strings.TrimPrefix(line, "| `")
		default:
			continue
		}
		name, _, ok := strings.Cut(rest, "`]")
		if !ok {
			name, _, ok = strings.Cut(rest, "`")
		}
		if !ok {
			continue
		}
		names = append(names, name)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	sort.Strings(names)
	return names
}

func packageHasCompiledExample(t *testing.T, dir string) bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", filepath.Join(dir, entry.Name()), err)
		}
		if strings.Contains(string(content), "func Example") {
			return true
		}
	}
	return false
}

func chineseMarkdownFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	var walk func(string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				if entry.Name() == ".git" {
					continue
				}
				walk(path)
				continue
			}
			if strings.HasSuffix(entry.Name(), ".zh.md") {
				files = append(files, path)
			}
		}
	}
	walk(root)
	sort.Strings(files)
	return files
}

func localizedRootDocNames(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read %s: %v", root, err)
	}

	zhPairs := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zh.md") {
			continue
		}
		english := strings.TrimSuffix(entry.Name(), ".zh.md") + ".md"
		zhPairs[english] = true
	}

	var names []string
	for english := range zhPairs {
		if _, err := os.Stat(filepath.Join(root, english)); err == nil {
			names = append(names, english)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", english, err)
		}
	}
	sort.Strings(names)
	return names
}

func hasMarkdownLinkTo(content, target string) bool {
	for {
		start := strings.Index(content, "](")
		if start < 0 {
			return false
		}
		content = content[start+2:]
		end := strings.IndexByte(content, ')')
		if end < 0 {
			return false
		}
		link := strings.TrimSpace(content[:end])
		if hash := strings.IndexByte(link, '#'); hash >= 0 {
			link = link[:hash]
		}
		if strings.HasSuffix(link, target) {
			return true
		}
		content = content[end+1:]
	}
}

func isOptionalOrNonCorePackage(path string) bool {
	if strings.Contains(path, "/examples/") || strings.Contains(path, "/benchmark") {
		return true
	}
	return false
}

func isIntegrationPackage(path string) bool {
	return path == modulePath+"/zapctx" ||
		path == modulePath+"/zapctx/zapctxgin" ||
		path == modulePath+"/zapctx/zapctxgrpc"
}

func consumerMain(imports []string) string {
	var b strings.Builder
	b.WriteString("package main\n\nimport (\n")
	for _, path := range imports {
		fmt.Fprintf(&b, "\t_ %q\n", path)
	}
	b.WriteString(")\n\nfunc main() {}\n")
	return b.String()
}

func importListContainsPrefix(imports, prefix string) bool {
	scanner := bufio.NewScanner(strings.NewReader(imports))
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), prefix) {
			return true
		}
	}
	return false
}

func stressScriptPackages(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close %s: %v", path, err)
		}
	}()

	var packages []string
	inPackages := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "packages=(" {
			inPackages = true
			continue
		}
		if !inPackages {
			continue
		}
		if line == ")" {
			break
		}
		if strings.HasPrefix(line, "./") {
			packages = append(packages, strings.TrimPrefix(line, "./"))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	if len(packages) == 0 {
		t.Fatalf("%s: no stress packages found", path)
	}
	return packages
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
