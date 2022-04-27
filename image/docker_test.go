package image

import (
	"context"
	"fmt"
	"testing"
)

func TestFromExported(t *testing.T) {
	ctx := context.TODO()
	di, err := NewDockerLocalImage(ctx, "testdata/algo", "testdata/results")
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	fmt.Println(di)
}
