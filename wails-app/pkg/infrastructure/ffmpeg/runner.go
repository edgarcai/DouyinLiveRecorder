package ffmpeg

import (
	"bufio"
	"context"
	"douyinrecorder/pkg/domain"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunnerAPI 定义FFmpeg运行器接口，便于测试与依赖注入。
type RunnerAPI interface {
	// BuildArgs 根据任务构造参数。
	BuildArgs(task domain.RecordTask) []string
	// Start 启动FFmpeg并返回任务ID（或进程标识）。
	Start(ctx context.Context, task domain.RecordTask) (string, error)
	// Stop 停止指定任务ID的进程。
	Stop(id string) error
}

// Runner 为默认实现。
type Runner struct {
	procs     map[string]*exec.Cmd
	onSegment func(string)
	onError   func(string, string)
}

// BuildArgs 根据录制任务构造 FFmpeg 参数列表。
func (r *Runner) BuildArgs(task domain.RecordTask) []string {
	input := task.Room.M3U8URL
	if input == "" && len(task.Room.PlayURLList) > 0 {
		input = task.Room.PlayURLList[0]
	}
	outDir := filepath.Dir(task.OutputPath)
	outFile := filepath.Base(task.OutputPath)

	args := []string{"-y", "-hide_banner", "-i", input}
	if task.SegmentSec > 0 {
		args = append(args, "-f", "segment", "-segment_time", fmt.Sprintf("%d", task.SegmentSec))
	}
	switch task.Format {
	case "mp4":
		args = append(args, "-c", "copy", filepath.Join(outDir, outFile+".mp4"))
	case "flv":
		args = append(args, "-c", "copy", filepath.Join(outDir, outFile+".flv"))
	case "mkv":
		args = append(args, "-c", "copy", filepath.Join(outDir, outFile+".mkv"))
	case "ts":
		args = append(args, "-c", "copy", filepath.Join(outDir, outFile+".ts"))
	default:
		args = append(args, "-c", "copy", filepath.Join(outDir, outFile+".ts"))
	}
	return args
}

// Start 启动FFmpeg进程。
func (r *Runner) Start(ctx context.Context, task domain.RecordTask) (string, error) {
	if r.procs == nil {
		r.procs = make(map[string]*exec.Cmd)
	}
	args := r.BuildArgs(task)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stderr, _ := cmd.StderrPipe()
	// 解析 stderr 文本，分类错误与段事件（最小实现）。
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			typ := ClassifyFFmpegError(line)
			if typ != "unknown" && r.onError != nil {
				r.onError(typ, line)
			}
            // 简单段事件识别：匹配 Opening '... for writing' 作为分段输出
            if strings.Contains(line, "Opening ") && strings.Contains(line, " for writing") && r.onSegment != nil {
                path := ExtractSegmentPath(line)
                if path != "" { r.onSegment(path) }
            }
		}
	}()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	id := task.Room.Platform + ":" + task.Room.AnchorName + ":" + time.Now().Format("20060102150405")
	r.procs[id] = cmd
	return id, nil
}

// SetOnSegment 注册段事件回调。
func (r *Runner) SetOnSegment(fn func(string)) { r.onSegment = fn }

// SetOnError 注册错误事件回调。
func (r *Runner) SetOnError(fn func(string, string)) { r.onError = fn }

// Stop 终止FFmpeg进程。
func (r *Runner) Stop(id string) error {
	if r.procs == nil {
		return nil
	}
	cmd, ok := r.procs[id]
	if !ok {
		return nil
	}
	err := cmd.Process.Kill()
	delete(r.procs, id)
	return err
}
