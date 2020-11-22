package podwatch

// import (
// 	"goflow/logs"
// 	"goflow/podutils"
// 	"testing"
// )

// func TestFindContainerCompleteEvent(t *testing.T) {
// 	logs.InfoLogger.Println("Testing Container complete event")
// 	defer podutils.CleanUpPods(KUBECLIENT)
// 	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient())
// 	dagRun.createPod()
// 	dagRun.callFuncUntilPodSucceedOrFail(func() {
// 		logs.InfoLogger.Println("I'm waiting...")
// 	})
// 	// dagRun.getLogsContainerNotFound()
// 	// panic("test")
// }
