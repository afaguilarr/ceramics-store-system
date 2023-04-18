package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestScan(t *testing.T) {
	tests := []struct {
		value       interface{}
		expected    textArray
		expectedErr error
	}{
		// Test case where value is nil
		{
			value:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		// Test case where value is not a []uint8
		{
			value:       "not a []uint8",
			expected:    nil,
			expectedErr: fmt.Errorf("failed to scan text array field: value is not []uint8"),
		},
		// Test case where value is a valid text array
		{
			value:       []uint8("{foo,bar,baz}"),
			expected:    []string{"foo", "bar", "baz"},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		var ta textArray
		err := ta.Scan(tc.value)
		if (err == nil && tc.expectedErr != nil) ||
			(err != nil && tc.expectedErr == nil) ||
			(err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error()) {
			t.Errorf("unexpected error, expected %v but got %v", tc.expectedErr, err)
		}
		if !reflect.DeepEqual(ta, tc.expected) {
			t.Errorf("unexpected result, expected %v but got %v", tc.expected, ta)
		}
	}
}
