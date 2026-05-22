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

func TestChineseMarkdownLinksLocalizedReleaseDocs(t *testing.T) {
	root := repoRoot(t)
	for _, path := range chineseMarkdownFiles(t, root) {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, stale := range []string{"README.md", "CHANGELOG.md", "MIGRATION.md"} {
			if hasMarkdownLinkTo(string(content), stale) {
				rel, _ := filepath.Rel(root, path)
				t.Fatalf("%s links English root doc %s despite localized docs", rel, stale)
			}
		}
	}

	readmeZH, err := os.ReadFile(filepath.Join(root, "README.zh.md"))
	if err != nil {
		t.Fatalf("read README.zh.md: %v", err)
	}
	for _, want := range []string{"CHANGELOG.zh.md", "MIGRATION.zh.md"} {
		if !strings.Contains(string(readmeZH), want) {
			t.Fatalf("README.zh.md does not link %s", want)
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
		if !packageHasCompiledExample(t, dir) {
			t.Fatalf("%s: README-covered public package has no compiled Example test", name)
		}
	}
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

func readmePackageNames(t *testing.T, path string) []string {
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

	var names []string
	inPackages := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			inPackages = line == "## Packages" || line == "## 包"
			continue
		}
		if !inPackages {
			continue
		}
		if !strings.HasPrefix(line, "| [`") {
			continue
		}
		rest := strings.TrimPrefix(line, "| [`")
		name, _, ok := strings.Cut(rest, "`]")
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
