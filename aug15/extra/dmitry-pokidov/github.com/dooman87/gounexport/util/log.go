//Package util provides simple wrapper on top of log package.
//You can set the logging level using Level variable.
package util

import (
	"log"
)

var (
	//Level is current level of logging. See #Levels
	Level = "INFO"
	//Levels is all possible logging levels
	Levels = map[string]int{
		"FATAL": 1,
		"ERROR": 2,
		"WARN":  3,
		"INFO":  4,
		"DEBUG": 5,
	}
)

//Fatalf prints fatal message and exit the program
func Fatalf(message string, fmt ...interface{}) {
	_log("FATAL", message, fmt...)
}

//Err prints error message if the level enabled
func Err(message string, fmt ...interface{}) {
	_log("ERROR", message, fmt...)
}

//Info prints info message if the level enabled
func Info(message string, fmt ...interface{}) {
	_log("INFO", message, fmt...)
}

//Warn prints warning message if the level enabled
func Warn(message string, fmt ...interface{}) {
	_log("WARN", message, fmt...)
}

//Debug prints debug message if the level enabled
func Debug(message string, fmt ...interface{}) {
	_log("DEBUG", message, fmt...)
}

func _log(level string, message string, fmt ...interface{}) {
	if !isLevelEnabled(level) {
		return
	}
	switch level {
	case "FATAL":
		log.Fatalf(level+": "+message, fmt...)
	default:
		log.Printf(level+": "+message, fmt...)
	}
}

func isLevelEnabled(lvl string) bool {
	return Levels[lvl] <= Levels[Level]
}
