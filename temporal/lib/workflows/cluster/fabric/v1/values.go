package cluster

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// ToProtoStruct converts any Go struct to a protobuf Struct
// Uses JSON marshalling as an intermediate format for structure-agnostic conversion
func ToProtoStruct(v interface{}) (*structpb.Struct, error) {
	// Marshal to JSON
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal to json: %w", err)
	}

	// Unmarshal to map
	var m map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}

	// Convert map to protobuf Struct
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil, fmt.Errorf("create struct: %w", err)
	}

	return s, nil
}
