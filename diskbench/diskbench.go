package diskbench

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rockwell-uk/go-progress/progress"
	"github.com/rockwell-uk/go-utils/fileutils"
)

const (
	DiskBenchNumLines = 10000
	FolderName        = "bench"
)

type DiskBenchResult struct {
	Performed bool
	Path      string
	NumLines  int
	Writes    int
	JobName   string
	Duration  time.Duration
}

func (r DiskBenchResult) String() string {
	return fmt.Sprintf("\t\t\t"+"Performed: %v"+"\n"+
		"\t\t\t"+"Path: %v"+"\n"+
		"\t\t\t"+"NumLines: %v"+"\n"+
		"\t\t\t"+"Writes: %v",
		r.Performed,
		r.Path,
		r.NumLines,
		r.Writes,
	)
}

func BenchDisk(d DiskBench) (DiskBenchResult, error) {
	var jobName string = "Benchmarking Disk"
	var took time.Duration
	var absPath string

	if d.Seconds < 1 {
		return DiskBenchResult{}, fmt.Errorf("seconds cannot be less than 1 [%v]", d.Seconds)
	}

	var targetFolder string = fmt.Sprintf("%v/%v", d.Folder, FolderName)

	if !fileutils.FolderExists(targetFolder) {
		err := fileutils.MkDir(targetFolder)
		if err != nil {
			return DiskBenchResult{}, fmt.Errorf("unable to create folder %v", targetFolder)
		}
	}

	absPath, err := filepath.Abs(targetFolder)
	if err != nil {
		return DiskBenchResult{}, err
	}

	// Sequential Writes Job
	var j progress.ProgressJob = &SequentialWritesJob{}
	job, err := j.Setup(jobName, d)
	if err != nil {
		return DiskBenchResult{}, err
	}
	defer job.End(true)

	res, err := j.Run(job, d)
	if err != nil {
		return DiskBenchResult{}, err
	}

	if _, ok := res.(SequentialWritesResult); !ok {
		return DiskBenchResult{}, fmt.Errorf("incorrect result from job %v", res)
	}

	sequentialWritesResult, ok := res.(SequentialWritesResult)
	if !ok {
		return DiskBenchResult{}, err
	}

	// brutally remove the target folder
	os.Remove(targetFolder)

	return DiskBenchResult{
		Performed: true,
		Path:      absPath,
		NumLines:  DiskBenchNumLines,
		Writes:    sequentialWritesResult.FilesWritten,
		JobName:   jobName,
		Duration:  took,
	}, nil
}
