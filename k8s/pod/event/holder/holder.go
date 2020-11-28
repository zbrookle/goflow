package holder

import (
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
