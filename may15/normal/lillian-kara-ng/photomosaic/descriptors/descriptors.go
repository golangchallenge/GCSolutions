// The Descriptors package is used to create descriptors that describe an
// image quantitatively. These descriptors can be compared to each other to
// determine how similar they are to one another.

package descriptors

import (
	"image"
)

// Describes the similarity between two descriptions
// Computes the MatchScore based on Description.Normalized
// Lower MatchScores mean that two descriptions are more similar
type Description interface {
	MatchScore(other Description) (float64, error)
}

// Used to build a description for a region of an image
type DescriptionBuilder interface {
	GetDescription(m image.Image, b image.Rectangle) (Description, error)
}
