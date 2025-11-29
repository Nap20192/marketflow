package conc

import "fmt"

type Worker struct {
	ID         int
	Name       string
	CountTasks int
}

func NewWorker(opts ...func(*Worker)) *Worker {
	worker := &Worker{}
	for _, opt := range opts {
		opt(worker)
	}
	return worker
}

func CreateNWorkers(n int, opts ...func(*Worker)) []*Worker {
	workers := make([]*Worker, n)
	for i := 0; i < n; i++ {
		optsWithID := append([]func(*Worker){WithID(i + 1)}, opts...)
		workers[i] = NewWorker(optsWithID...)
	}
	return workers
}

func WithID(id int) func(*Worker) {
	return func(w *Worker) {
		w.ID = id
	}
}

func WithName(name string) func(*Worker) {
	return func(w *Worker) {
		w.Name = name
	}
}

func (w *Worker) Stat() string {
	return fmt.Sprintf("Worker: %s | ID: %d | Tasks processed: %d", w.Name, w.ID, w.CountTasks)
}
