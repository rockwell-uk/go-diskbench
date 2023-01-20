package diskbench

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rockwell-uk/go-progress/progress"
	"github.com/rockwell-uk/go-utils/fileutils"
	"github.com/rockwell-uk/go-utils/timeutils"
)

var (
	start time.Time
)

type SequentialWritesJob progress.Job

type DiskBench struct {
	Folder  string
	Seconds int
}

type SequentialWritesResult struct {
	FilesWritten int
	Took         time.Duration
}

func (j *SequentialWritesJob) Setup(jobName string, input interface{}) (*progress.Job, error) {

	if bench, ok := input.(DiskBench); ok {

		var tasks []*progress.Task
		for i := 1; i <= bench.Seconds; i++ {
			tasks = append(tasks, &progress.Task{
				ID:        strconv.Itoa(i),
				Magnitude: 1,
			})
		}

		job := progress.SetupJob(jobName, tasks)

		return job, nil
	}

	return nil, fmt.Errorf("unexpected type %T", input)
}

func (j *SequentialWritesJob) Task(job *progress.Job, input interface{}) (interface{}, error) {
	return struct{}{}, nil
}

func (j *SequentialWritesJob) Run(job *progress.Job, input interface{}) (interface{}, error) {

	start = time.Now()
	var took time.Duration

	if bench, ok := input.(DiskBench); ok {

		var duration time.Duration = time.Second * time.Duration(bench.Seconds)
		var filesWritten int

		var condition bool

		for ok := true; ok; ok = condition {

			condition = time.Since(start) < duration

			elapsed := time.Since(start)

			f := fmt.Sprintf("%v/%v.txt", bench.Folder, filesWritten)
			d, _ := fileutils.GetFile(f)
			err := writeLines(d, duration)
			if err != nil {
				d.Close()
				condition = false
			}
			d.Close()
			filesWritten++

			sPassed := int(elapsed.Seconds())
			for i := 1; i < int(duration.Seconds()); i++ {
				if sPassed >= i {
					task, _ := job.GetTask(strconv.Itoa(i))
					if task.StartTime == nil {
						task.Start()
					}
					if task.Took == nil {
						task.End()
					}
					nTask, _ := job.GetTask(strconv.Itoa(i + 1))
					if nTask.StartTime == nil {
						nTask.Start()
					}
				}
			}
			job.UpdateBar()
		}

		nTask, _ := job.GetTask(strconv.Itoa(bench.Seconds))
		if nTask.Took == nil {
			nTask.End()
		}

		took = timeutils.Took(start)

		return SequentialWritesResult{
			FilesWritten: filesWritten,
			Took:         took,
		}, nil
	}

	return struct{}{}, nil
}

func writeLines(f *os.File, duration time.Duration) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{}, 1)

	go func(d *os.File, ctx context.Context) {

		for i := 0; i <= DiskBenchNumLines; i++ {

			if _, err := d.WriteString(fmt.Sprintf("%v\n", time.Now().UnixNano())); err != nil {
				return
			}

			select {
			default:
				if i == DiskBenchNumLines {
					done <- struct{}{}
				}
			case <-ctx.Done():
				return
			}
		}

	}(f, ctx)

	end := time.Until(start.Add(duration))

	select {
	case <-done:
		return nil
	case <-time.After(end):
		return ctx.Err()
	}
}
