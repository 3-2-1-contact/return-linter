package main

import (
	"github.com/jc/return-linter"
	"golang.org/x/tools/go/analysis"
)

// AnalyzerPlugin is the entry point for golangci-lint
type AnalyzerPlugin struct{}

// GetAnalyzers returns the analyzers provided by this plugin
func (*AnalyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		returnlinter.Analyzer,
	}
}

// New creates a new instance of the plugin
func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{returnlinter.Analyzer}, nil
}
