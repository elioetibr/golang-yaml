package main

import (
	"fmt"
	"log"

	"github.com/elioetibr/golang-yaml/pkg/decoder"
	"github.com/elioetibr/golang-yaml/pkg/encoder"
	"github.com/elioetibr/golang-yaml/pkg/node"
	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
	"github.com/elioetibr/golang-yaml/pkg/transform"
)

// Config represents an application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Features []string       `yaml:"features"`
	Debug    bool           `yaml:"debug"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password,omitempty"`
}

func main() {
	fmt.Println("=== YAML Library Full Demo ===")

	// 1. Marshal Go struct to YAML
	fmt.Println("1. Marshal Go struct to YAML:")
	config := Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			Username: "admin",
		},
		Features: []string{"auth", "api", "metrics"},
		Debug:    true,
	}

	yamlData, err := encoder.Marshal(config)
	if err != nil {
		log.Fatalf("Marshal error: %v", err)
	}

	fmt.Printf("%s\n", yamlData)

	// 2. Parse YAML to AST
	fmt.Println("2. Parse YAML string to AST:")
	yamlInput := `# Application config
name: MyApp
version: 1.0.0

# Server settings
server:
  host: 0.0.0.0
  port: 3000

# Feature flags
features:
  - logging
  - caching
  - monitoring

settings:
  timeout: 30
  retries: 3
  debug: false
`

	root, err := parser.ParseString(yamlInput)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}
	fmt.Println("✓ Successfully parsed YAML to AST")

	// 3. Sort the AST
	fmt.Println("\n3. Sort the AST (ascending):")
	sortConfig := &transform.SortConfig{
		Mode:   transform.SortModeAscending,
		SortBy: transform.SortByKey,
	}

	sorter := transform.NewSorter(sortConfig)
	sorted := sorter.Sort(root)

	// Serialize sorted AST back to YAML
	sortedYAML, err := serializer.SerializeToString(sorted, nil)
	if err != nil {
		log.Fatalf("Serialize error: %v", err)
	}
	fmt.Printf("%s\n", sortedYAML)

	// 4. Apply formatting
	fmt.Println("4. Apply custom formatting:")
	formatConfig := &transform.FormatConfig{
		DefaultBlankLinesBeforeComment: 1,
		BlankLinesBeforeHeadComment:    1,
		PreserveOriginal:               false,
	}

	formatter := transform.NewFormatter(formatConfig)
	formatted := formatter.Format(root)

	_, err = serializer.SerializeToString(formatted, nil)
	if err != nil {
		log.Fatalf("Serialize error: %v", err)
	}
	fmt.Println("✓ Applied formatting with blank lines")

	// 5. Unmarshal YAML to Go struct
	fmt.Println("\n5. Unmarshal YAML to Go struct:")
	var loadedConfig Config
	configYAML := `
server:
  host: example.com
  port: 443
database:
  driver: mysql
  host: db.example.com
  port: 3306
  username: root
features:
  - ssl
  - cache
  - api
debug: false
`

	err = decoder.Unmarshal([]byte(configYAML), &loadedConfig)
	if err != nil {
		log.Fatalf("Unmarshal error: %v", err)
	}

	fmt.Printf("Server: %s:%d\n", loadedConfig.Server.Host, loadedConfig.Server.Port)
	fmt.Printf("Database: %s://%s@%s:%d\n",
		loadedConfig.Database.Driver,
		loadedConfig.Database.Username,
		loadedConfig.Database.Host,
		loadedConfig.Database.Port)
	fmt.Printf("Features: %v\n", loadedConfig.Features)
	fmt.Printf("Debug: %v\n", loadedConfig.Debug)

	// 6. Build AST programmatically
	fmt.Println("\n6. Build AST programmatically:")
	builder := &node.DefaultBuilder{}

	programmaticAST := builder.BuildMapping([]*node.MappingPair{
		{
			Key:   builder.BuildScalar("name", node.StylePlain),
			Value: builder.BuildScalar("Generated App", node.StylePlain),
		},
		{
			Key: builder.BuildScalar("config", node.StylePlain),
			Value: builder.BuildMapping([]*node.MappingPair{
				{
					Key:   builder.BuildScalar("enabled", node.StylePlain),
					Value: builder.BuildScalar("true", node.StylePlain),
				},
				{
					Key:   builder.BuildScalar("level", node.StylePlain),
					Value: builder.BuildScalar("info", node.StylePlain),
				},
			}, node.StyleBlock),
		},
		{
			Key: builder.BuildScalar("modules", node.StylePlain),
			Value: builder.BuildSequence([]node.Node{
				builder.BuildScalar("core", node.StylePlain),
				builder.BuildScalar("auth", node.StylePlain),
				builder.BuildScalar("api", node.StylePlain),
			}, node.StyleBlock),
		},
	}, node.StyleBlock)

	// Add comments
	node.AssociateComment(programmaticAST, "# Generated configuration", node.CommentPositionAbove, 0)

	// Serialize with custom options
	serializeOpts := &serializer.Options{
		Indent:                2,
		PreserveComments:      true,
		ExplicitDocumentStart: true,
		ExplicitDocumentEnd:   true,
	}

	generatedYAML, err := serializer.SerializeToString(programmaticAST, serializeOpts)
	if err != nil {
		log.Fatalf("Serialize error: %v", err)
	}

	fmt.Println("Generated YAML with comments and document markers:")
	fmt.Printf("%s\n", generatedYAML)

	// 7. Round-trip test
	fmt.Println("7. Round-trip test (parse -> modify -> serialize):")
	roundTripYAML := `version: 1.0.0
name: TestApp
active: true`

	// Parse
	rtRoot, err := parser.ParseString(roundTripYAML)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Modify (sort)
	rtSorted := sorter.Sort(rtRoot)

	// Serialize back
	rtResult, err := serializer.SerializeToString(rtSorted, nil)
	if err != nil {
		log.Fatalf("Serialize error: %v", err)
	}

	fmt.Println("Original:")
	fmt.Println(roundTripYAML)
	fmt.Println("\nAfter round-trip with sorting:")
	fmt.Print(rtResult)

	fmt.Println("\n=== Demo Complete ===")
}
