package merge

import "github.com/elioetibr/golang-yaml/pkg/node"

// NodeProcessor handles common node processing operations
type NodeProcessor struct{}

// NewNodeProcessor creates a new node processor
func NewNodeProcessor() *NodeProcessor {
	return &NodeProcessor{}
}

// CleanScalarHeadComment removes head comments from scalar nodes to prevent formatting issues
// This ensures values-with-comments stay on the same line as their keys
func (p *NodeProcessor) CleanScalarHeadComment(n node.Node) node.Node {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return &node.ScalarNode{
			BaseNode: node.BaseNode{
				TagValue:    scalar.TagValue,
				AnchorValue: scalar.AnchorValue,
				LineComment: scalar.LineComment,
				FootComment: scalar.FootComment,
				// Deliberately not copying HeadComment
			},
			Value: scalar.Value,
			Style: scalar.Style,
			Alias: scalar.Alias,
		}
	}
	return n
}

// PreserveMetadata copies metadata from base to result node
func (p *NodeProcessor) PreserveMetadata(result, base node.Node, opts *Options) {
	// Get base nodes if available
	var resultBase, baseBase *node.BaseNode

	switch r := result.(type) {
	case *node.MappingNode:
		resultBase = &r.BaseNode
	case *node.SequenceNode:
		resultBase = &r.BaseNode
	case *node.ScalarNode:
		resultBase = &r.BaseNode
	}

	switch b := base.(type) {
	case *node.MappingNode:
		baseBase = &b.BaseNode
	case *node.SequenceNode:
		baseBase = &b.BaseNode
	case *node.ScalarNode:
		baseBase = &b.BaseNode
	}

	if resultBase != nil && baseBase != nil {
		if opts.PreserveComments {
			if resultBase.HeadComment == nil {
				resultBase.HeadComment = baseBase.HeadComment
			}
			if resultBase.FootComment == nil {
				resultBase.FootComment = baseBase.FootComment
			}
			if resultBase.LineComment == nil {
				resultBase.LineComment = baseBase.LineComment
			}
		}

		if opts.PreserveBlankLines {
			resultBase.BlankLinesBefore = baseBase.BlankLinesBefore
			resultBase.BlankLinesAfter = baseBase.BlankLinesAfter
		}
	}
}

// CreateMappingPair creates a new mapping pair with preserved metadata
func (p *NodeProcessor) CreateMappingPair(key, value node.Node, basePair *node.MappingPair, opts *Options) *node.MappingPair {
	pair := &node.MappingPair{
		Key:   key,
		Value: value,
	}

	if basePair != nil {
		if opts.PreserveComments {
			pair.KeyComment = basePair.KeyComment
			pair.ValueComment = basePair.ValueComment
		}

		if opts.PreserveBlankLines {
			pair.BlankLinesBefore = basePair.BlankLinesBefore
			pair.BlankLinesAfter = basePair.BlankLinesAfter
		}
	}

	return pair
}

// GetScalarValue safely extracts the value from a scalar node
func (p *NodeProcessor) GetScalarValue(n node.Node) (string, bool) {
	if scalar, ok := n.(*node.ScalarNode); ok {
		return scalar.Value, true
	}
	return "", false
}

// PreserveKeyNode preserves comments from key nodes
func (p *NodeProcessor) PreserveKeyNode(baseKey, overrideKey node.Node, opts *Options) node.Node {
	if !opts.PreserveComments {
		return baseKey
	}

	// If both are scalar nodes, preserve comments from base key
	baseScalar, baseOk := baseKey.(*node.ScalarNode)
	_, overrideOk := overrideKey.(*node.ScalarNode)

	if baseOk && overrideOk {
		// Create new key node that preserves all comments from base
		return &node.ScalarNode{
			BaseNode: node.BaseNode{
				HeadComment:      baseScalar.HeadComment, // Comments above the key
				LineComment:      baseScalar.LineComment, // Inline comments
				FootComment:      baseScalar.FootComment, // Comments after
				TagValue:         baseScalar.TagValue,
				AnchorValue:      baseScalar.AnchorValue,
				BlankLinesBefore: baseScalar.BlankLinesBefore,
				BlankLinesAfter:  baseScalar.BlankLinesAfter,
			},
			Value: baseScalar.Value,
			Style: baseScalar.Style,
			Alias: baseScalar.Alias,
		}
	}

	return baseKey
}

// TransferInterFieldComments transfers comments stored in value HeadComments to next key HeadComments
func (p *NodeProcessor) TransferInterFieldComments(pairs []*node.MappingPair) {
	for i := 0; i < len(pairs)-1; i++ {
		// Check if current value has HeadComment (which is actually for next field)
		if scalar, ok := pairs[i].Value.(*node.ScalarNode); ok {
			if scalar.HeadComment != nil && len(scalar.HeadComment.Comments) > 0 {
				// Transfer to next key if it doesn't have comments
				if nextKey, ok := pairs[i+1].Key.(*node.ScalarNode); ok {
					if nextKey.HeadComment == nil {
						nextKey.HeadComment = scalar.HeadComment
						// Clear from value to avoid duplication
						scalar.HeadComment = nil
					}
				}
			}
		}
	}
}
