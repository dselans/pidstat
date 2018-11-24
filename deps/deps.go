package deps

import (
	"fmt"

	"github.com/dselans/go-pidstat/pid"
)

type Dependencies struct {
	Statter pid.Statter
}

func New() (*Dependencies, error) {
	d := &Dependencies{}

	// Setup dep 1
	p, err := pid.New()
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate pid: %v", err)
	}

	d.Statter = p

	// Setup dep 2

	// .. and so on

	return d, nil
}