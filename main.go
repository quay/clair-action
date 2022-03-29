package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/quay/claircore/libindex"
	"github.com/quay/claircore/libvuln/updates"
)

func main() {
	ctx := context.Background()

	imgName := os.Getenv("IMAGE_TAG")
	if imgName == "" {
		fmt.Printf("no $IMAGE_TAG set")
	}

	img, err := NewPodmanImage(ctx, imgName)
	if err != nil {
		fmt.Printf("error getting image information %v\n", err)
		return
	}

	mf, err := img.GetManifest()
	if err != nil {
		fmt.Printf("error creating manifest %v\n", err)
		return
	}

	indexerOpts := &libindex.Options{
		Store:      NewLocalIndexerStore(),
		Locker:     updates.NewLocalLockSource(),
		FetchArena: &LocalFetchArena{},
	}
	li, err := libindex.New(ctx, indexerOpts, http.DefaultClient)
	if err != nil {
		fmt.Printf("error creating Libindex %v\n", err)
		return
	}
	ir, err := li.Index(ctx, &mf)
	if err != nil {
		fmt.Printf("error creating index report %v\n", err)
		return
	}

	blob, err := json.MarshalIndent(ir, "", "  ")
	if err != nil {
		fmt.Printf("error marshaling index report %v\n", err)
		return
	}
	fmt.Println(string(blob))
}
