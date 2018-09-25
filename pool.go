// Package pool provides worker pool with job queue.
package pool

import (
	"log"
	"os"
	"sync"
	"time"
)

type (
	// IPool is the Pool interface.
	IPool interface {
		Start()
		Closed() bool
		Size() int
		Undispatch(...int)
		SetMaxPoolNum(int)
		SetLoadFactor(int)
		SetResizeSuccessThreshold(int)
		SetResizePeriodSeconds(int)
	}

	// Pool represents a pool with dispatcher.
	Pool struct {
		// JobQueue channel for incoming job request,
		// user should NOT close this channel to stop Pool,
		// instead done channel is used for stopping Pool.
		JobQueue chan Job

		// JobHandlerGenerator is used for new JobHandler.
		JobHandlerGenerator JobHandlerGen

		poolConfig Config

		// current number of dispatcher
		poolNum int

		// done channel signals the pool to stop,
		// user should stop pushing job request to JobQueue before closing done channel
		done chan struct{}

		doneMe chan struct{}

		// wait to drain JobQueue
		wg *sync.WaitGroup

		// closed is true when pool received a signal to stop
		closed bool

		mu *sync.Mutex
	}

	// Config used to init Pool.
	Config struct {
		// initial number of dispatcher
		InitPoolNum int

		// maximum number of dispatcher
		MaxPoolNum int

		// number of workers in a dispatcher
		WorkerNum int

		// LoadFactor determines number of jobs in JobQueue divided by number of workers in the Pool,
		// for example LoadFactor 20 means 1 worker handles 20 jobs in a second,
		// if current load exceeds LoadFactor then resizing number of pools upward
		LoadFactor int

		Resize

		// JobQueue channel buffer size
		JobQueueBufferSize int

		// Verbose logging mode if it's true, by default it's false
		Verbose bool
	}

	// Resize related config.
	Resize struct {
		// the number of times the check needs to succeed before running resize
		SuccessThreshold int
		// how often to check LoadFactor to determine whether to resize
		PeriodSeconds int
		// the number of second to wait after the Pool has started before running the check
		InitialDelaySeconds int
	}

	// Option applies config to Pool Config.
	Option = func(*Config) error

	// JobHandlerGen returns a JobHandler when it's called.
	JobHandlerGen = func() JobHandler
)

var (
	// DefaultConfig is the default Pool Config.
	DefaultConfig = Config{
		InitPoolNum: 1,
		MaxPoolNum:  5,
		WorkerNum:   500,
		LoadFactor:  20,
		Resize: Resize{
			SuccessThreshold:    2,
			PeriodSeconds:       10,
			InitialDelaySeconds: 60,
		},
		JobQueueBufferSize: 10000,
	}
)

// New creates a pool.
func New(done chan struct{}, jobHandlerGenerator JobHandlerGen, options ...Option) *Pool {
	pConfig := DefaultConfig
	setOption(&pConfig, options...)

	if pConfig.InitPoolNum <= 0 {
		log.Panicln("config InitPoolNum should not be less than 1")
	}

	if pConfig.MaxPoolNum < pConfig.InitPoolNum {
		log.Panicln("config MaxPoolNum should not be less than config InitPoolNum")
	}

	if pConfig.WorkerNum <= 0 {
		log.Panicln("config WorkerNum should not be less than 1")
	}

	if pConfig.Verbose {
		os.Setenv("Pool.Log.Verbose", "true")
	}

	return &Pool{
		JobQueue:            make(chan Job, pConfig.JobQueueBufferSize),
		poolConfig:          pConfig,
		JobHandlerGenerator: jobHandlerGenerator,
		done:                done,
		doneMe:              make(chan struct{}, pConfig.InitPoolNum),
		mu:                  &sync.Mutex{},
		wg:                  &sync.WaitGroup{},
	}
}

// Start run dispatchers in the pool.
func (p *Pool) Start() {
	for range Range(p.poolConfig.InitPoolNum) {
		p.newDispatcher()
	}
	p.mu.Lock()
	p.poolNum = p.poolConfig.InitPoolNum
	p.mu.Unlock()
	go p.listen()
	go p.autoScale()
}

