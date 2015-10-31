package util

// Boundary is a wrapper that contains the position of the vertices for a square or rectangle.
// Not detailed enough for other quadrilaterals.
type Boundary struct {
	LeftX   float32
	RightX  float32
	BottomY float32
	TopY    float32
}
