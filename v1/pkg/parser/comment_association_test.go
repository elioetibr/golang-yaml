package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elioetibr/golang-yaml/v1/pkg/lexer"
	"github.com/elioetibr/golang-yaml/v1/pkg/node"
)

func TestCommentAssociationWithMappingKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, n node.Node)
	}{
		{
			name: "head comments before key",
			input: `# Comment before key
key: value`,
			validate: func(t *testing.T, n node.Node) {
				mapNode, ok := n.(*node.MappingNode)
				require.True(t, ok, "expected MappingNode")
				require.Len(t, mapNode.Pairs, 1)

				key := mapNode.Pairs[0].Key
				scalarKey, ok := key.(*node.ScalarNode)
				require.True(t, ok, "key should be ScalarNode")
				require.NotNil(t, scalarKey.HeadComment, "expected head comment on key")
				assert.Contains(t, scalarKey.HeadComment.Comments[0], "Comment before key")
			},
		},
		{
			name: "inline comment on same line as value",
			input: `key: value # Inline comment`,
			validate: func(t *testing.T, n node.Node) {
				mapNode, ok := n.(*node.MappingNode)
				require.True(t, ok, "expected MappingNode")
				require.Len(t, mapNode.Pairs, 1)
				
				value := mapNode.Pairs[0].Value
				scalarValue, ok := value.(*node.ScalarNode)
				require.True(t, ok, "value should be ScalarNode")
				require.NotNil(t, scalarValue.LineComment, "expected line comment on value")
				assert.Contains(t, scalarValue.LineComment.Comments[0], "Inline comment")
			},
		},
		{
			name: "foot comment after value",
			input: `key: value
# Comment after value`,
			validate: func(t *testing.T, n node.Node) {
				mapNode, ok := n.(*node.MappingNode)
				require.True(t, ok, "expected MappingNode")
				require.Len(t, mapNode.Pairs, 1)
				
				value := mapNode.Pairs[0].Value
				scalarValue, ok := value.(*node.ScalarNode)
				require.True(t, ok, "value should be ScalarNode")
				require.NotNil(t, scalarValue.FootComment, "expected foot comment on value")
				assert.Contains(t, scalarValue.FootComment.Comments[0], "Comment after value")
			},
		},
		{
			name: "multiple comments with different keys",
			input: `# Comment for key1
key1: value1 # Inline for value1
# Comment for key2
key2: value2`,
			validate: func(t *testing.T, n node.Node) {
				mapNode, ok := n.(*node.MappingNode)
				require.True(t, ok, "expected MappingNode")
				require.Len(t, mapNode.Pairs, 2)
				
				// First key should have head comment
				key1 := mapNode.Pairs[0].Key
				scalarKey1, ok := key1.(*node.ScalarNode)
				require.True(t, ok, "key1 should be ScalarNode")
				require.NotNil(t, scalarKey1.HeadComment, "expected head comment on key1")
				assert.Contains(t, scalarKey1.HeadComment.Comments[0], "Comment for key1")
				
				// First value should have inline comment
				value1 := mapNode.Pairs[0].Value
				scalarValue1, ok := value1.(*node.ScalarNode)
				require.True(t, ok, "value1 should be ScalarNode")
				require.NotNil(t, scalarValue1.LineComment, "expected line comment on value1")
				assert.Contains(t, scalarValue1.LineComment.Comments[0], "Inline for value1")
				
				// Second key should have head comment
				key2 := mapNode.Pairs[1].Key
				scalarKey2, ok := key2.(*node.ScalarNode)
				require.True(t, ok, "key2 should be ScalarNode")
				require.NotNil(t, scalarKey2.HeadComment, "expected head comment on key2")
				assert.Contains(t, scalarKey2.HeadComment.Comments[0], "Comment for key2")
			},
		},
		{
			name: "complex nested structure with comments",
			input: `# Root comment
configMap: # ConfigMap for application
  # Properties file comment
  application.properties:
    auth.client_uuid: 0d0b1d94 # UUID for authentication
# Environment variables section  
envVars:
  JAVA_OPTS: "-Xms100m" # JVM options`,
			validate: func(t *testing.T, n node.Node) {
				mapNode, ok := n.(*node.MappingNode)
				require.True(t, ok, "expected MappingNode")
				require.Len(t, mapNode.Pairs, 2)
				
				// configMap key should have head comment
				configMapKey := mapNode.Pairs[0].Key
				scalarConfigMapKey, ok := configMapKey.(*node.ScalarNode)
				require.True(t, ok, "configMapKey should be ScalarNode")
				require.NotNil(t, scalarConfigMapKey.HeadComment, "expected head comment on configMap")
				assert.Contains(t, scalarConfigMapKey.HeadComment.Comments[0], "Root comment")
				
				// configMap value should have line comment
				configMapValue := mapNode.Pairs[0].Value
				mappingConfigMapValue, ok := configMapValue.(*node.MappingNode)
				require.True(t, ok, "configMapValue should be MappingNode")
				require.NotNil(t, mappingConfigMapValue.LineComment, "expected line comment on configMap value")
				assert.Contains(t, mappingConfigMapValue.LineComment.Comments[0], "ConfigMap for application")
				
				// Check nested structure
				require.Greater(t, len(mappingConfigMapValue.Pairs), 0)
				
				// envVars key should have head comment
				envVarsKey := mapNode.Pairs[1].Key
				scalarEnvVarsKey, ok := envVarsKey.(*node.ScalarNode)
				require.True(t, ok, "envVarsKey should be ScalarNode")
				require.NotNil(t, scalarEnvVarsKey.HeadComment, "expected head comment on envVars")
				assert.Contains(t, scalarEnvVarsKey.HeadComment.Comments[0], "Environment variables section")
			},
		},
		{
			name: "empty mapping with comments should use block style",
			input: `# Comment before empty mapping
configMap: {}`,
			validate: func(t *testing.T, n node.Node) {
				mapNode, ok := n.(*node.MappingNode)
				require.True(t, ok, "expected MappingNode")
				require.Len(t, mapNode.Pairs, 1)
				
				// Check that key has head comment
				key := mapNode.Pairs[0].Key
				scalarKey, ok := key.(*node.ScalarNode)
				require.True(t, ok, "key should be ScalarNode")
				require.NotNil(t, scalarKey.HeadComment, "expected head comment on key")
				assert.Contains(t, scalarKey.HeadComment.Comments[0], "Comment before empty mapping")
				
				// Value should be a MappingNode with block style, not flow
				valueMap, ok := mapNode.Pairs[0].Value.(*node.MappingNode)
				require.True(t, ok, "expected value to be MappingNode")
				assert.Equal(t, node.StyleBlock, valueMap.Style, "empty mapping should use block style, not flow")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexerFromString(tt.input)
			p := NewParser(l)
			n, err := p.Parse()
			require.NoError(t, err)
			require.NotNil(t, n)
			
			tt.validate(t, n)
		})
	}
}