func (p *Pool) newDispatcher() {
	j := p.JobHandlerGenerator()
	p.wg.Add(1)
	d := NewDispatcher(p.doneMe, p.wg, p.poolConfig.WorkerNum, p.JobQueue, j)
	d.Run()
}

// listen for signals from done channel.
func (p *Pool) listen() {
	for {
		select {
		case _, open := <-p.done:
			if open {
				p.Undispatch()
			} else {
				p.mu.Lock()
				if !p.closed {
					p.closed = true
				}
				p.poolNum = 0
				p.mu.Unlock()
				close(p.JobQueue)
				p.wg.Wait()
				close(p.doneMe)
				return
			}
		}
	}
}

// autoScale check LoadFactor peridically to determine whether to resize.
func (p *Pool) autoScale() {
	time.Sleep(time.Duration(p.poolConfig.InitialDelaySeconds) * time.Second)
	count := 0
	t := time.NewTicker((time.Duration(p.poolConfig.PeriodSeconds) * time.Second))
	for {
		select {
		case <-t.C:
			p.mu.Lock()
			currentLoad := len(p.JobQueue) / (p.poolNum * p.poolConfig.WorkerNum)
			if currentLoad > p.poolConfig.LoadFactor {
				count++
			}
			if !p.closed && count == p.poolConfig.SuccessThreshold && p.poolNum < p.poolConfig.MaxPoolNum {
				count = 0
				p.newDispatcher()
				p.poolNum++
			}
			p.mu.Unlock()
		case _, open := <-p.doneMe:
			if !open {
				t.Stop()
				return
			}
		}
	}
}

// setOption takes one or more Option function and applies them in order to Pool Config.
func setOption(p *Config, options ...func(*Config) error) error {
	for _, opt := range options {
		if err := opt(p); err != nil {
			return err
		}
	}
	return nil
}

// SetMaxPoolNum applies MaxPoolNum to Pool Config.
func (p *Pool) SetMaxPoolNum(maxPoolNum int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	opt := func(c *Config) error {
		c.MaxPoolNum = maxPoolNum
		return nil
	}
	setOption(&p.poolConfig, opt)
}

// SetLoadFactor applies LoadFactor to Pool Config.
func (p *Pool) SetLoadFactor(loadFactor int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	opt := func(c *Config) error {
		c.LoadFactor = loadFactor
		return nil
	}
	setOption(&p.poolConfig, opt)
}

// SetResizeSuccessThreshold applies Resize SuccessThreshold to Pool Config.
func (p *Pool) SetResizeSuccessThreshold(resizeSuccessThreshold int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	opt := func(c *Config) error {
		c.SuccessThreshold = resizeSuccessThreshold
		return nil
	}
	setOption(&p.poolConfig, opt)
}

// SetResizePeriodSeconds applies Resize PeriodSeconds to Pool Config.
func (p *Pool) SetResizePeriodSeconds(resizePeriodSeconds int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	opt := func(c *Config) error {
		c.PeriodSeconds = resizePeriodSeconds
		return nil
	}
	setOption(&p.poolConfig, opt)
}

// Undispatch signals dispatcher to stop,
// num is the number of dispatcher to stop, default to 1.
func (p *Pool) Undispatch(num ...int) {
	n := Min(1, p.poolNum)
	if len(num) > 0 && num[0] > 1 {
		n = Min(num[0], p.poolNum)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.closed && p.poolNum-n > 0 {
		for range Range(n) {
			p.doneMe <- struct{}{}
			p.poolNum--
			p.wg.Done()
		}
	}
}

// Closed returns true if pool received a signal to stop.
func (p *Pool) Closed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

// Size returns current number of dispatcher.
func (p *Pool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.poolNum
}

func verbose() (v bool) {
	if value, ok := os.LookupEnv("Pool.Log.Verbose"); ok && value == "true" {
		v = true
	}
	return
}

// Range creates a range progressing from zero up to, but not including end.
func Range(end int) []struct{} {
	return make([]struct{}, end)
}

// Max returns the larger of x or y.
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min returns the smaller of x or y.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
