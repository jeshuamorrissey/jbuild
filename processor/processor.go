package processor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jeshuam/jbuild/common"
	"github.com/jeshuam/jbuild/config"
	"github.com/jeshuam/jbuild/processor/cc"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("jbuild")
)

type ProcessingResult struct {
	Target *config.Target
	Err    error
}

type Processor interface {
	// Process the given target using this processor.
	Process(*config.Target, chan common.CmdSpec) error
}

func (e *ProcessingResult) Error() string {
	return fmt.Sprintf("Processing target %s failed: %s", e.Target, e.Err)
}

func makeProcessingResult(target *config.Target, err error) ProcessingResult {
	return ProcessingResult{target, err}
}

func Process(target *config.Target, ch chan ProcessingResult, taskQueue chan common.CmdSpec) error {
	// Switch on the processor type.
	var p Processor
	if strings.HasPrefix(target.Type, "c++/") {
		p = new(cc.CCProcessor)
	} else {
		return errors.New(fmt.Sprintf("Unknown target type '%s'", target.Type))
	}

	// Process the target.
	go func() {
		err := p.Process(target, taskQueue)
		target.Processed = true
		ch <- makeProcessingResult(target, err)
	}()

	return nil
}