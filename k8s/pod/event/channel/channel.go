package channel

import (
	core "k8s.io/api/core/v1"
)

// FuncChannelGroup holds the various states of a pod's events'
type FuncChannelGroup struct {
	Ready  chan *core.Pod
	Update chan *core.Pod
	Remove chan *core.Pod
}

// New returns a new FuncChannel struct
func New() *FuncChannelGroup {
	return &FuncChannelGroup{
		make(chan *core.Pod, 1),
		make(chan *core.Pod, 1),
		make(chan *core.Pod, 1),
	}
}
