package deps

import (
	"fmt"

	"github.com/gobuffalo/packr/v2"

	"github.com/dselans/go-pidstat/stat"
)

type Dependencies struct {
	Statter  stat.Statter
	PackrBox *packr.Box
}

func New() (*Dependencies, error) {
	d := &Dependencies{}

	// Setup process statter
	p, err := stat.New()
	if err != nil {
		return nil, fmt.Errorf("unable to instan3tiate stat: %v", err)
	}

	d.Statter = p

	// Setup assets
	d.PackrBox = packr.New("assets", "../assets")

	return d, nil
}
