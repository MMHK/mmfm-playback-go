package logger

import (
	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("mmfm-playback")

func init() {
	format := logging.MustStringFormatter(
		`%{color}mmfm-playback %{shortfunc} %{level:.4s} %{shortfile}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "mmfm-playback")
}
