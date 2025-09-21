package main

import (
    "fmt"
    "io/ioutil"
    "strings"
    "github.com/elioetibr/golang-yaml/v1/pkg/parser"
    "github.com/elioetibr/golang-yaml/v1/pkg/lexer"
    "github.com/elioetibr/golang-yaml/v1/pkg/node"
    "github.com/elioetibr/golang-yaml/v1/pkg/merge"
    "github.com/elioetibr/golang-yaml/v1/pkg/serializer"
)

func main() {
    // Read and merge
    base, _ := ioutil.ReadFile("v1/examples/values/base-helm-chart-values.yaml")
    baseNode, _ := parser.NewParser(lexer.NewLexerFromString(string(base))).Parse()
    
    override, _ := ioutil.ReadFile("v1/examples/values/legacy-helm-chart-values.yaml")
    overrideNode, _ := parser.NewParser(lexer.NewLexerFromString(string(override))).Parse()
    
    opts := &merge.Options{
        PreserveComments: true,
        Strategy: merge.StrategyDeep,
    }
    merged, _ := merge.NewMerger(opts).Merge(baseNode, overrideNode)
    
    // Serialize
    var result strings.Builder
    s := serializer.NewSerializer(&result, &serializer.Options{
        PreserveComments: true,
    })
    s.Serialize(merged)
    
    // Check for key comments in output
    lines := strings.Split(result.String(), "\n")
    for i, line := range lines {
        if strings.HasPrefix(line, "awsAccountId:") || strings.HasPrefix(line, "configMap:") {
            fmt.Printf("Line %d: %s\n", i+1, line)
            if i > 0 {
                fmt.Printf("  Previous line: %s\n", lines[i-1])
            }
        }
    }
}
