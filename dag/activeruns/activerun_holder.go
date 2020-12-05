package activeruns

import "sync"

// ActiveRuns tracks how many dag runs are actively running
type ActiveRuns struct {
	lock  *sync.Mutex
	count int
}

// New returns a new instance of ActiveRuns
func New() *ActiveRuns {
	return &ActiveRuns{&sync.Mutex{}, 0}
}

// Get returns the number of active runs
func (runs *ActiveRuns) Get() (count int) {
	runs.lock.Lock()
	count = runs.count
	runs.lock.Unlock()
	return
}

// Inc increments the number of runs
func (runs *ActiveRuns) Inc() {
	runs.lock.Lock()
	runs.count++
	runs.lock.Unlock()
}

// Dec decrements the number of runs
func (runs *ActiveRuns) Dec() {
	runs.lock.Lock()
	runs.count--
	runs.lock.Unlock()
}
