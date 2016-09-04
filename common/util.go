package common

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("jbuild")
)

func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}

	return false
}

type CmdSpec struct {
	Cmd      *exec.Cmd
	Result   chan error
	Complete func(error, time.Duration)
}

type TestResult struct {
	Passed   bool
	Result   string
	Duration time.Duration
}

func SaveTestResult(testExe string, passed bool, result string, duration time.Duration) {
	fileName := testExe + ".result"
	file, err := os.Create(fileName)
	defer file.Close()

	if err != nil {
		log.Errorf("Could not save cached test result file %s", fileName)
	}

	encoder := gob.NewEncoder(file)
	encoder.Encode(TestResult{passed, result, duration})
}

func LoadTestResult(testExe string) *TestResult {
	fileName := testExe + ".result"
	if !FileExists(fileName) {
		return nil
	}

	// Check if the given exe is newer than the cache; if it is, then we shouldn't
	// return the cached results.
	exeStat, _ := os.Stat(testExe)
	cacheStat, _ := os.Stat(fileName)
	if exeStat.ModTime().After(cacheStat.ModTime()) {
		return nil
	}

	result := new(TestResult)
	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		log.Errorf("Could not load cached test result file %s", fileName)
	}

	decoder := gob.NewDecoder(file)
	decoder.Decode(result)
	return result
}

func RunCommand(cmd *exec.Cmd, result chan error, complete func(error, time.Duration)) {
	// Print the command.
	if DryRun {
		log.Infof("DRY_RUN: %s", cmd.Args)
		complete(nil)
		result <- nil
		return
	} else {
		log.Debug(cmd.Args)
	}

	// Save the command output.
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Run the command.
	startTime := time.Now()
	err := cmd.Run()
	elaspedTime := time.Since(startTime)
	if err != nil {
		if out.String() != "" {
			complete(errors.New(out.String()), elaspedTime)
			result <- errors.New(out.String())
		} else {
			complete(err, elaspedTime)
			result <- err
		}

		return
	}

	complete(nil, elaspedTime)
	result <- nil
}
