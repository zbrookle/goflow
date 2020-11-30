package holder

import (
	"fmt"
	"goflow/k8s/pod/event/channel"
)

// ChannelHolder wraps around a dictionary to hold the various channels to be accessed by pod watcher
type ChannelHolder struct {
	channelMap map[string]*channel.FuncChannelGroup
}

// New creates a new channel holder
func New() ChannelHolder {
	return ChannelHolder{
		make(map[string]*channel.FuncChannelGroup),
	}
}

// AddChannelGroup adds a new channel group with the given name
func (holder *ChannelHolder) AddChannelGroup(name string) {
	holder.channelMap[name] = channel.New()
}

// DeleteChannelGroup deletes a channel gropu with the given name
func (holder *ChannelHolder) DeleteChannelGroup(name string) {
	delete(holder.channelMap, name)
}

// GetChannelGroup returns the channel group for the given pod name
func (holder *ChannelHolder) GetChannelGroup(name string) *channel.FuncChannelGroup {
	group, ok := holder.channelMap[name]
	if !ok {
		panic(fmt.Sprintf("Group for pod %s not found!", name))
	}
	return group
}

// Contains returns true if the given name is in the channel holder
func (holder *ChannelHolder) Contains(name string) bool {
	_, ok := holder.channelMap[name]
	return ok
}

// List returns containing pointers to all of the the channel groups
func (holder *ChannelHolder) List() []*channel.FuncChannelGroup {
	groupList := make([]*channel.FuncChannelGroup, len(holder.channelMap))
	for key := range holder.channelMap {
		groupList = append(groupList, holder.channelMap[key])
	}
	return groupList
}
