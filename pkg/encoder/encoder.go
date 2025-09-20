package encoder

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/node"
	"github.com/elioetibr/golang-yaml/pkg/parser"
	"github.com/elioetibr/golang-yaml/pkg/serializer"
)

// Marshal returns the YAML encoding of v
func Marshal(v interface{}) ([]byte, error) {
	n, err := valueToNode(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}

	opts := serializer.DefaultOptions()
	result, err := serializer.SerializeToString(n, opts)
	if err != nil {
		return nil, err
	}

	return []byte(result), nil
}

// MarshalWithOptions returns the YAML encoding of v with custom options
func MarshalWithOptions(v interface{}, opts *serializer.Options) ([]byte, error) {
	n, err := valueToNode(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}

	result, err := serializer.SerializeToString(n, opts)
	if err != nil {
		return nil, err
	}

	return []byte(result), nil
}

// Encoder writes YAML values-with-comments to an output stream
type Encoder struct {
	writer  io.Writer
	options *serializer.Options
	builder node.Builder
}

// NewEncoder returns a new encoder that writes to w
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer:  w,
		options: serializer.DefaultOptions(),
		builder: &node.DefaultBuilder{},
	}
}

// SetOptions sets the serialization options
func (e *Encoder) SetOptions(opts *serializer.Options) {
	e.options = opts
}

// Encode writes the YAML encoding of v to the stream
func (e *Encoder) Encode(v interface{}) error {
	n, err := valueToNode(reflect.ValueOf(v))
	if err != nil {
		return err
	}

	s := serializer.NewSerializer(e.writer, e.options)
	return s.Serialize(n)
}

// valueToNode converts a Go value to a YAML node
func valueToNode(v reflect.Value) (node.Node, error) {
	builder := &node.DefaultBuilder{}

	// Handle nil and zero values-with-comments
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return builder.BuildScalar("null", node.StylePlain), nil
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return builder.BuildScalar(v.String(), node.StylePlain), nil

	case reflect.Bool:
		// Store as plain scalars, serializer won't quote these
		if v.Bool() {
			return builder.BuildScalar("true", node.StylePlain), nil
		}
		return builder.BuildScalar("false", node.StylePlain), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return builder.BuildScalar(fmt.Sprintf("%d", v.Int()), node.StylePlain), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return builder.BuildScalar(fmt.Sprintf("%d", v.Uint()), node.StylePlain), nil

	case reflect.Float32, reflect.Float64:
		return builder.BuildScalar(fmt.Sprintf("%v", v.Float()), node.StylePlain), nil

	case reflect.Slice, reflect.Array:
		return sliceToNode(v)

	case reflect.Map:
		return mapToNode(v)

	case reflect.Struct:
		return structToNode(v)

	case reflect.Interface:
		if v.IsNil() {
			return builder.BuildScalar("null", node.StylePlain), nil
		}
		return valueToNode(v.Elem())

	default:
		return nil, fmt.Errorf("unsupported type: %v", v.Type())
	}
}

// sliceToNode converts a slice or array to a sequence node
func sliceToNode(v reflect.Value) (node.Node, error) {
	builder := &node.DefaultBuilder{}
	items := make([]node.Node, v.Len())

	for i := 0; i < v.Len(); i++ {
		item, err := valueToNode(v.Index(i))
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	return builder.BuildSequence(items, node.StyleBlock), nil
}

// mapToNode converts a map to a mapping node
func mapToNode(v reflect.Value) (node.Node, error) {
	builder := &node.DefaultBuilder{}
	pairs := make([]*node.MappingPair, 0, v.Len())

	for _, key := range v.MapKeys() {
		keyNode, err := valueToNode(key)
		if err != nil {
			return nil, err
		}

		valueNode, err := valueToNode(v.MapIndex(key))
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, &node.MappingPair{
			Key:   keyNode,
			Value: valueNode,
		})
	}

	return builder.BuildMapping(pairs, node.StyleBlock), nil
}

// structToNode converts a struct to a mapping node
func structToNode(v reflect.Value) (node.Node, error) {
	builder := &node.DefaultBuilder{}
	t := v.Type()
	pairs := make([]*node.MappingPair, 0)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from yaml tag if present
		name := field.Name
		if tag := field.Tag.Get("yaml"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
			// Skip if tag is "-"
			if name == "-" {
				continue
			}

			// Handle omitempty
			if len(parts) > 1 && parts[1] == "omitempty" {
				if isEmptyValue(v.Field(i)) {
					continue
				}
			}
		}

		keyNode := builder.BuildScalar(name, node.StylePlain)
		valueNode, err := valueToNode(v.Field(i))
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, &node.MappingPair{
			Key:   keyNode,
			Value: valueNode,
		})
	}

	return builder.BuildMapping(pairs, node.StyleBlock), nil
}

// isEmptyValue checks if a value is empty for omitempty purposes
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// MarshalNode converts a YAML node to its byte representation
func MarshalNode(n node.Node) ([]byte, error) {
	opts := serializer.DefaultOptions()
	result, err := serializer.SerializeToString(n, opts)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

// UnmarshalNode parses YAML input and returns the AST node
func UnmarshalNode(data []byte) (node.Node, error) {
	return parser.ParseString(string(data))
}
