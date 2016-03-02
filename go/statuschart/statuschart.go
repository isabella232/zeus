package statuschart

import (
	"github.com/burke/ttyutils"
	"github.com/stripe/zeus/go/processtree"
	slog "github.com/stripe/zeus/go/shinylog"
	"os"
	"sync"
)

type StatusChart struct {
	RootSlave *processtree.SlaveNode
	update    chan bool

	numberOfSlaves int
	Commands       []*processtree.CommandNode
	L              sync.Mutex
	drawnInitial   bool

	directLogger *slog.ShinyLogger

	extraOutput       string
	terminalSupported bool

	previousStates []*string
}

var theChart *StatusChart

func Start(tree *processtree.ProcessTree, done chan bool) chan bool {
	quit := make(chan bool)

	theChart = &StatusChart{}
	theChart.RootSlave = tree.Root
	theChart.numberOfSlaves = len(tree.SlavesByName)
	theChart.Commands = tree.Commands
	theChart.update = make(chan bool, 10)
	theChart.directLogger = slog.NewShinyLogger(os.Stdout, os.Stderr)
	theChart.terminalSupported = ttyutils.IsTerminal(os.Stdout.Fd())

	if theChart.terminalSupported {
		ttyStart(tree, done, quit)
	} else {
		stdoutStart(tree, done, quit)
	}
	return quit
}

func stateSuffix(state string) string {
	status := ""

	switch state {
	case processtree.SUnbooted:
		status = "{U}"
	case processtree.SBooting:
		status = "{B}"
	case processtree.SCrashed:
		status = "{!C}"
	case processtree.SReady:
		status = "{R}"
	default:
		status = "{?}"
	}

	return status
}

func printStateInfo(indentation, identifier, state string, verbose, printNewline bool) {
	log := theChart.directLogger
	newline := ""
	suffix := ""
	if printNewline {
		newline = "\n"
	}
	if verbose {
		suffix = stateSuffix(state)
	}
	switch state {
	case processtree.SUnbooted:
		log.ColorizedSansNl(indentation + "{magenta}" + identifier + suffix + "\033[K" + newline)
	case processtree.SBooting:
		log.ColorizedSansNl(indentation + "{blue}" + identifier + suffix + "\033[K" + newline)
	case processtree.SCrashed:
		log.ColorizedSansNl(indentation + "{red}" + identifier + suffix + "\033[K" + newline)
	case processtree.SReady:
		// no status suffix, as that's the optimal state
		log.ColorizedSansNl(indentation + "{green}" + identifier + suffix + "\033[K" + newline)
	case processtree.SWaiting:
		fallthrough
	default:
		log.ColorizedSansNl(indentation + "{yellow}" + identifier + suffix + "\033[K" + newline)
	}
}
