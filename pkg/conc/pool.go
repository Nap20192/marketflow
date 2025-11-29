package conc

import (
	"encoding/json"
	"fmt"
	"time"

	"marketflow/internal/core/model"
)

type PoolHandler interface {
	Handle(
		workerID int,
		task Task,
		result chan<- Result,
	)
}

type Pool struct {
	workers []*Worker
	Name    string
	pool    chan *Worker
	handler PoolHandler
}

func NewWorkerPool(workers []*Worker, handler PoolHandler) *Pool {
	return &Pool{
		workers: workers,
		pool:    make(chan *Worker, len(workers)),
		handler: handler,
	}
}

func (p *Pool) Create() {
	for _, worker := range p.workers {
		p.pool <- worker
	}
}

func (p *Pool) Work(task Task, result chan<- Result) {
	worker := <-p.pool
	go func(w *Worker, t Task, result chan<- Result) {
		p.handler.Handle(w.ID, t, result)
		w.CountTasks++
		p.pool <- w
	}(worker, task, result)
}

func (p *Pool) Wait() {
	for range len(p.workers) {
		<-p.pool
	}
}

func (p *Pool) PrintStats() {
	report := fmt.Sprintf("------------------------------- Worker Pool: %s Stats -------------------------------\n", p.Name)
	for _, w := range p.workers {
		report += w.Stat() + "\n"
	}
	fmt.Println(report)
}

type HandlerFunc func(workerID int, task Task, result chan<- Result)

func (f HandlerFunc) Handle(workerID int, task Task, result chan<- Result) {
	f(workerID, task, result)
}

type Result struct {
	Name      string
	Symbol    string
	Price     float64
	Timestamp int64

	Err error
}

var HandlerExample = HandlerFunc(func(workerID int, task Task, result chan<- Result) {
	time.Sleep(5 * time.Millisecond)
	var trade model.Trade

	data := []byte(task.Data)

	err := json.Unmarshal(data, &trade)
	fmt.Println(trade)
	if err != nil {
		fmt.Printf("Worker %d failed to unmarshal task data: %v\n", workerID, err)
		return
	}
})
