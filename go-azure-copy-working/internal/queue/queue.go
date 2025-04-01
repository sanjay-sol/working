package queue

import (
	"sync"
	"time"

	"goaz/internal/models"
)

type Queue struct {
	mu    sync.Mutex
	tasks []models.ImageTask
	cond  *sync.Cond
}

func NewQueue() *Queue {
	q := &Queue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Push(task models.ImageTask) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, task)
	q.cond.Signal()
}

// Regular Pop() method (blocking)
func (q *Queue) Pop() (models.ImageTask, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.tasks) == 0 {
		q.cond.Wait()
	}
	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task, true
}

// ✅ Fixed `PopWithTimeout()`
func (q *Queue) PopWithTimeout(timeout time.Duration) (models.ImageTask, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Use a timer to enforce the timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for len(q.tasks) == 0 {
		q.mu.Unlock() // Unlock before waiting to prevent deadlock
		select {
		case <-timer.C: // Timeout reached
			q.mu.Lock() // Relock before returning
			return models.ImageTask{}, false
		default:
			time.Sleep(10 * time.Millisecond) // Small delay before checking again
		}
		q.mu.Lock() // Relock after waiting
	}

	// If tasks exist, pop the first one
	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task, true
}

// Optional: Close method to clear the queue
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = nil
}
