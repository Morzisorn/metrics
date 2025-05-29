package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAddStaticAnalyzers(t *testing.T) {
	var analyzers Analyzers

	analyzers.AddStaticAnalyzers()
	analyzersCount := 21
	assert.Equal(t, analyzersCount, len(analyzers.analyzers))
}

func TestAddSAStaticcheckAnalyzers(t *testing.T) {
	var analyzers Analyzers

	analyzers.AddSAStaticcheckAnalyzers()
	analyzersCount := 90
	assert.Equal(t, analyzersCount, len(analyzers.analyzers))
}

func TestAddNotSAStaticcheckAnalyzers(t *testing.T) {
	var analyzers Analyzers

	analyzers.AddNotSAStaticcheckAnalyzers()
	analyzersCount := 3
	assert.Equal(t, analyzersCount, len(analyzers.analyzers))
}

func TestAddThirdPartyAnalyzers(t *testing.T) {
	var analyzers Analyzers

	analyzers.AddThirdPartyAnalyzers()
	analyzersCount := 2
	assert.Equal(t, analyzersCount, len(analyzers.analyzers))
}

func TestAddNoExitAnalyzer(t *testing.T) {
	var analyzers Analyzers

	analyzers.AddNoExitAnalyzer()
	analyzersCount := 1
	assert.Equal(t, analyzersCount, len(analyzers.analyzers))

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, noExitInMainAnalyzer, "a")
}
