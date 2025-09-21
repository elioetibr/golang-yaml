package main

import (
	"fmt"
	"log"

	"github.com/elioetibr/golang-yaml/pkg/merge"
)

func main() {
	// Test case that reproduces the helm-charts-migrator issue
	base := `
# yaml-language-server: $schema=values.schema.json
# Default values for livecomments.

# @schema
# enum: ["218894879100", "825534976873"]
# required: true
# additionalProperties: false
# @schema
# -- AWS Account ID
awsAccountId: ""

# @schema
# type: string
# additionalProperties: false
# @schema
# -- Override for chart name
nameOverride: ""

# @schema
# type: string
# additionalProperties: false
# @schema
# -- Complete override for resource naming
fullnameOverride: ""

# @schema
# type: integer
# exclusiveMinimum: 0
# maximum: 500
# @schema
# -- Number of replicas
replicaCount: 1
`

	override := `
awsAccountId: "218894879100"
nameOverride: ""
fullnameOverride: livecomments
replicaCount: 1
`

	// Use default options with comment preservation
	opts := merge.DefaultOptions()
	opts.PreserveComments = true
	opts.PreserveBlankLines = true

	result, err := merge.MergeStringsWithOptions(base, override, opts)
	if err != nil {
		log.Fatalf("Merge failed: %v", err)
	}

	fmt.Println("=== MERGED RESULT ===")
	fmt.Println(result)
}