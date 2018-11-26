package deps

import (
	"fmt"

	"github.com/dselans/go-pidstat/stat"
)

type Dependencies struct {
	Statter stat.Statter
}

func New() (*Dependencies, error) {
	d := &Dependencies{}

	// Setup process statter
	p, err := stat.New()
	if err != nil {
		return nil, fmt.Errorf("unable to instan3tiate stat: %v", err)
	}

	d.Statter = p

	return d, nil
}
