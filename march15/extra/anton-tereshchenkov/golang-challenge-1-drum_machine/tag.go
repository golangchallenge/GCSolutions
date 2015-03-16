package drum

import (
	"fmt"
	"strconv"
	"strings"
)

// spliceTag is a splice tag of a struct field.
type spliceTag struct {
	Size int
	opts string
}

// parseSpliceTag parses a raw tag string into a spliceTag object.
func parseSpliceTag(tag string) (*spliceTag, error) {
	s := &spliceTag{}
	// Separate tag from options
	if idx := strings.Index(tag, ","); idx != -1 {
		tag, s.opts = tag[:idx], tag[idx+1:]
	}
	if tag == "" {
		return s, nil
	}

	size, err := strconv.Atoi(tag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tag: %v", err)
	}
	s.Size = size
	return s, nil
}

// HasOptions checks if the option is specified in tag.
func (s *spliceTag) HasOption(option string) bool {
	for _, opt := range strings.Split(s.opts, ",") {
		if opt == option {
			return true
		}
	}
	return false
}
