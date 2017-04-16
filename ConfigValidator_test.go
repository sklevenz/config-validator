package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	yaml "gopkg.in/yaml.v2"
)

var simpleDocument = `
first:
  second:
    string: "ss"
    int: 2323
    float: 3.5
    array: ["a", "b"]
    map:
      string: "abc"
      int: 123

single: "abc"
`

var validationDocument = `
    mandatory_valid: abc
    optional_present: abc
    deprecated_present: abc
    obsolete_present: abc
`
var walkerDocument = `
    a1:
      b1:
        c1: 1
        c2: 2
        c3:
      b2:
      b3: ["1", "2"]
    a2:
`

/*
a1.b1.c1 1
a1.b1.c2 2
a1.b1.c3 nil
a1.b2 nil
a1.b3 ["1", ""]
a2 nil
*/

var validationSchemaDocument = `
schema:
  properties:
  - property: mandatory_valid
    annotations:
    - required
    description: "description"
    default: 123
  - property: mandatory_missing
    annotations:
    - required
    description: "description"
    default: 123
  - property: optional_present
    annotations:
    - optional
    descriptions: "description"
    default: 123
  - property: optional_missing
    annotations:
    - optional
    description: "description"
    default: 123
  - property: deprecated_present
    annotations:
    - deprecated
    description: "description"
    default: 123
`

var simpleMap anonymousMap
var walkerMap anonymousMap
var validationSchema *SchemaType

var validationConfig []anonymousMap

func init() {
	err := yaml.Unmarshal([]byte(simpleDocument), &simpleMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var tmpMap anonymousMap
	err = yaml.Unmarshal([]byte(validationDocument), &tmpMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	validationConfig = []anonymousMap{tmpMap}

	err = yaml.Unmarshal([]byte(walkerDocument), &walkerMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	schemaBytes := []byte(validationSchemaDocument)
	validationSchema = unmarshalSchema(&schemaBytes)
}

func TestWalker(t *testing.T) {
	resultMap := walkerMap.walk("")

	value, ok := resultMap["a2"]
	assert.True(t, ok, "value should be found")
	assert.Nil(t, value, "value should be nil")

	value, ok = resultMap["a1.b2"]
	assert.True(t, ok, "value should be found")
	assert.Nil(t, value, "value should be nil")

	value, ok = resultMap["a1.b3"]
	assert.True(t, ok, "value should be found")
	assert.IsType(t, []interface{}{}, value, "value should be nil")
	assert.Equal(t, []interface{}{"1", "2"}, value, "value should be nil")

	value, ok = resultMap["a1.b1.c1"]
	assert.True(t, ok, "value should be found")
	assert.Equal(t, 1, value, "value should be nil")

	value, ok = resultMap["a1.b1.c2"]
	assert.True(t, ok, "value should be found")
	assert.Equal(t, 2, value, "value should be nil")

	value, ok = resultMap["a1.b1.c3"]
	assert.True(t, ok, "value should be found")
	assert.Nil(t, value, "value should be nil")

	_, ok = resultMap["a1"]
	assert.False(t, ok, "value should not be found")

	_, ok = resultMap["a1.b1"]
	assert.False(t, ok, "value should not be found")
}

func TestValidation(t *testing.T) {
	validationResult := validate(validationSchema, &validationConfig)

	assert.Equal(t, 6, len(validationResult), "result size doesn't fit")

	assert.Equal(t, "mandatory_valid", validationResult[0].PropertyName, "property name should be equal")
	assert.Equal(t, valid, validationResult[0].Status, "status should be valid")

	assert.Equal(t, "mandatory_missing", validationResult[1].PropertyName, "property name should be equal")
	assert.Equal(t, flaw, validationResult[1].Status, "status should be flaw")

	assert.Equal(t, "optional_present", validationResult[2].PropertyName, "property name should be equal")
	assert.Equal(t, valid, validationResult[2].Status, "status should be valid")

	assert.Equal(t, "optional_missing", validationResult[3].PropertyName, "property name should be equal")
	assert.Equal(t, valid, validationResult[3].Status, "status should be valid")

	assert.Equal(t, "deprecated_present", validationResult[4].PropertyName, "property name should be equal")
	assert.Equal(t, valid, validationResult[4].Status, "status should be deprecated")

	assert.Equal(t, "obsolete_present", validationResult[5].PropertyName, "property name should be equal")
	assert.Equal(t, obsolete, validationResult[5].Provided, "provided should be obsolete")
}

func TestContainsKeyFit(t *testing.T) {
	var val interface{}
	var ok bool

	val, ok = simpleMap.containsKey("single")
	assert.Equal(t, ok, true, "key should be found")
	assert.Equal(t, val, "abc", "value should fit")

	val, ok = simpleMap.containsKey("first.second.string")
	assert.Equal(t, ok, true, "key should be found")
	assert.Equal(t, val, "ss", "value should fit")

	val, ok = simpleMap.containsKey("first.second.int")
	assert.Equal(t, ok, true, "key should be found")
	assert.Equal(t, val, 2323, "value should fit")

	val, ok = simpleMap.containsKey("first.second.float")
	assert.Equal(t, ok, true, "key should be found")
	assert.Equal(t, val, 3.5, "value should fit")

	val, ok = simpleMap.containsKey("first.second.array")
	assert.Equal(t, ok, true, "key should be found")
	assert.Equal(t, val, []interface{}{"a", "b"}, "value should fit")

	val, ok = simpleMap.containsKey("first.second.map")
	assert.Equal(t, ok, true, "key should be found")
	assert.Equal(t, val, anonymousMap{"string": "abc", "int": 123}, "value should fit")
}

func TestContainsKeyMissFit(t *testing.T) {
	var val interface{}
	var ok bool

	val, ok = simpleMap.containsKey("not_found")
	assert.Equal(t, ok, false, "key should not be found")
	assert.Equal(t, val, nil, "value should be nil")

	val, ok = simpleMap.containsKey("not_found.2ndkey")
	assert.Equal(t, ok, false, "key should not be found")
	assert.Equal(t, val, nil, "value should be nil")
}
