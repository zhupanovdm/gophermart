package task

import (
	"context"
	"sync"
	"time"
)

// VoidTask is used to define absence of work
var VoidTask = Task(func(ctx context.Context) {})

// Task represents a unit of work to be done in specified context.
type Task func(ctx context.Context)

// Option can define some desired Task side effect.
type Option func(Task) Task

// With wraps receiver with Option from arguments.
func (t Task) With(options ...Option) Task {
	task := t
	for _, opt := range options {
		task = opt(task)
	}
	return task
}

// CompletionWait provides ability to track Task execution completion with the given sync.WaitGroup instance.
func CompletionWait(wg *sync.WaitGroup) Option {
	return func(task Task) Task {
		wg.Add(1)
		return func(ctx context.Context) {
			defer wg.Done()
			task(ctx)
		}
	}
}

// PeriodicRun repeatedly runs Task with the specified interval.
func PeriodicRun(interval time.Duration) Option {
	return func(task Task) Task {
		ticker := time.NewTicker(interval)

		return func(ctx context.Context) {
			child, cancel := context.WithCancel(ctx)
			defer cancel()

			done := child.Done()
			for active := true; active; {
				select {
				case <-ticker.C:
					task(child)
				case <-done:
					ticker.Stop()
					active = false
				}
			}
		}
	}
}
