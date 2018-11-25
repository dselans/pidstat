package stat

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
	StatInterval             = 5 * time.Second
)

var (
	NotWatchedErr     = errors.New("pid is not actively watched")
	AlreadyWatchedErr = errors.New("pid is already being watched")

	sugar *zap.SugaredLogger
)

type Statter interface {
	GetProcesses() ([]ProcInfo, error)
	GetStats() (map[string][]ProcInfo, error)
	GetStatsForPID(pid int32) ([]ProcInfo, error)
	StartWatchProcess(pid int32) error
	StopWatchProcess(pid int32) error
}

type Stat struct {
	// ProcInfo list
	processList []ProcInfo

	// Looper used for fetching + caching process list
	processListLooper *director.TimedLooper

	// Lock used for accessing process list
	processListLock *sync.Mutex

	// Map containing all actively watched processes (and their info)
	watched map[int32]Proc

	// TODO: Should we keep history?

	// Lock used for accessing watched map
	watchedLock *sync.Mutex
}

type Proc struct {
	ProcInfo ProcInfo
	Looper   *director.TimedLooper
	Process  *process.Process
}

type ProcInfo struct {
	// Available in both Stat.processList AND Proc.Metrics
	PID     int32  `json:"pid"`
	Name    string `json:"name"`
	CmdLine string `json:"cmd_line"`

	// Available only in Proc.Metrics
	Metrics     []ProcInfoMetrics `json:"data,omitempty"`
	MetricsLock *sync.Mutex       `json:"-"`
}

type ProcInfoMetrics struct {
	VMS       uint64    `json:"vms"`
	RSS       uint64    `json:"rss"`
	Swap      uint64    `json:"swap"`
	CPU       float64   `json:"cpu"`
	Threads   int32     `json:"threads"`
	Timestamp time.Time `json:"timestamp"`
}

func init() {
	logger, err := util.CreateLogger(false, map[string]interface{}{"pkg": "stat"})
	if err != nil {
		panic(fmt.Sprintf("unable to setup logger: %v", err))
	}

	sugar = logger.Sugar()
}

func New() (*Stat, error) {
	s := &Stat{
		processListLooper: director.NewImmediateTimedLooper(director.FOREVER, CacheProcessListInterval, nil),
		processListLock:   &sync.Mutex{},
		processList:       make([]ProcInfo, 0),
		watchedLock:       &sync.Mutex{},
		watched:           make(map[int32]Proc, 0),
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

func (s *Stat) fetchProcessList() ([]ProcInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	entries := make([]ProcInfo, 0)

	for _, p := range processes {
		entry := ProcInfo{
			PID: p.Pid,
		}

		cmdLine, err := p.Cmdline()
		if err != nil {
			//sugar.Debugf("unable to determine cmdline for stat '%v': %v", p.Pid, err)
			continue
		}

		name, err := p.Name()
		if err != nil {
			//sugar.Debugf("unable to determine name for stat '%v': %v", p.Pid, err)
			continue
		}

		entry.Name = name
		entry.CmdLine = cmdLine

		entries = append(entries, entry)
	}

	return entries, nil
}

// Get a list of all running processes (from cache)
func (s *Stat) GetProcesses() ([]ProcInfo, error) {
	s.processListLock.Lock()

	processList := make([]ProcInfo, len(s.processList))
	copy(processList, s.processList)

	s.processListLock.Unlock()

	return processList, nil
}

func (s *Stat) getProcInfoProcessList(pid int32) (ProcInfo, error) {
	s.processListLock.Lock()
	defer s.processListLock.Unlock()

	for _, p := range s.processList {
		if p.PID == pid {
			return p, nil
		}
	}

	return ProcInfo{}, fmt.Errorf("pid '%v' not found", pid)
}

func (s *Stat) isWatched(pid int32) bool {
	s.watchedLock.Lock()
	defer s.watchedLock.Unlock()

	if _, ok := s.watched[pid]; ok {
		return true
	}

	return false
}

// Start gathering watched for a specific process
func (s *Stat) StartWatchProcess(pid int32) error {
	// Is this is known pid?
	procInfo, err := s.getProcInfoProcessList(pid)
	if err != nil {
		return fmt.Errorf("pid '%v' is not in process list", pid)
	}

	// Is the process already being watched?
	if s.isWatched(pid) {
		return AlreadyWatchedErr
	}

	// Get internals ready
	procInfo.Metrics = make([]ProcInfoMetrics, 0)
	procInfo.MetricsLock = &sync.Mutex{}

	// Instantiate process
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("unable to instantiate process: %v", err)
	}

	// Is the process still running?
	if _, err := proc.Status(); err != nil {
		return fmt.Errorf("unable to fetch initial status for pid '%v': %v", pid, err)
	}

	looper := director.NewImmediateTimedLooper(director.FOREVER, StatInterval, nil)

	s.watched[pid] = Proc{
		ProcInfo: procInfo,
		Process:  proc,
		Looper:   looper,
	}

	watchedProc := s.watched[pid]

	// Gather watched in a goroutine
	go func(watchedProc *Proc) {
		// Stop watching process if loop ever exits
		defer func(pid int32) {
			// TODO: Should we do anything smarter here?
			if err := s.StopWatchProcess(pid); err != nil {
				sugar.Errorf("unable to stop watching pid '%v': %v", pid, err)
			}
		}(watchedProc.Process.Pid)

		watchedProc.Looper.Loop(func() error {
			sugar.Debugf("Fetching watched for stat %v", watchedProc.Process.Pid)

			// Is the process still around?
			if _, err := watchedProc.Process.Status(); err != nil {
				fullErr := fmt.Errorf("cannot fetch watched pid '%v' status (no longer running?): %v", pid, err)
				sugar.Error(fullErr)
				return fullErr
			}

			// Generate watched for the process
			metrics, err := s.getMetrics(watchedProc.Process)
			if err != nil {
				fullErr := fmt.Errorf("unable to fetch metrics for pid '%v': %v", pid, err)
				sugar.Error(fullErr)
				return fullErr
			}

			// Save metrics
			watchedProc.ProcInfo.MetricsLock.Lock()
			watchedProc.ProcInfo.Metrics = append(watchedProc.ProcInfo.Metrics, *metrics)
			watchedProc.ProcInfo.MetricsLock.Unlock()

			return nil
		})
	}(&watchedProc)

	return nil
}

func (s *Stat) getMetrics(proc *process.Process) (*ProcInfoMetrics, error) {
	meminfo, err := proc.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch memory info: %v", err)
	}

	percent, err := proc.Percent(0)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch CPU usage info: %v", err)
	}

	threads, err := proc.NumThreads()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch thread count: %v", err)
	}

	return &ProcInfoMetrics{
		RSS:       meminfo.RSS,
		VMS:       meminfo.VMS,
		Swap:      meminfo.Swap,
		CPU:       percent,
		Threads:   threads,
		Timestamp: time.Now(),
	}, nil
}

// Stop gathering watched for a specific process
func (s *Stat) StopWatchProcess(pid int32) error {
	// Is the PID actively being watched?

	// Looper stop
	// TODO: Can we stop an already stopped looper? (in case .Loop returned err)

	// Remove map entry

	return nil
}

// Get statistics for a specific stat
func (s *Stat) GetStatsForPID(pid int32) ([]ProcInfo, error) {
	return nil, NotWatchedErr
}

// Get all watched
func (s *Stat) GetStats() (map[string][]ProcInfo, error) {
	return nil, fmt.Errorf("shit broke")
}
