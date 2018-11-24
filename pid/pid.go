package pid

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dselans/go-pidstat/util"
	"github.com/relistan/go-director"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"
)

const (
	CacheProcessListInterval = 5 * time.Second
)

var (
	NotWatchedErr = errors.New("pid is not actively watched")

	sugar *zap.SugaredLogger
)

type Statter interface {
	GetProcesses() ([]Entry, error)
	StartWatchProcess(pid int32) error
	GetStats() (map[string][]Entry, error)
	GetStatsForPID(pid int32) ([]Entry, error)
	StopWatchProcess(pid int32) error
}

type Stat struct {
	processList       []Entry
	processListLooper *director.TimedLooper
	processListLock   *sync.Mutex
	stats             map[int32][]Entry
}

type Entry struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	PID       int32      `json:"pid"`
	Name      string     `json:"name"`
	CmdLine   string     `json:"cmd_line"`
}

func init() {
	logger, err := util.CreateLogger(false, map[string]interface{}{"pkg": "pid"})
	if err != nil {
		panic(fmt.Sprintf("unable to setup logger: %v", err))
	}

	sugar = logger.Sugar()
}

func New() (*Stat, error) {
	s := &Stat{
		processListLooper: director.NewImmediateTimedLooper(director.FOREVER, CacheProcessListInterval, nil),
		processListLock:   &sync.Mutex{},
		processList:       make([]Entry, 0),
	}

	// run processlist fetcher on an interval
	go s.cacheProcessList()

	return s, nil
}

func (s *Stat) cacheProcessList() error {
	s.processListLooper.Loop(func() error {
		processList, err := s.fetchProcessList()
		if err != nil {
			sugar.Errorf("unable to fetch processlist: %v", err)
			return nil
		}

		s.processListLock.Lock()
		s.processList = processList
		s.processListLock.Unlock()

		return nil
	})

	sugar.Debugf("runCacheProcessList exiting")
	return nil
}

func (s *Stat) fetchProcessList() ([]Entry, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0)
	now := time.Now()

	for _, p := range processes {
		entry := Entry{
			PID: p.Pid,
		}

		cmdLine, err := p.Cmdline()
		if err != nil {
			//sugar.Debugf("unable to determine cmdline for pid '%v': %v", p.Pid, err)
			continue
		}

		name, err := p.Name()
		if err != nil {
			//sugar.Debugf("unable to determine name for pid '%v': %v", p.Pid, err)
			continue
		}

		entry.Name = name
		entry.CmdLine = cmdLine
		entry.Timestamp = &now

		entries = append(entries, entry)
	}

	return entries, nil
}

// Get a list of all running processes (from cache)
func (s *Stat) GetProcesses() ([]Entry, error) {
	s.processListLock.Lock()

	processList := make([]Entry, len(s.processList))
	copy(processList, s.processList)

	s.processListLock.Unlock()

	return processList, nil
}

// Start gathering stats for a specific process
func (s *Stat) StartWatchProcess(pid int32) error {
	return nil
}

// Stop gathering stats for a specific process
func (s *Stat) StopWatchProcess(pid int32) error {
	return nil
}

// Get statistics for a specific pid
func (s *Stat) GetStatsForPID(pid int32) ([]Entry, error) {
	return nil, NotWatchedErr
}

// Get all stats
func (s *Stat) GetStats() (map[string][]Entry, error) {
	return nil, fmt.Errorf("shit broke")
}
