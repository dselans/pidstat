package pid

type Statter interface {

}

type Stat struct {

}

func New() (*Stat, error) {
	return &Stat{}, nil
}