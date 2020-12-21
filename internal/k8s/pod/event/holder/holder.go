package holder

import (
	"fmt"
	"goflow/internal/k8s/pod/event/channel"
	"sync"
)

// ChannelHolder wraps around a dictionary to hold the various channels to be accessed by pod watcher
type ChannelHolder struct {
	channelMap map[string]*channel.FuncChannelGroup
	lock       sync.RWMutex
}

// New creates a new channel holder
func New() *ChannelHolder {
	return &ChannelHolder{
		make(map[string]*channel.FuncChannelGroup),
		sync.RWMutex{},
	}
}

func (holder *ChannelHolder) lockUnlockW(concurrentFunc func()) {
	holder.lock.Lock()
	concurrentFunc()
	holder.lock.Unlock()
}

func (holder *ChannelHolder) concurrentChanMapRead(name string) (*channel.FuncChannelGroup, bool) {
	holder.lock.RLock()
	group, ok := holder.channelMap[name]
	holder.lock.RUnlock()
	return group, ok
}

// AddChannelGroup adds a new channel group with the given name
func (holder *ChannelHolder) AddChannelGroup(name string) {
	holder.lockUnlockW(func() { holder.channelMap[name] = channel.New() })
}

// DeleteChannelGroup deletes a channel group with the given name
func (holder *ChannelHolder) DeleteChannelGroup(name string) {
	holder.lockUnlockW(func() { delete(holder.channelMap, name) })
}

// GetChannelGroup returns the channel group for the given pod name
func (holder *ChannelHolder) GetChannelGroup(name string) *channel.FuncChannelGroup {
	group, ok := holder.concurrentChanMapRead(name)
	if !ok {
		panic(fmt.Sprintf("Group for pod %s not found!", name))
	}
	return group
}

// Contains returns true if the given name is in the channel holder
func (holder *ChannelHolder) Contains(name string) bool {
	_, ok := holder.concurrentChanMapRead(name)
	return ok
}

// List returns a slice containing pointers to all of the the channel groups
func (holder *ChannelHolder) List() []*channel.FuncChannelGroup {
	groupList := make([]*channel.FuncChannelGroup, len(holder.channelMap))
	for key := range holder.channelMap {
		group, _ := holder.concurrentChanMapRead(key)
		groupList = append(groupList, group)
	}
	return groupList
}
