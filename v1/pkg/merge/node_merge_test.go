package merge

import (
	"testing"

	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

// TestNodeMerge tests merging YAML nodes without comments (TDD Case 03)
func TestNodeMerge(t *testing.T) {
	builder := node.NewBuilder()

	// Create base YAML structure
	baseMapping := builder.BuildMapping(nil, node.StyleBlock)

	// Add company field
	companyKey := builder.BuildScalar("company", node.StylePlain)
	companyValue := builder.BuildScalar("Umbrella Corp.", node.StylePlain)
	baseMapping.Pairs = append(baseMapping.Pairs, &node.MappingPair{Key: companyKey, Value: companyValue})

	// Add city field
	cityKey := builder.BuildScalar("city", node.StylePlain)
	cityValue := builder.BuildScalar("Raccoon City", node.StylePlain)
	baseMapping.Pairs = append(baseMapping.Pairs, &node.MappingPair{Key: cityKey, Value: cityValue})

	// Add employees mapping
	employeesKey := builder.BuildScalar("employees", node.StylePlain)
	employeesMapping := builder.BuildMapping(nil, node.StyleBlock)

	// Add Bob
	bobKey := builder.BuildScalar("bob@umbreallacorp.co", node.StylePlain)
	bobMapping := builder.BuildMapping(nil, node.StyleBlock)
	bobNameKey := builder.BuildScalar("name", node.StylePlain)
	bobNameValue := builder.BuildScalar("Bob Sinclair", node.StylePlain)
	bobDeptKey := builder.BuildScalar("department", node.StylePlain)
	bobDeptValue := builder.BuildScalar("Cloud Computing", node.StylePlain)
	bobMapping.Pairs = []*node.MappingPair{
		{Key: bobNameKey, Value: bobNameValue},
		{Key: bobDeptKey, Value: bobDeptValue},
	}
	employeesMapping.Pairs = append(employeesMapping.Pairs, &node.MappingPair{Key: bobKey, Value: bobMapping})

	// Add Alice
	aliceKey := builder.BuildScalar("alice@umbreallacorp.co", node.StylePlain)
	aliceMapping := builder.BuildMapping(nil, node.StyleBlock)
	aliceNameKey := builder.BuildScalar("name", node.StylePlain)
	aliceNameValue := builder.BuildScalar("Alice Abernathy", node.StylePlain)
	aliceDeptKey := builder.BuildScalar("department", node.StylePlain)
	aliceDeptValue := builder.BuildScalar("Project", node.StylePlain)
	aliceMapping.Pairs = []*node.MappingPair{
		{Key: aliceNameKey, Value: aliceNameValue},
		{Key: aliceDeptKey, Value: aliceDeptValue},
	}
	employeesMapping.Pairs = append(employeesMapping.Pairs, &node.MappingPair{Key: aliceKey, Value: aliceMapping})

	baseMapping.Pairs = append(baseMapping.Pairs, &node.MappingPair{Key: employeesKey, Value: employeesMapping})

	// Verify base structure
	if len(baseMapping.Pairs) != 3 {
		t.Errorf("Base mapping should have 3 pairs (company, city, employees), got %d", len(baseMapping.Pairs))
	}

	// Create override YAML structure
	overrideMapping := builder.BuildMapping(nil, node.StyleBlock)

	// Add updated company field
	overrideCompanyKey := builder.BuildScalar("company", node.StylePlain)
	overrideCompanyValue := builder.BuildScalar("Umbrella Corporation.", node.StylePlain)
	overrideMapping.Pairs = append(overrideMapping.Pairs, &node.MappingPair{Key: overrideCompanyKey, Value: overrideCompanyValue})

	// Add city field (same)
	overrideCityKey := builder.BuildScalar("city", node.StylePlain)
	overrideCityValue := builder.BuildScalar("Raccoon City", node.StylePlain)
	overrideMapping.Pairs = append(overrideMapping.Pairs, &node.MappingPair{Key: overrideCityKey, Value: overrideCityValue})

	// Add employees with Red Queen
	overrideEmployeesKey := builder.BuildScalar("employees", node.StylePlain)
	overrideEmployeesMapping := builder.BuildMapping(nil, node.StyleBlock)

	// Add Red Queen
	redQueenKey := builder.BuildScalar("redqueen@umbreallacorp.co", node.StylePlain)
	redQueenMapping := builder.BuildMapping(nil, node.StyleBlock)
	redQueenNameKey := builder.BuildScalar("name", node.StylePlain)
	redQueenNameValue := builder.BuildScalar("Red Queen", node.StylePlain)
	redQueenDeptKey := builder.BuildScalar("department", node.StylePlain)
	redQueenDeptValue := builder.BuildScalar("Security", node.StylePlain)
	redQueenMapping.Pairs = []*node.MappingPair{
		{Key: redQueenNameKey, Value: redQueenNameValue},
		{Key: redQueenDeptKey, Value: redQueenDeptValue},
	}
	overrideEmployeesMapping.Pairs = append(overrideEmployeesMapping.Pairs, &node.MappingPair{Key: redQueenKey, Value: redQueenMapping})

	overrideMapping.Pairs = append(overrideMapping.Pairs, &node.MappingPair{Key: overrideEmployeesKey, Value: overrideEmployeesMapping})

	// Test merge functionality
	mergedNode, err := Merge(baseMapping, overrideMapping)
	if err != nil {
		t.Fatalf("Failed to merge nodes: %v", err)
	}

	// Verify merged structure
	mergedMapping, ok := mergedNode.(*node.MappingNode)
	if !ok {
		t.Fatal("Merged node should be a mapping")
	}

	// Check that company value was updated
	var mergedCompanyValue string
	for _, pair := range mergedMapping.Pairs {
		if keyScalar, ok := pair.Key.(*node.ScalarNode); ok && keyScalar.Value == "company" {
			if valueScalar, ok := pair.Value.(*node.ScalarNode); ok {
				mergedCompanyValue = valueScalar.Value
			}
		}
	}
	if mergedCompanyValue != "Umbrella Corporation." {
		t.Errorf("Company value should be updated to 'Umbrella Corporation.', got %s", mergedCompanyValue)
	}

	// Check that employees contains both original and new employees
	for _, pair := range mergedMapping.Pairs {
		if keyScalar, ok := pair.Key.(*node.ScalarNode); ok && keyScalar.Value == "employees" {
			if employeesNode, ok := pair.Value.(*node.MappingNode); ok {
				// Should have 3 employees: Bob, Alice, and Red Queen
				if len(employeesNode.Pairs) != 3 {
					t.Errorf("Merged employees should have 3 entries, got %d", len(employeesNode.Pairs))
				}

				// Check for specific employees
				hasRedQueen := false
				hasBob := false
				hasAlice := false
				for _, empPair := range employeesNode.Pairs {
					if keyScalar, ok := empPair.Key.(*node.ScalarNode); ok {
						switch keyScalar.Value {
						case "redqueen@umbreallacorp.co":
							hasRedQueen = true
						case "bob@umbreallacorp.co":
							hasBob = true
						case "alice@umbreallacorp.co":
							hasAlice = true
						}
					}
				}

				if !hasRedQueen {
					t.Error("Merged employees should contain Red Queen")
				}
				if !hasBob {
					t.Error("Merged employees should contain Bob")
				}
				if !hasAlice {
					t.Error("Merged employees should contain Alice")
				}
			}
		}
	}

	// Verify override structure is created correctly
	if len(overrideMapping.Pairs) != 3 {
		t.Errorf("Override mapping should have 3 pairs, got %d", len(overrideMapping.Pairs))
	}
}

// TestMergeWithOptions tests various merge strategies
func TestMergeWithOptions(t *testing.T) {
	builder := node.NewBuilder()

	// Create base structure
	baseMapping := builder.BuildMapping(nil, node.StyleBlock)
	configKey := builder.BuildScalar("config", node.StylePlain)
	configMapping := builder.BuildMapping(nil, node.StyleBlock)

	nestedKey := builder.BuildScalar("nested", node.StylePlain)
	nestedMapping := builder.BuildMapping(nil, node.StyleBlock)
	valueKey := builder.BuildScalar("value", node.StylePlain)
	valueScalar := builder.BuildScalar("base", node.StylePlain)
	otherKey := builder.BuildScalar("other", node.StylePlain)
	otherScalar := builder.BuildScalar("keep", node.StylePlain)

	nestedMapping.Pairs = []*node.MappingPair{
		{Key: valueKey, Value: valueScalar},
		{Key: otherKey, Value: otherScalar},
	}
	configMapping.Pairs = []*node.MappingPair{
		{Key: nestedKey, Value: nestedMapping},
	}
	baseMapping.Pairs = []*node.MappingPair{
		{Key: configKey, Value: configMapping},
	}

	// Create override structure
	overrideMapping := builder.BuildMapping(nil, node.StyleBlock)
	overrideConfigKey := builder.BuildScalar("config", node.StylePlain)
	overrideConfigMapping := builder.BuildMapping(nil, node.StyleBlock)

	overrideNestedKey := builder.BuildScalar("nested", node.StylePlain)
	overrideNestedMapping := builder.BuildMapping(nil, node.StyleBlock)
	overrideValueKey := builder.BuildScalar("value", node.StylePlain)
	overrideValueScalar := builder.BuildScalar("override", node.StylePlain)

	overrideNestedMapping.Pairs = []*node.MappingPair{
		{Key: overrideValueKey, Value: overrideValueScalar},
	}
	overrideConfigMapping.Pairs = []*node.MappingPair{
		{Key: overrideNestedKey, Value: overrideNestedMapping},
	}
	overrideMapping.Pairs = []*node.MappingPair{
		{Key: overrideConfigKey, Value: overrideConfigMapping},
	}

	t.Run("deep merge preserves unmodified fields", func(t *testing.T) {
		opts := DefaultOptions().WithStrategy(StrategyDeep)
		result, err := WithOptions(baseMapping, overrideMapping, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that the merge preserved the "other" field from base
		foundOther := false
		foundValue := false
		var mergedValueContent string

		if resultMapping, ok := result.(*node.MappingNode); ok {
			for _, pair := range resultMapping.Pairs {
				if keyScalar, ok := pair.Key.(*node.ScalarNode); ok && keyScalar.Value == "config" {
					if configNode, ok := pair.Value.(*node.MappingNode); ok {
						for _, configPair := range configNode.Pairs {
							if keyScalar, ok := configPair.Key.(*node.ScalarNode); ok && keyScalar.Value == "nested" {
								if nestedNode, ok := configPair.Value.(*node.MappingNode); ok {
									for _, nestedPair := range nestedNode.Pairs {
										if keyScalar, ok := nestedPair.Key.(*node.ScalarNode); ok {
											if keyScalar.Value == "other" {
												foundOther = true
											} else if keyScalar.Value == "value" {
												foundValue = true
												if valueScalar, ok := nestedPair.Value.(*node.ScalarNode); ok {
													mergedValueContent = valueScalar.Value
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		if !foundOther {
			t.Error("Deep merge should preserve unmodified nested fields")
		}
		if !foundValue {
			t.Error("Deep merge should include merged fields")
		}
		if mergedValueContent != "override" {
			t.Errorf("Deep merge should override values, got %s", mergedValueContent)
		}
	})
}
