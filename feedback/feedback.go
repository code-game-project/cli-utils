package feedback

import (
	"fmt"
)

const FeedbackPkg = Package("feedback")

type Package string

type Severity string

const (
	SeverityDebug Severity = "debug"
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
	SeverityFatal Severity = "fatal"
)

type ProgressCallback func(pkg Package, key, message string, current, total float64)

type FeedbackReceiver interface {
	Log(pkg Package, severity Severity, message string)
	Progress(pkg Package, process, message string, current, total float64)
}

var (
	feedbackReceiver    FeedbackReceiver
	enabled             bool
	disabledLogPackages map[Package]int              = make(map[Package]int)
	interceptProgress   map[Package]ProgressCallback = make(map[Package]ProgressCallback)
)

func Progress(pkg Package, process, message string, current, total float64) {
	if !enabled {
		return
	}
	if pr, ok := interceptProgress[pkg]; ok {
		if pr != nil {
			pr(pkg, process, message, current, total)
		}
	} else {
		feedbackReceiver.Progress(pkg, process, message, current, total)
	}
}

func Debug(pkg Package, msgFormat string, msgArgs ...any) {
	Log(pkg, SeverityDebug, msgFormat, msgArgs...)
}

func Info(pkg Package, msgFormat string, msgArgs ...any) {
	Log(pkg, SeverityInfo, msgFormat, msgArgs...)
}

func Warn(pkg Package, msgFormat string, msgArgs ...any) {
	Log(pkg, SeverityWarn, msgFormat, msgArgs...)
}

func Error(pkg Package, msgFormat string, msgArgs ...any) {
	Log(pkg, SeverityError, msgFormat, msgArgs...)
}

func Fatal(pkg Package, msgFormat string, msgArgs ...any) {
	Log(pkg, SeverityFatal, msgFormat, msgArgs...)
}

func Log(pkg Package, severity Severity, msgFormat string, msgArgs ...any) {
	if !enabled || (severity != SeverityDebug && disabledLogPackages[pkg] > 0) {
		return
	}
	feedbackReceiver.Log(pkg, severity, fmt.Sprintf(msgFormat, msgArgs...))
}

func Enable(receiver FeedbackReceiver) {
	if receiver != nil {
		enabled = true
	}
	feedbackReceiver = receiver
}

func Disable() {
	enabled = false
}

func Reenable() {
	if feedbackReceiver != nil {
		enabled = true
	}
}

func DisableLog(pkg Package) {
	disabledLogPackages[pkg] += 1
}

func ReenableLog(pkg Package) {
	disabledLogPackages[pkg] -= 1
	if disabledLogPackages[pkg] == 0 {
		delete(disabledLogPackages, pkg)
	}
}

func InterceptProgress(target Package, progressCallback ProgressCallback) {
	interceptProgress[target] = progressCallback
}

func UninterceptProgress(target Package) {
	delete(interceptProgress, target)
}
