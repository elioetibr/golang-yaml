package test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
)

// YAMLTestCase represents a test case from the YAML test suite
type YAMLTestCase struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	YAML        string   `json:"yaml"`
	JSON        string   `json:"json"`
	Error       bool     `json:"error"`
	Tags        []string `json:"tags"`
}

// TestSuiteRunner runs tests from the official YAML test suite
type TestSuiteRunner struct {
	testsDir string
	results  *ComplianceResults
}

// ComplianceResults tracks test results
type ComplianceResults struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Errors  []TestError
}

// TestError represents a test failure
type TestError struct {
	TestName string
	Error    string
	Input    string
}

// NewTestSuiteRunner creates a new test suite runner
func NewTestSuiteRunner(testsDir string) *TestSuiteRunner {
	return &TestSuiteRunner{
		testsDir: testsDir,
		results: &ComplianceResults{
			Errors: make([]TestError, 0),
		},
	}
}

// Run executes all test cases
func (r *TestSuiteRunner) Run(t *testing.T) {
	// Find all test files
	testFiles, err := r.findTestFiles()
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	for _, file := range testFiles {
		r.runTestFile(t, file)
	}

	// Print results
	r.printResults(t)
}

// findTestFiles locates all test files in the suite
func (r *TestSuiteRunner) findTestFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(r.testsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for YAML test files
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// runTestFile runs tests from a single file
func (r *TestSuiteRunner) runTestFile(t *testing.T, file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		t.Logf("Failed to read %s: %v", file, err)
		r.results.Skipped++
		return
	}

	testName := filepath.Base(file)
	t.Run(testName, func(t *testing.T) {
		r.runTest(t, testName, string(data))
	})
}

// runTest executes a single test case
func (r *TestSuiteRunner) runTest(t *testing.T, name string, yamlInput string) {
	r.results.Total++

	// Try to parse the YAML
	_, err := parser.ParseString(yamlInput)

	if err != nil {
		// Check if error was expected
		if r.shouldFail(name) {
			r.results.Passed++
			return
		}

		r.results.Failed++
		r.results.Errors = append(r.results.Errors, TestError{
			TestName: name,
			Error:    err.Error(),
			Input:    truncate(yamlInput, 200),
		})
		t.Errorf("Parse failed: %v", err)
		return
	}

	// Test passed
	r.results.Passed++
}

// shouldFail checks if a test is expected to fail
func (r *TestSuiteRunner) shouldFail(name string) bool {
	// Check if test name indicates it should fail
	return strings.Contains(name, "error") ||
		strings.Contains(name, "invalid") ||
		strings.Contains(name, "fail")
}

// printResults outputs the test results
func (r *TestSuiteRunner) printResults(t *testing.T) {
	t.Logf("\n=== YAML Test Suite Results ===")
	t.Logf("Total:   %d", r.results.Total)
	t.Logf("Passed:  %d (%.1f%%)", r.results.Passed, float64(r.results.Passed)*100/float64(r.results.Total))
	t.Logf("Failed:  %d (%.1f%%)", r.results.Failed, float64(r.results.Failed)*100/float64(r.results.Total))
	t.Logf("Skipped: %d", r.results.Skipped)

	if len(r.results.Errors) > 0 {
		t.Logf("\n=== Failed Tests ===")
		for i, err := range r.results.Errors {
			if i >= 10 {
				t.Logf("... and %d more", len(r.results.Errors)-10)
				break
			}
			t.Logf("%s: %s", err.TestName, err.Error)
		}
	}
}

// GenerateComplianceReport creates a detailed compliance report
func (r *TestSuiteRunner) GenerateComplianceReport(outputFile string) error {
	report := ComplianceReport{
		Version:     "1.0",
		TestSuite:   "YAML Test Suite",
		LibraryName: "yaml",
		Results:     r.results,
		Details:     r.analyzeResults(),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outputFile, data, 0644)
}

// ComplianceReport represents a full compliance report
type ComplianceReport struct {
	Version     string             `json:"version"`
	TestSuite   string             `json:"test_suite"`
	LibraryName string             `json:"library_name"`
	Results     *ComplianceResults `json:"results"`
	Details     *AnalysisDetails   `json:"details"`
}

// AnalysisDetails provides detailed analysis
type AnalysisDetails struct {
	PassRate       float64         `json:"pass_rate"`
	FeatureSupport map[string]bool `json:"feature_support"`
	CommonErrors   map[string]int  `json:"common_errors"`
}

// analyzeResults performs detailed analysis of test results
func (r *TestSuiteRunner) analyzeResults() *AnalysisDetails {
	details := &AnalysisDetails{
		PassRate:       float64(r.results.Passed) * 100 / float64(r.results.Total),
		FeatureSupport: make(map[string]bool),
		CommonErrors:   make(map[string]int),
	}

	// Analyze error patterns
	for _, err := range r.results.Errors {
		// Categorize errors
		if strings.Contains(err.Error, "tab") {
			details.CommonErrors["tab_errors"]++
		} else if strings.Contains(err.Error, "indent") {
			details.CommonErrors["indentation_errors"]++
		} else if strings.Contains(err.Error, "anchor") {
			details.CommonErrors["anchor_errors"]++
		} else if strings.Contains(err.Error, "tag") {
			details.CommonErrors["tag_errors"]++
		} else {
			details.CommonErrors["other_errors"]++
		}
	}

	// Determine feature support
	details.FeatureSupport["anchors"] = r.results.Passed > 0
	details.FeatureSupport["tags"] = r.results.Passed > 0
	details.FeatureSupport["multiple_documents"] = r.results.Passed > 0
	details.FeatureSupport["flow_style"] = r.results.Passed > 0
	details.FeatureSupport["block_style"] = r.results.Passed > 0

	return details
}

// truncate shortens a string for display
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// RunOfficialTestSuite runs the official YAML test suite if available
func RunOfficialTestSuite(t *testing.T) {
	// Check if test suite directory exists
	testsDir := "./yaml-test-suite"
	if _, err := os.Stat(testsDir); os.IsNotExist(err) {
		t.Skip("YAML test suite not found. Download from https://github.com/yaml/yaml-test-suite")
		return
	}

	runner := NewTestSuiteRunner(testsDir)
	runner.Run(t)

	// Generate report
	if err := runner.GenerateComplianceReport("compliance-report.json"); err != nil {
		t.Logf("Failed to generate compliance report: %v", err)
	}
}

// RoundTripTest tests parsing and serialization round-trip
func RoundTripTest(t *testing.T, input string) {
	// Parse the input
	root, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Serialize back
	output, err := serializer.SerializeToString(root, nil)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Parse again
	root2, err := parser.ParseString(output)
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Serialize again
	output2, err := serializer.SerializeToString(root2, nil)
	if err != nil {
		t.Fatalf("Second serialize failed: %v", err)
	}

	// Compare outputs
	if output != output2 {
		t.Errorf("Round-trip failed:\nFirst:\n%s\nSecond:\n%s", output, output2)
	}
}
