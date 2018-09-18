

# pool
`import "github.com/andy2046/pool"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-overview">Overview</a>
Package pool provides worker pool with job queue.




## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func Max(x, y int) int](#Max)
* [func Min(x, y int) int](#Min)
* [func Range(end int) []struct{}](#Range)
* [type Config](#Config)
* [type Dispatcher](#Dispatcher)
  * [func NewDispatcher(done chan struct{}, wgPool *sync.WaitGroup, numWorkers int, jobQueue chan Job, jobHandler JobHandler) *Dispatcher](#NewDispatcher)
  * [func (d *Dispatcher) Closed() bool](#Dispatcher.Closed)
  * [func (d *Dispatcher) DeWorker(num ...int)](#Dispatcher.DeWorker)
  * [func (d *Dispatcher) Run()](#Dispatcher.Run)
* [type IDispatcher](#IDispatcher)
* [type IPool](#IPool)
* [type IWorker](#IWorker)
* [type Job](#Job)
* [type JobHandler](#JobHandler)
* [type JobHandlerGen](#JobHandlerGen)
* [type Option](#Option)
* [type Pool](#Pool)
  * [func New(done chan struct{}, jobHandlerGenerator JobHandlerGen, options ...Option) *Pool](#New)
  * [func (p *Pool) Closed() bool](#Pool.Closed)
  * [func (p *Pool) SetLoadFactor(loadFactor int)](#Pool.SetLoadFactor)
  * [func (p *Pool) SetMaxPoolNum(maxPoolNum int)](#Pool.SetMaxPoolNum)
  * [func (p *Pool) SetResizePeriodSeconds(resizePeriodSeconds int)](#Pool.SetResizePeriodSeconds)
  * [func (p *Pool) SetResizeSuccessThreshold(resizeSuccessThreshold int)](#Pool.SetResizeSuccessThreshold)
  * [func (p *Pool) Size() int](#Pool.Size)
  * [func (p *Pool) Start()](#Pool.Start)
  * [func (p *Pool) Undispatch(num ...int)](#Pool.Undispatch)
* [type Resize](#Resize)
* [type Worker](#Worker)
  * [func NewWorker(done chan struct{}, workerPool chan chan Job, wg *sync.WaitGroup, jobPool chan struct{}) *Worker](#NewWorker)
  * [func (w *Worker) Closed() bool](#Worker.Closed)
  * [func (w *Worker) Start(handler JobHandler)](#Worker.Start)

#### <a name="pkg-examples">Examples</a>
* [Pool](#example_Pool)

#### <a name="pkg-files">Package files</a>
[dispatch.go](./dispatch.go) [pool.go](./pool.go) [work.go](./work.go) 



## <a name="pkg-variables">Variables</a>
``` go
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
```


## <a name="Max">func</a> [Max](./pool.go?s=6849:6871#L300)
``` go
func Max(x, y int) int
```
Max returns the larger of x or y.



## <a name="Min">func</a> [Min](./pool.go?s=6951:6973#L308)
``` go
func Min(x, y int) int
```
Min returns the smaller of x or y.



## <a name="Range">func</a> [Range](./pool.go?s=6746:6776#L295)
``` go
func Range(end int) []struct{}
```
Range creates a range progressing from zero up to, but not including end.




## <a name="Config">type</a> [Config](./pool.go?s=1070:1584#L54)
``` go
type Config struct {
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
}
```
Config used to init Pool.










## <a name="Dispatcher">type</a> [Dispatcher](./dispatch.go?s=230:1035#L17)
``` go
type Dispatcher struct {
    // contains filtered or unexported fields
}
```
Dispatcher represents the dispatcher that dispatch the job.







### <a name="NewDispatcher">func</a> [NewDispatcher](./dispatch.go?s=1078:1210#L56)
``` go
func NewDispatcher(done chan struct{}, wgPool *sync.WaitGroup, numWorkers int, jobQueue chan Job, jobHandler JobHandler) *Dispatcher
```
NewDispatcher creates a dispatcher.





### <a name="Dispatcher.Closed">func</a> (\*Dispatcher) [Closed](./dispatch.go?s=2713:2747#L124)
``` go
func (d *Dispatcher) Closed() bool
```
Closed returns true if dispatcher received a signal to stop.




### <a name="Dispatcher.DeWorker">func</a> (\*Dispatcher) [DeWorker](./dispatch.go?s=2895:2936#L132)
``` go
func (d *Dispatcher) DeWorker(num ...int)
```
DeWorker signals worker to stop,
num is the number of workers to stop, default to 1.




### <a name="Dispatcher.Run">func</a> (\*Dispatcher) [Run](./dispatch.go?s=1671:1697#L74)
``` go
func (d *Dispatcher) Run()
```
Run creates the workers pool and dispatches available jobs.




## <a name="IDispatcher">type</a> [IDispatcher](./dispatch.go?s=94:163#L10)
``` go
type IDispatcher interface {
    Run()
    Closed() bool
    DeWorker(...int)
}
```
IDispatcher is the Dispatcher interface.










## <a name="IPool">type</a> [IPool](./pool.go?s=143:328#L12)
``` go
type IPool interface {
    Start()
    Closed() bool
    Size() int
    Undispatch(...int)
    SetMaxPoolNum(int)
    SetLoadFactor(int)
    SetResizeSuccessThreshold(int)
    SetResizePeriodSeconds(int)
}
```
IPool is the Pool interface.










## <a name="IWorker">type</a> [IWorker](./work.go?s=364:422#L21)
``` go
type IWorker interface {
    Start(JobHandler)
    Closed() bool
}
```
IWorker is the Worker interface.










## <a name="Job">type</a> [Job](./work.go?s=87:261#L10)
``` go
type Job struct {
    Name     string `json:"name,omitempty"`
    ID       int64  `json:"id,omitempty"`
    Key      string `json:"key,omitempty"`
    WorkLoad []byte `json:"workload"`
}
```
Job represents the job to be run.










## <a name="JobHandler">type</a> [JobHandler](./work.go?s=298:324#L18)
``` go
type JobHandler func(Job) error
```
JobHandler completes the job.










## <a name="JobHandlerGen">type</a> [JobHandlerGen](./pool.go?s=2060:2093#L89)
``` go
type JobHandlerGen = func() JobHandler
```
JobHandlerGen returns a JobHandler when it's called.










## <a name="Option">type</a> [Option](./pool.go?s=1972:2000#L86)
``` go
type Option = func(*Config) error
```
Option applies config to Pool Config.










## <a name="Pool">type</a> [Pool](./pool.go?s=375:1037#L24)
``` go
type Pool struct {
    // JobQueue channel for incoming job request,
    // user should NOT close this channel to stop Pool,
    // instead done channel is used for stopping Pool.
    JobQueue chan Job

    // JobHandlerGenerator is used for new JobHandler.
    JobHandlerGenerator JobHandlerGen
    // contains filtered or unexported fields
}
```
Pool represents a pool with dispatcher.







### <a name="New">func</a> [New](./pool.go?s=2413:2501#L109)
``` go
func New(done chan struct{}, jobHandlerGenerator JobHandlerGen, options ...Option) *Pool
```
New creates a pool.





### <a name="Pool.Closed">func</a> (\*Pool) [Closed](./pool.go?s=6455:6483#L281)
``` go
func (p *Pool) Closed() bool
```
Closed returns true if pool received a signal to stop.




### <a name="Pool.SetLoadFactor">func</a> (\*Pool) [SetLoadFactor](./pool.go?s=5215:5259#L228)
``` go
func (p *Pool) SetLoadFactor(loadFactor int)
```
SetLoadFactor applies LoadFactor to Pool Config.




### <a name="Pool.SetMaxPoolNum">func</a> (\*Pool) [SetMaxPoolNum](./pool.go?s=4972:5016#L217)
``` go
func (p *Pool) SetMaxPoolNum(maxPoolNum int)
```
SetMaxPoolNum applies MaxPoolNum to Pool Config.




### <a name="Pool.SetResizePeriodSeconds">func</a> (\*Pool) [SetResizePeriodSeconds](./pool.go?s=5787:5849#L250)
``` go
func (p *Pool) SetResizePeriodSeconds(resizePeriodSeconds int)
```
SetResizePeriodSeconds applies Resize PeriodSeconds to Pool Config.




### <a name="Pool.SetResizeSuccessThreshold">func</a> (\*Pool) [SetResizeSuccessThreshold](./pool.go?s=5483:5551#L239)
``` go
func (p *Pool) SetResizeSuccessThreshold(resizeSuccessThreshold int)
```
SetResizeSuccessThreshold applies Resize SuccessThreshold to Pool Config.




### <a name="Pool.Size">func</a> (\*Pool) [Size](./pool.go?s=6586:6611#L288)
``` go
func (p *Pool) Size() int
```
Size returns current number of dispatcher.




### <a name="Pool.Start">func</a> (\*Pool) [Start](./pool.go?s=3259:3281#L137)
``` go
func (p *Pool) Start()
```
Start run dispatchers in the pool.




### <a name="Pool.Undispatch">func</a> (\*Pool) [Undispatch](./pool.go?s=6108:6145#L262)
``` go
func (p *Pool) Undispatch(num ...int)
```
Undispatch signals dispatcher to stop,
num is the number of dispatcher to stop, default to 1.




## <a name="Resize">type</a> [Resize](./pool.go?s=1614:1927#L76)
``` go
type Resize struct {
    // the number of times the check needs to succeed before running resize
    SuccessThreshold int
    // how often to check LoadFactor to determine whether to resize
    PeriodSeconds int
    // the number of second to wait after the Pool has started before running the check
    InitialDelaySeconds int
}
```
Resize related config.










## <a name="Worker">type</a> [Worker](./work.go?s=481:828#L27)
``` go
type Worker struct {
    // contains filtered or unexported fields
}
```
Worker represents the worker that executes the job.







### <a name="NewWorker">func</a> [NewWorker](./work.go?s=863:974#L49)
``` go
func NewWorker(done chan struct{}, workerPool chan chan Job, wg *sync.WaitGroup, jobPool chan struct{}) *Worker
```
NewWorker creates a worker.





### <a name="Worker.Closed">func</a> (\*Worker) [Closed](./work.go?s=1813:1843#L90)
``` go
func (w *Worker) Closed() bool
```
Closed returns true if worker received a signal to stop.




### <a name="Worker.Start">func</a> (\*Worker) [Start](./work.go?s=1205:1247#L61)
``` go
func (w *Worker) Start(handler JobHandler)
```
Start pushes the worker into worker queue, listens for signal to stop.




