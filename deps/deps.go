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

	// Setup dep 1
	p, err := stat.New()
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate stat: %v", err)
	}

	d.Statter = p

	// Setup dep 2

	// .. and so on

	return d, nil
}
