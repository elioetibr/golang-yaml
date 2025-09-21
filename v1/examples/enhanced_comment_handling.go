package main

import (
	"fmt"
	"strings"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
	"github.com/elioetibr/golang-yaml/v1/pkg/parser"
	"github.com/elioetibr/golang-yaml/v1/pkg/serializer"
)

func main() {
	// Example YAML with complex comment structure
	yamlContent := `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# @schema
# additionalProperties: false
# @schema
# -- Kubernetes deployment strategy for managing pod updates
strategy:
  # -- Strategy type: RollingUpdate or Recreate
  type: RollingUpdate

  # -- Rolling Update Configuration
  rollingUpdate:
    # -- Maximum extra pods during update
    maxSurge: 1
    # -- Maximum unavailable pods
    maxUnavailable: 0


# @schema
# additionalProperties: true
# @schema
# -- Pod Disruption Budget configuration
podDisruptionBudget:
  # -- Enable Pod Disruption Budget
  enabled: false
  # -- Maximum pods unavailable during disruptions
  maxUnavailable: 1
  # Alternative: minAvailable: 1


# @schema
# additionalProperties: true
# @schema
# -- List of image pull secrets
imagePullSecrets: []
# Example:
# imagePullSecrets:
#   - name: aws-ecr-secret`

	fmt.Println("=== Enhanced YAML Parser Demo ===\n")

	// Create parser with enhanced options
	opts := &parser.ParserOptions{
		PreserveComments:            true,
		PreserveEmptyLines:          true,
		KeepSectionBoundaries:       true,
		DefaultLinesBetweenSections: 1,
		AutoDetectSections:          true,
		MergeAdjacentComments:       true,
	}

	// Parse the YAML
	l := lexer.NewLexer(strings.NewReader(yamlContent))
	p := parser.NewEnhancedParser(l, opts)

	root, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Println("✓ Successfully parsed YAML with enhanced comment handling\n")

	// Analyze the parsed structure
	analyzeNode(root, 0)

	// Demonstrate serialization with preserved comments
	fmt.Println("\n=== Serialization with Preserved Comments ===\n")

	s := serializer.NewSerializer()
	output, err := s.Serialize(root)
	if err != nil {
		fmt.Printf("Serialization error: %v\n", err)
		return
	}

	fmt.Println(output)

	// Demonstrate merge with comment preservation
	fmt.Println("\n=== Merge Example with Comment Preservation ===\n")

	overlayYAML := `# Override configuration
strategy:
  type: Recreate  # Changed from RollingUpdate

# New configuration
resources:
  # Resource limits
  limits:
    cpu: 1000m
    memory: 512Mi
  # Resource requests
  requests:
    cpu: 100m
    memory: 128Mi`

	l2 := lexer.NewLexer(strings.NewReader(overlayYAML))
	p2 := parser.NewEnhancedParser(l2, opts)
	overlay, err := p2.Parse()
	if err != nil {
		fmt.Printf("Parse overlay error: %v\n", err)
		return
	}

	// Here you would merge the nodes - simplified for demo
	fmt.Println("✓ Successfully parsed overlay with comments preserved")
}

// analyzeNode recursively analyzes the node structure
func analyzeNode(n node.Node, depth int) {
	if n == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	switch node := n.(type) {
	case *node.DocumentNode:
		fmt.Printf("%sDocument Node:\n", indent)
		if node.HeadComment != nil && len(node.HeadComment.Comments) > 0 {
			fmt.Printf("%s  Document Comments: %d lines\n", indent, len(node.HeadComment.Comments))
			for i, comment := range node.HeadComment.Comments {
				if comment == "##EMPTY_LINE##" {
					fmt.Printf("%s    [%d] (empty line)\n", indent, i+1)
				} else {
					fmt.Printf("%s    [%d] %s\n", indent, i+1, truncate(comment, 50))
				}
			}
		}

		if len(node.Sections) > 0 {
			fmt.Printf("%s  Sections: %d detected\n", indent, len(node.Sections))
			for _, section := range node.Sections {
				fmt.Printf("%s    - Section '%s' (Type: %s)\n", indent, section.ID, section.Type)
			}
		}

		if node.Content != nil {
			analyzeNode(node.Content, depth+1)
		}

		for _, child := range node.Nodes {
			analyzeNode(child, depth+1)
		}

	case *node.MappingNode:
		fmt.Printf("%sMapping Node:\n", indent)

		if node.HeadComment != nil && len(node.HeadComment.Comments) > 0 {
			fmt.Printf("%s  Header Comments: %d\n", indent, len(node.HeadComment.Comments))
		}

		if node.Section() != nil {
			fmt.Printf("%s  Section: %s\n", indent, node.Section().ID)
		}

		for _, pair := range node.Pairs {
			if keyNode, ok := pair.Key.(*node.ScalarNode); ok {
				fmt.Printf("%s  Key: '%s'\n", indent, keyNode.Value)

				// Check for comments on the key
				if pair.KeyComment != nil && len(pair.KeyComment.Comments) > 0 {
					fmt.Printf("%s    Key Comments: %d\n", indent, len(pair.KeyComment.Comments))
				}
			}

			if pair.Value != nil {
				analyzeNode(pair.Value, depth+2)
			}
		}

	case *node.ScalarNode:
		fmt.Printf("%sScalar: '%s'\n", indent, node.Value)

		if node.HeadComment != nil && len(node.HeadComment.Comments) > 0 {
			fmt.Printf("%s  Header Comments: %d\n", indent, len(node.HeadComment.Comments))
		}

		if node.LineComment != nil && len(node.LineComment.Comments) > 0 {
			fmt.Printf("%s  Inline Comment: '%s'\n", indent, node.LineComment.Comments[0])
		}

		if node.FootComment != nil && len(node.FootComment.Comments) > 0 {
			fmt.Printf("%s  Footer Comments: %d\n", indent, len(node.FootComment.Comments))
		}

	case *node.SequenceNode:
		fmt.Printf("%sSequence Node (%d items):\n", indent, len(node.Items))

		if node.HeadComment != nil && len(node.HeadComment.Comments) > 0 {
			fmt.Printf("%s  Header Comments: %d\n", indent, len(node.HeadComment.Comments))
		}

		for i, item := range node.Items {
			fmt.Printf("%s  [%d]:\n", indent, i)
			analyzeNode(item, depth+2)
		}

	case *node.SectionNode:
		section := node.GetSection()
		fmt.Printf("%sSection: %s (Type: %s)\n", indent, section.ID, section.Type)
		if section.Title != "" {
			fmt.Printf("%s  Title: %s\n", indent, section.Title)
		}
	}
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}