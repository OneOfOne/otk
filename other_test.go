package otk

import (
	"reflect"
	"testing"
)

func TestMergeMap(t *testing.T) {
	tests := []struct {
		name     string
		dst      M
		src      M
		expected M
	}{
		{
			name: "Merge with non-empty maps",
			dst:  M{"a": 1, "b": M{"c": 2}},
			src:  M{"b": M{"d": 3}, "e": 4},
			expected: M{
				"a": 1,
				"b": M{
					"c": 2,
					"d": 3,
				},
				"e": 4,
			},
		},
		{
			name:     "Merge with empty source map",
			dst:      M{"a": 1},
			src:      M{},
			expected: M{"a": 1},
		},
		{
			name:     "Merge with empty destination map",
			dst:      M{},
			src:      M{"a": 1},
			expected: M{"a": 1},
		},
		{
			name:     "Merge with nil destination map",
			dst:      nil,
			src:      M{"a": 1},
			expected: M{"a": 1},
		},
		{
			name: "Merge with nested empty map",
			dst:  M{"a": M{"b": 2}},
			src:  M{"a": M{}},
			expected: M{
				"a": M{
					"b": 2,
				},
			},
		},
		{
			name: "Merge with nil values",
			dst:  M{"a": 1, "b": nil},
			src:  M{"b": 2, "c": nil},
			expected: M{
				"a": 1,
				"b": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeMap(tt.dst, tt.src)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
