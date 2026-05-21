package goost

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		if !strings.HasPrefix(path, modulePath) || isOptionalOrNonCorePackage(path) {
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

func isOptionalOrNonCorePackage(path string) bool {
	if strings.HasPrefix(path, modulePath+"/examples/") {
		return true
	}
	switch path {
	case modulePath + "/lru/benchmark",
		modulePath + "/slogctx/slogctxotel",
		modulePath + "/zapctx/zapctxgin",
		modulePath + "/zapctx/zapctxgrpc",
		modulePath + "/zapctx/zapctxotel":
		return true
	default:
		return false
	}
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
