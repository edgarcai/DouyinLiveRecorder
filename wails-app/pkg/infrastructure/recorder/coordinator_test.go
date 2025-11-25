package recorder

import (
    "context"
    "testing"
    "douyinrecorder/pkg/domain"
)

type fakeRunner struct{}

func (f *fakeRunner) BuildArgs(task domain.RecordTask) []string { return []string{"-i", "http://"} }
func (f *fakeRunner) Start(ctx context.Context, task domain.RecordTask) (string, error) { return "id", nil }
func (f *fakeRunner) Stop(id string) error { return nil }

// TestCoordinatorStartStop 验证登记与调用流程。
func TestCoordinatorStartStop(t *testing.T) {
    c := &Coordinator{tasks: make(map[string]domain.RecordTask), run: &fakeRunner{}}
    task := domain.RecordTask{Room: domain.RoomInfo{Platform: "d", AnchorName: "a"}}
    if err := c.StartRecord(task); err != nil { t.Fatal(err) }
    if len(c.tasks) != 1 { t.Fatalf("tasks len=%d", len(c.tasks)) }
    if err := c.StopRecord("d:a"); err != nil { t.Fatal(err) }
}

