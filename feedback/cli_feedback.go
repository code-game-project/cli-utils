package feedback

import (
	"github.com/code-game-project/cli-utils/cli"
)

type CLIFeedback struct {
	severity Severity
}

func NewCLIFeedback(minSeverity Severity) *CLIFeedback {
	return &CLIFeedback{
		severity: minSeverity,
	}
}

func (c *CLIFeedback) Log(pkg Package, severity Severity, message string) {
	if severity < c.severity {
		return
	}
	switch severity {
	case SeverityDebug:
		cli.Print("[DEBUG] %s: %s", pkg, message)
	case SeverityInfo:
		cli.Print("%s", message)
	case SeverityWarn:
		cli.PrintColor(cli.Yellow, "[WARNING] %s", message)
	case SeverityError:
		cli.PrintColor(cli.Red, "[ERROR] %s", message)
	case SeverityFatal:
		cli.PrintColor(cli.RedBold, "[FATAL] %s", message)
	}
}

func (c *CLIFeedback) Progress(pkg Package, process, message string, current, total int64, unit cli.Unit) {
	cli.Progress(process, message, current, total, unit)
}
