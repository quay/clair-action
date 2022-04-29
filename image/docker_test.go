package image

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func setup(resultsDir string) {
	os.Mkdir(resultsDir, 0700)
}

func teardown(resultsDir string) {
	os.RemoveAll(resultsDir)
}

func TestFromExported(t *testing.T) {
	resultsDir := "testdata/results"
	setup(resultsDir)
	defer teardown(resultsDir)
	ctx := context.TODO()
	di, err := NewDockerLocalImage(ctx, "testdata/algo", resultsDir)
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	fmt.Println(di)
}
