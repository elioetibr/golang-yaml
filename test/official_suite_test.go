package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elioetibr/golang-yaml/v0/pkg/parser"
)

// OfficialTestResult tracks results from the official YAML test suite
type OfficialTestResult struct {
	TestID      string `json:"test_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Passed      bool   `json:"passed"`
	Error       string `json:"error,omitempty"`
	Duration    string `json:"duration"`
}

// OfficialSuiteResults aggregates all test results
type OfficialSuiteResults struct {
	Timestamp      string               `json:"timestamp"`
	TotalTests     int                  `json:"total_tests"`
	Passed         int                  `json:"passed"`
	Failed         int                  `json:"failed"`
	PassRate       float64              `json:"pass_rate"`
	Results        []OfficialTestResult `json:"results"`
	FailureSummary map[string]int       `json:"failure_summary"`
}

// TestOfficialYAMLSuite runs the official YAML test suite
func TestOfficialYAMLSuite(t *testing.T) {
	suiteDir := "./yaml-test-suite/src"

	// Check if suite exists
	if _, err := os.Stat(suiteDir); os.IsNotExist(err) {
		t.Skip("Official YAML test suite not found. Clone from https://github.com/yaml/yaml-test-suite")
		return
	}

	// Read all test files
	files, err := ioutil.ReadDir(suiteDir)
	if err != nil {
		t.Fatalf("Failed to read test suite directory: %v", err)
	}

	results := &OfficialSuiteResults{
		Timestamp:      time.Now().Format(time.RFC3339),
		Results:        make([]OfficialTestResult, 0),
		FailureSummary: make(map[string]int),
	}

	// Process each test file
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		testID := strings.TrimSuffix(file.Name(), ".yaml")
		testPath := filepath.Join(suiteDir, file.Name())

		// Run individual test
		result := runOfficialTest(testID, testPath)
		results.Results = append(results.Results, result)
		results.TotalTests++

		if result.Passed {
			results.Passed++
		} else {
			results.Failed++
			categorizeFailure(result.Error, results.FailureSummary)
		}

		// Log progress every 50 tests
		if results.TotalTests%50 == 0 {
			t.Logf("Progress: %d tests processed, %d passed, %d failed",
				results.TotalTests, results.Passed, results.Failed)
		}
	}

	// Calculate pass rate
	if results.TotalTests > 0 {
		results.PassRate = float64(results.Passed) * 100 / float64(results.TotalTests)
	}

	// Generate report
	generateOfficialReport(t, results)

	// Summary
	t.Logf("\n=== Official YAML Test Suite Results ===")
	t.Logf("Total Tests: %d", results.TotalTests)
	t.Logf("Passed: %d (%.1f%%)", results.Passed, results.PassRate)
	t.Logf("Failed: %d", results.Failed)

	// Show failure categories
	if len(results.FailureSummary) > 0 {
		t.Logf("\n=== Failure Categories ===")
		for category, count := range results.FailureSummary {
			t.Logf("%s: %d", category, count)
		}
	}

	// Determine if we meet minimum compliance
	minCompliance := 80.0 // 80% pass rate for acceptance
	if results.PassRate < minCompliance {
		t.Errorf("Pass rate %.1f%% is below minimum compliance of %.1f%%",
			results.PassRate, minCompliance)
	}
}

// runOfficialTest runs a single test from the suite
func runOfficialTest(testID, testPath string) OfficialTestResult {
	start := time.Now()
	result := OfficialTestResult{
		TestID: testID,
		Name:   filepath.Base(testPath),
	}

	// Read test file
	data, err := ioutil.ReadFile(testPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read test file: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}

	// Extract test description from comments if available
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			result.Description = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Try to parse the YAML
	_, parseErr := parser.ParseString(string(data))

	// Check if this is an error test (should fail)
	isErrorTest := strings.Contains(testID, "err") ||
		strings.Contains(testID, "invalid") ||
		strings.Contains(strings.ToLower(result.Description), "error") ||
		strings.Contains(strings.ToLower(result.Description), "invalid")

	if isErrorTest {
		// Error tests should fail parsing
		if parseErr != nil {
			result.Passed = true
		} else {
			result.Error = "Expected parse error but succeeded"
		}
	} else {
		// Valid tests should parse successfully
		if parseErr == nil {
			result.Passed = true
		} else {
			result.Error = fmt.Sprintf("Parse failed: %v", parseErr)
		}
	}

	result.Duration = time.Since(start).String()
	return result
}

// categorizeFailure categorizes the failure for summary
func categorizeFailure(errorMsg string, summary map[string]int) {
	switch {
	case strings.Contains(errorMsg, "tab"):
		summary["Tab-related"]++
	case strings.Contains(errorMsg, "indent"):
		summary["Indentation"]++
	case strings.Contains(errorMsg, "anchor"):
		summary["Anchor/Alias"]++
	case strings.Contains(errorMsg, "tag"):
		summary["Tags"]++
	case strings.Contains(errorMsg, "comment"):
		summary["Comments"]++
	case strings.Contains(errorMsg, "scalar"):
		summary["Scalars"]++
	case strings.Contains(errorMsg, "flow"):
		summary["Flow style"]++
	case strings.Contains(errorMsg, "block"):
		summary["Block style"]++
	case strings.Contains(errorMsg, "document"):
		summary["Documents"]++
	case strings.Contains(errorMsg, "directive"):
		summary["Directives"]++
	case strings.Contains(errorMsg, "Expected parse error"):
		summary["False positive"]++
	default:
		summary["Other"]++
	}
}

// generateOfficialReport generates a detailed compliance report
func generateOfficialReport(t *testing.T, results *OfficialSuiteResults) {
	// Generate JSON report
	reportPath := "official-suite-report.json"
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Logf("Failed to marshal report: %v", err)
		return
	}

	if err := ioutil.WriteFile(reportPath, data, 0644); err != nil {
		t.Logf("Failed to write report: %v", err)
		return
	}

	t.Logf("Compliance report written to %s", reportPath)

	// Also generate a markdown summary
	mdReport := generateMarkdownReport(results)
	mdPath := "official-suite-report.md"
	if err := ioutil.WriteFile(mdPath, []byte(mdReport), 0644); err != nil {
		t.Logf("Failed to write markdown report: %v", err)
		return
	}

	t.Logf("Markdown report written to %s", mdPath)
}

// generateMarkdownReport creates a markdown summary
func generateMarkdownReport(results *OfficialSuiteResults) string {
	var sb strings.Builder

	sb.WriteString("# Official YAML Test Suite Results\n\n")
	sb.WriteString(fmt.Sprintf("**Date**: %s\n\n", results.Timestamp))
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tests**: %d\n", results.TotalTests))
	sb.WriteString(fmt.Sprintf("- **Passed**: %d (%.1f%%)\n", results.Passed, results.PassRate))
	sb.WriteString(fmt.Sprintf("- **Failed**: %d\n\n", results.Failed))

	if len(results.FailureSummary) > 0 {
		sb.WriteString("## Failure Categories\n\n")
		sb.WriteString("| Category | Count |\n")
		sb.WriteString("|----------|-------|\n")
		for category, count := range results.FailureSummary {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", category, count))
		}
		sb.WriteString("\n")
	}

	// List first 20 failures for investigation
	sb.WriteString("## Sample Failures\n\n")
	failCount := 0
	for _, result := range results.Results {
		if !result.Passed && failCount < 20 {
			sb.WriteString(fmt.Sprintf("### %s\n", result.TestID))
			if result.Description != "" {
				sb.WriteString(fmt.Sprintf("**Description**: %s\n", result.Description))
			}
			sb.WriteString(fmt.Sprintf("**Error**: %s\n\n", result.Error))
			failCount++
		}
	}

	sb.WriteString("## Compliance Assessment\n\n")
	if results.PassRate >= 90 {
		sb.WriteString("✅ **Excellent**: Library achieves >90% compliance with official test suite.\n")
	} else if results.PassRate >= 80 {
		sb.WriteString("⚠️ **Good**: Library achieves 80-90% compliance with official test suite.\n")
	} else if results.PassRate >= 70 {
		sb.WriteString("⚠️ **Fair**: Library achieves 70-80% compliance with official test suite.\n")
	} else {
		sb.WriteString("❌ **Needs Improvement**: Library achieves <70% compliance with official test suite.\n")
	}

	return sb.String()
}

// TestSpecificYAMLFeatures tests specific YAML features against the suite
func TestSpecificYAMLFeatures(t *testing.T) {
	suiteDir := "./yaml-test-suite/src"

	if _, err := os.Stat(suiteDir); os.IsNotExist(err) {
		t.Skip("Official YAML test suite not found")
		return
	}

	// Test specific features
	features := map[string][]string{
		"Anchors":   {"26DV", "2AUY", "2JQS"},
		"Tags":      {"27NA", "2CMS", "3MYT"},
		"Multiline": {"2LFX", "2SXE", "3GZX"},
		"Flow":      {"2XXW", "36F6", "3HFZ"},
	}

	for feature, testIDs := range features {
		t.Run(feature, func(t *testing.T) {
			passed := 0
			failed := 0

			for _, id := range testIDs {
				testPath := filepath.Join(suiteDir, id+".yaml")
				if _, err := os.Stat(testPath); os.IsNotExist(err) {
					continue
				}

				result := runOfficialTest(id, testPath)
				if result.Passed {
					passed++
				} else {
					failed++
					t.Logf("Test %s failed: %s", id, result.Error)
				}
			}

			t.Logf("%s: %d passed, %d failed", feature, passed, failed)
		})
	}
}

// BenchmarkOfficialSuite benchmarks parsing of suite files
func BenchmarkOfficialSuite(b *testing.B) {
	suiteDir := "./yaml-test-suite/src"

	if _, err := os.Stat(suiteDir); os.IsNotExist(err) {
		b.Skip("Official YAML test suite not found")
		return
	}

	// Read all test files
	files, _ := ioutil.ReadDir(suiteDir)
	var testData [][]byte

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") {
			data, err := ioutil.ReadFile(filepath.Join(suiteDir, file.Name()))
			if err == nil {
				testData = append(testData, data)
			}
		}

		// Limit to 100 files for benchmark
		if len(testData) >= 100 {
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range testData {
			_, _ = parser.ParseString(string(data))
		}
	}
}
