package drum

import (
	"fmt"
	"path"
	"testing"
	"time"
)

func TestPlay(t *testing.T) {
	decoded, err := DecodeFile(path.Join("fixtures", "pattern_3.splice"))

	if err != nil {
		panic(err)
	}

	fmt.Println("Playing music...")
	decoded.Play(10 * time.Second)
}
