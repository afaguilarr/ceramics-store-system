package main

import (
	"fmt"
	"strings"
)

// Define custom Categories and Images fields with []string type
type textArray []string

// Scan converts a database value (of type []uint8) to a text array ([]string).
// Implements the database/sql Scanner interface.
func (ta *textArray) Scan(value interface{}) error {
	// If the value is nil, assign nil to the text array pointer and return nil error.
	if value == nil {
		*ta = nil
		return nil
	}

	// Check that the value is a []uint8 (which is how PostgreSQL returns text arrays).
	u, ok := value.([]uint8)
	if !ok {
		return fmt.Errorf("failed to scan text array field: value is not []uint8")
	}

	// Convert the []uint8 to a string.
	str := string(u[:])

	// Remove the opening and closing braces from the string.
	str = strings.Trim(str, "{}")

	// Split the string into individual values (which are surrounded by quotes).
	values := strings.Split(str, ",")

	// Trim the quotes from the values and add them to a []string.
	var ss []string
	for _, v := range values {
		ss = append(ss, strings.Trim(v, `"'`))
	}

	// Assign the []string to the text array pointer and return nil error.
	*ta = ss
	return nil
}