func TestCommentPreservationInParser(t *testing.T) {
	input := strings.TrimSpace(`
# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

# Auto-scaling configuration
autoScaling:
  enabled: false # Whether to enable auto-scaling
  maxReplicas: 1
  minReplicas: 1

# ConfigMap for application configuration
configMap:
  # Application properties file
  application.properties:
    auth.client_uuid: 0d0b1d94-4971-43ef-bdaa-8108ac1ffc55
    auth.service_uri: https://auth/authorize_client

# Environment variables for the container
envVars:
  JAVA_OPTS: '-XX:NativeMemoryTracking=summary -Xms100m -Xmx100m'
`)

	l := lexer.NewLexerFromString(input)
	p := NewParser(l)
	n, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, n)
	
	mapNode, ok := n.(*node.MappingNode)
	require.True(t, ok, "expected MappingNode at root")
	
	// Should have document head comments
	require.True(t, mapNode.HasDocumentHeadComments, "should have document head comments flag")
	require.NotNil(t, mapNode.HeadComment, "should have document head comments")
	assert.Contains(t, mapNode.HeadComment.Comments[0], "yaml-language-server")
	
	// Find autoScaling key and check its comments
	var autoScalingPair *node.MappingPair
	for _, pair := range mapNode.Pairs {
		if scalar, ok := pair.Key.(*node.ScalarNode); ok && scalar.Value == "autoScaling" {
			autoScalingPair = pair
			break
		}
	}
	require.NotNil(t, autoScalingPair, "should find autoScaling key")
	
	// autoScaling key should have head comment
	scalarAutoScalingKey, ok := autoScalingPair.Key.(*node.ScalarNode)
	require.True(t, ok, "autoScaling key should be ScalarNode")
	require.NotNil(t, scalarAutoScalingKey.HeadComment, "autoScaling key should have head comment")
	assert.Contains(t, scalarAutoScalingKey.HeadComment.Comments[0], "Auto-scaling configuration")
	
	// Check autoScaling value structure
	autoScalingValue, ok := autoScalingPair.Value.(*node.MappingNode)
	require.True(t, ok, "autoScaling value should be MappingNode")
	
	// Find enabled field in autoScaling
	var enabledPair *node.MappingPair
	for _, pair := range autoScalingValue.Pairs {
		if scalar, ok := pair.Key.(*node.ScalarNode); ok && scalar.Value == "enabled" {
			enabledPair = pair
			break
		}
	}
	require.NotNil(t, enabledPair, "should find enabled key")
	
	// enabled value should have inline comment
	enabledValue := enabledPair.Value
	scalarEnabledValue, ok := enabledValue.(*node.ScalarNode)
	require.True(t, ok, "enabled value should be ScalarNode")
	require.NotNil(t, scalarEnabledValue.LineComment, "enabled value should have line comment")
	assert.Contains(t, scalarEnabledValue.LineComment.Comments[0], "Whether to enable auto-scaling")
	
	// Find configMap and check its structure
	var configMapPair *node.MappingPair
	for _, pair := range mapNode.Pairs {
		if scalar, ok := pair.Key.(*node.ScalarNode); ok && scalar.Value == "configMap" {
			configMapPair = pair
			break
		}
	}
	require.NotNil(t, configMapPair, "should find configMap key")
	
	// configMap key should have head comment
	scalarConfigMapKey, ok := configMapPair.Key.(*node.ScalarNode)
	require.True(t, ok, "configMap key should be ScalarNode")
	require.NotNil(t, scalarConfigMapKey.HeadComment, "configMap key should have head comment")
	assert.Contains(t, scalarConfigMapKey.HeadComment.Comments[0], "ConfigMap for application configuration")
	
	// configMap value should be MappingNode with block style
	configMapValue, ok := configMapPair.Value.(*node.MappingNode)
	require.True(t, ok, "configMap value should be MappingNode")
	assert.Equal(t, node.StyleBlock, configMapValue.Style, "configMap should use block style")
}
