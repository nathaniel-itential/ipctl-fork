// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SampleStruct is a sample data structure for testing.
type SampleStruct struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

// TestToMapSuccess tests the successful conversion of a struct to a map.
func TestToMapSuccess(t *testing.T) {
	input := SampleStruct{
		Name:  "test",
		Value: 42,
	}

	var output map[string]interface{}
	err := ToMap(input, &output)

	assert.NoError(t, err)
	assert.Equal(t, "test", output["name"])
	assert.Equal(t, float64(42), output["value"]) // JSON unmarshals numbers into float64
}

// TestToMapInvalidInput tests ToMap with an invalid input (channel can't be marshaled).
func TestToMapInvalidInput(t *testing.T) {
	ch := make(chan int)
	var output map[string]interface{}
	err := ToMap(ch, &output)

	assert.Error(t, err)
}

// TestUnmarshalDataJSON tests successful JSON unmarshalling.
func TestUnmarshalDataJSON(t *testing.T) {
	data := []byte(`{"name": "json", "value": 123}`)
	var obj SampleStruct

	err := UnmarshalData(data, &obj)

	assert.NoError(t, err)
	assert.Equal(t, "json", obj.Name)
	assert.Equal(t, 123, obj.Value)
}

// TestUnmarshalDataYAML tests fallback to YAML unmarshalling.
func TestUnmarshalDataYAML(t *testing.T) {
	data := []byte("name: yaml\nvalue: 456")
	var obj SampleStruct

	err := UnmarshalData(data, &obj)

	assert.NoError(t, err)
	assert.Equal(t, "yaml", obj.Name)
	assert.Equal(t, 456, obj.Value)
}

// TestUnmarshalDataInvalid verifies that malformed data which is neither valid
// JSON nor valid YAML returns an error instead of terminating the process.
func TestUnmarshalDataInvalid(t *testing.T) {
	data := []byte(`"unterminated`)
	var obj SampleStruct

	err := UnmarshalData(data, &obj)

	assert.Error(t, err)
}
