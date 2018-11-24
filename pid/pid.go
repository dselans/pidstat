package pid

import (
	"fmt"
	"time"

	"github.com/dselans/go-pidstat/util"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"
)

var (
	sugar *zap.SugaredLogger
)

type Statter interface {
	GetProcesses() ([]*process.Process, error)
	WatchProcess(pid int32) error
	GetStats() (map[string][]Entry, error)
	GetStatsForPID(pid int32) ([]Entry, error)
}

type Stat struct {
	stats map[int32][]Entry
}

type Entry struct {
	Timestamp time.Time
	Process   *process.Process
}

func init() {
	logger, err := util.CreateLogger(false, map[string]interface{}{"pkg": "pid"})
	if err != nil {
		panic(fmt.Sprintf("unable to setup logger: %v", err))
	}

	sugar = logger.Sugar()
}

func New() (*Stat, error) {
	return &Stat{}, nil
}

// Get a list of all running processes
func (s *Stat) GetProcesses() ([]*process.Process, error) {
	return process.Processes()
}

// Start gathering stats for a specific process
func (s *Stat) WatchProcess(pid int32) error {
	return nil
}

// Get statistics for a specific pid
func (s *Stat) GetStatsForPID(pid int32) ([]Entry, error) {
	return nil, nil
}

// Get all stats
func (s *Stat) GetStats() (map[string][]Entry, error) {
	return nil, nil
}
