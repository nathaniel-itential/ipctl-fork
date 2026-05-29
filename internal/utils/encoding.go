// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package utils

import (
	"encoding/json"
	"fmt"

	"github.com/itential/ipctl/internal/logging"
	"gopkg.in/yaml.v2"
)

// ToMap accepts any object and will return at as a map using json marshal and
// unmarshal.  This fuction will return an error if if fails to marshal or
// unmarshal the input object.
func ToMap(in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &out)
}

// UnmarshalData will attempt to unmarshal a byte array into an object.  It
// will first attempt to unmarshal the byte array as JSON.  If that fails, it
// will fall back to unmarshalling the data as YAML.  An error is returned only
// when the data cannot be decoded as either format.
func UnmarshalData(data []byte, ptr any) error {
	if err := json.Unmarshal(data, ptr); err != nil {
		logging.Debug("failed to unmarshal data as json, falling back to yaml: %s", err)
		if yamlErr := yaml.Unmarshal(data, ptr); yamlErr != nil {
			return fmt.Errorf("failed to unmarshal data as json or yaml: %w", yamlErr)
		}
	}
	return nil
}
