package holder

import (
	"goflow/k8s/pod/event/channel"
	"testing"
)

const testName string = "test"

func makeTestHolder() ChannelHolder {
	return New()
}

func TestAddToHolder(t *testing.T) {
	holder := New()
	holder.AddChannelGroup(testName)

	_, ok := holder.channelMap[testName]
	if !ok {
		t.Errorf("FuncChannelGroup %s not found in group", testName)
	}
}

func TestDeleteFromHolder(t *testing.T) {
	holder := New()
	holder.channelMap[testName] = channel.New()
	holder.DeleteChannelGroup(testName)

	_, ok := holder.channelMap[testName]
	if ok {
		t.Errorf("%s should not be in channel map", testName)
	}
}

func TestGetGroupFromHolder(t *testing.T) {
	holder := New()
	channelGroup := channel.New()
	holder.channelMap[testName] = channelGroup

	retrievedGroup := holder.GetChannelGroup(testName)
	if retrievedGroup != channelGroup {
		t.Errorf("Channel groups do not match for pod with name %s", testName)
	}
}
