package decoder

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/elioetibr/golang-yaml/pkg/node"
	"github.com/elioetibr/golang-yaml/pkg/parser"
)

// Unmarshal parses the YAML-encoded data and stores the result
// in the value pointed to by v
func Unmarshal(data []byte, v interface{}) error {
	n, err := parser.ParseString(string(data))
	if err != nil {
		return err
	}

	return nodeToValue(n, reflect.ValueOf(v))
}

// Decoder reads and decodes YAML values-with-comments from an input stream
type Decoder struct {
	reader io.Reader
	buffer []byte
}

// NewDecoder returns a new decoder that reads from r
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// Decode reads the next YAML-encoded value from its input
// and stores it in the value pointed to by v
func (d *Decoder) Decode(v interface{}) error {
	if d.buffer == nil {
		data, err := ioutil.ReadAll(d.reader)
		if err != nil {
			return err
		}
		d.buffer = data
	}

	return Unmarshal(d.buffer, v)
}

// nodeToValue converts a YAML node to a Go value
func nodeToValue(n node.Node, v reflect.Value) error {
	if !v.IsValid() {
		return fmt.Errorf("invalid value")
	}

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return nodeToValue(n, v.Elem())
	}

	// Handle interfaces
	if v.Kind() == reflect.Interface && !v.IsNil() {
		return nodeToValue(n, v.Elem())
	}

	switch node := n.(type) {
	case *node.ScalarNode:
		return scalarToValue(node, v)
	case *node.SequenceNode:
		return sequenceToValue(node, v)
	case *node.MappingNode:
		return mappingToValue(node, v)
	default:
		return fmt.Errorf("unknown node type: %T", n)
	}
}

// scalarToValue converts a scalar node to a Go value
func scalarToValue(n *node.ScalarNode, v reflect.Value) error {
	value := n.Value

	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Bool:
		b, err := parseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)

	case reflect.Interface:
		// Try to parse the value as appropriate type
		parsed := parser.ParseValue(value)
		v.Set(reflect.ValueOf(parsed))

	default:
		return fmt.Errorf("cannot unmarshal scalar into %v", v.Type())
	}

	return nil
}

// sequenceToValue converts a sequence node to a Go value
func sequenceToValue(n *node.SequenceNode, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Slice:
		// Create a new slice with appropriate capacity
		slice := reflect.MakeSlice(v.Type(), len(n.Items), len(n.Items))
		for i, item := range n.Items {
			if err := nodeToValue(item, slice.Index(i)); err != nil {
				return err
			}
		}
		v.Set(slice)

	case reflect.Array:
		// Fill array elements
		for i := 0; i < len(n.Items) && i < v.Len(); i++ {
			if err := nodeToValue(n.Items[i], v.Index(i)); err != nil {
				return err
			}
		}

	case reflect.Interface:
		// Create a slice of interface{}
		slice := make([]interface{}, len(n.Items))
		for i, item := range n.Items {
			var val interface{}
			if err := nodeToValue(item, reflect.ValueOf(&val).Elem()); err != nil {
				return err
			}
			slice[i] = val
		}
		v.Set(reflect.ValueOf(slice))

	default:
		return fmt.Errorf("cannot unmarshal sequence into %v", v.Type())
	}

	return nil
}

// mappingToValue converts a mapping node to a Go value
func mappingToValue(n *node.MappingNode, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Map:
		return mappingToMap(n, v)
	case reflect.Struct:
		return mappingToStruct(n, v)
	case reflect.Interface:
		// Create a map[string]interface{}
		m := make(map[string]interface{})
		mapVal := reflect.ValueOf(m)
		if err := mappingToMap(n, mapVal); err != nil {
			return err
		}
		v.Set(mapVal)
	default:
		return fmt.Errorf("cannot unmarshal mapping into %v", v.Type())
	}

	return nil
}

// mappingToMap converts a mapping node to a Go map
func mappingToMap(n *node.MappingNode, v reflect.Value) error {
	// Create map if nil
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	for _, pair := range n.Pairs {
		// Get key as string (most common case)
		keyStr := ""
		if scalar, ok := pair.Key.(*node.ScalarNode); ok {
			keyStr = scalar.Value
		} else {
			return fmt.Errorf("non-scalar map keys not supported")
		}

		// Create values-with-comments for key and value
		keyVal := reflect.New(v.Type().Key()).Elem()
		valVal := reflect.New(v.Type().Elem()).Elem()

		// Set the key
		if v.Type().Key().Kind() == reflect.String {
			keyVal.SetString(keyStr)
		} else {
			if err := scalarToValue(&node.ScalarNode{Value: keyStr}, keyVal); err != nil {
				return err
			}
		}

		// Set the value
		if err := nodeToValue(pair.Value, valVal); err != nil {
			return err
		}

		v.SetMapIndex(keyVal, valVal)
	}

	return nil
}

// mappingToStruct converts a mapping node to a Go struct
func mappingToStruct(n *node.MappingNode, v reflect.Value) error {
	t := v.Type()

	// Build field map
	fieldMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Name
		if tag := field.Tag.Get("yaml"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
			if name == "-" {
				continue
			}
		}

		// Store both lowercase and original
		fieldMap[strings.ToLower(name)] = i
		fieldMap[name] = i
	}

	// Set fields from mapping
	for _, pair := range n.Pairs {
		// Get key as string
		keyStr := ""
		if scalar, ok := pair.Key.(*node.ScalarNode); ok {
			keyStr = scalar.Value
		} else {
			continue // Skip non-scalar keys
		}

		// Find field index
		fieldIndex, ok := fieldMap[keyStr]
		if !ok {
			// Try lowercase match
			fieldIndex, ok = fieldMap[strings.ToLower(keyStr)]
			if !ok {
				continue // Skip unknown fields
			}
		}

		// Set field value
		fieldVal := v.Field(fieldIndex)
		if fieldVal.CanSet() {
			if err := nodeToValue(pair.Value, fieldVal); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseBool parses a YAML boolean value
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "yes", "on", "y":
		return true, nil
	case "false", "no", "off", "n":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse %q as bool", s)
	}
}

// UnmarshalStrict is like Unmarshal but returns an error
// when the destination has fields that are not found in the source
func UnmarshalStrict(data []byte, v interface{}) error {
	// TODO: Implement strict unmarshaling
	return Unmarshal(data, v)
}
