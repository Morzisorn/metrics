// Command staticlint is a custom static analysis tool built using multichecker from
// golang.org/x/tools/go/analysis. It aggregates multiple analyzers to perform
// comprehensive static analysis of Go codebases.
//
// # Overview
//
// staticlint runs a combination of standard analyzers from the Go analysis framework,
// analyzers from the staticcheck.io toolset, a few additional third-party analyzers,
// and a custom analyzer that checks for forbidden usage of os.Exit in main.main.
//
// This tool is useful for automated code quality enforcement in development
// environments or CI pipelines.
//
// # Execution
//
// To run the tool, use:
//   staticlint ./...
//
// The tool uses multichecker.Main to register and execute the analyzers. Each analyzer
// inspects Go source files and reports potential issues such as bugs, code smells,
// style violations, and unsafe operations.

package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"golang.org/x/tools/go/analysis/passes/nilness"
)

type Analyzers struct {
	analyzers []*analysis.Analyzer
}

// main assembles and registers all analyzers with multichecker. These include:
// - standard analyzers from golang.org/x/tools,
// - SA-class bug detectors from staticcheck,
// - additional staticcheck analyzers for style issues,
// - external analyzers like nilness and simple,
// - and a custom analyzer for restricting os.Exit in main.
func main() {
	var analyzers Analyzers

	analyzers.AddStaticAnalyzers()
	analyzers.AddSAStaticcheckAnalyzers()
	analyzers.AddNotSAStaticcheckAnalyzers()
	analyzers.AddThirdPartyAnalyzers()
	analyzers.AddNoExitAnalyzer()

	multichecker.Main(analyzers.analyzers...)
}

// Standard analyzers from golang.org/x/tools/go/analysis/passes.
func (a *Analyzers) AddStaticAnalyzers() {
	// - asmdecl: reports mismatches between Go declarations and assembly.
	// - assign: detects useless assignments.
	// - atomic: checks for common mistakes using sync/atomic.
	// - bools: detects redundant boolean expressions.
	// - buildtag: reports incorrect build tag usage.
	// - cgocall: detects potentially unsafe cgo calls.
	// - composite: checks for unkeyed composite literals.
	// - copylock: checks for locks passed by value.
	// - httpresponse: detects mistakes in HTTP response handling (e.g., forgetting to close Body).
	// - loopclosure: detects capturing loop variables in goroutines.
	// - lostcancel: reports contexts that are not properly cancelled.
	// - nilfunc: detects calls to nil functions.
	// - printf: validates format strings in Printf-style functions.
	// - shift: detects suspicious bit shift operations.
	// - stdmethods: checks for misspelled method names like String() or Error().
	// - structtag: checks struct field tags for format issues.
	// - tests: checks for common test mistakes.
	// - unmarshal: checks for common JSON unmarshaling mistakes.
	// - unreachable: detects unreachable code.
	// - unsafeptr: detects conversions to unsafe.Pointer that may be invalid.
	// - unusedresult: checks for calls where results should not be ignored.
	a.analyzers = append(a.analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)
}

// Staticcheck analyzers: add all SA-class analyzers (serious bugs),
// and three more from other groups (e.g., S1000) for style/performance.
func (a *Analyzers) AddSAStaticcheckAnalyzers() {
	for _, v := range staticcheck.Analyzers {
		name := v.Analyzer.Name
		if len(name) >= 2 && name[:2] == "SA" {
			a.analyzers = append(a.analyzers, v.Analyzer)
			continue
		}
	}
}

// AddNotSAStaticcheckAnalyzers adds a small curated set of non-SA analyzers from the
// staticcheck.io suite to the analyzer list. These analyzers include stylistic and
// simplification checks that help enforce idiomatic and clean Go code.
func (a *Analyzers) AddNotSAStaticcheckAnalyzers() {
	a.analyzers = append(a.analyzers,
		simple.Analyzers[0].Analyzer,     // S1000: Simplifies boolean expressions like len(s) != 0 → s != ""
		simple.Analyzers[1].Analyzer,     // S1002: Omits redundant boolean comparisons like x == true → x
		stylecheck.Analyzers[5].Analyzer, // ST1005: Ensures that error messages start with lowercase letters
	)
}

// Additional third-party analyzers.
func (a *Analyzers) AddThirdPartyAnalyzers() {
	a.analyzers = append(a.analyzers,
		nilness.Analyzer,
		stylecheck.Analyzers[0].Analyzer,
	)
}

// Custom analyzer that forbids os.Exit in main.main.
func (a *Analyzers) AddNoExitAnalyzer() {
	a.analyzers = append(a.analyzers, noExitInMainAnalyzer)
}

// noExitInMainAnalyzer reports a diagnostic if os.Exit is used inside
// the main function of the main package.
var noExitInMainAnalyzer = &analysis.Analyzer{
	Name: "noexitmain",
	Doc:  "disallows direct call to os.Exit in main.main",
	Run:  runNoExitCheck,
}

// runNoExitCheck walks the AST of main.main and looks for os.Exit calls.
func runNoExitCheck(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, ".go") {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				ce, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := ce.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Exit" {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok || ident.Name != "os" {
					return true
				}

				pass.Reportf(n.Pos(), "direct call to os.Exit in main.main is not allowed")
				return true
			})
		}
	}

	return nil, nil
}
