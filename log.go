// Copyright 2023 Juca Crispim <juca@poraodojuca.net>

// This file is part of tupi.

// tupi is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// tupi is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with tupi. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"log"
	"os"
)

type logLevel int

const (
	LevelTrace logLevel = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
)

var traceLogger log.Logger = *log.New(os.Stdout, "[TRACE] ", 0)
var debugLogger log.Logger = *log.New(os.Stdout, "[DEBUG] ", 0)
var infoLogger log.Logger = *log.New(os.Stdout, "[INFO] ", 0)
var warningLogger log.Logger = *log.New(os.Stderr, "[WARNING] ", 0)
var errorLogger log.Logger = *log.New(os.Stderr, "[ERROR] ", 0)

var currentLogLevel logLevel = LevelInfo

func SetLogLevel(level logLevel) {
	currentLogLevel = level
}

func GetLogLevel() logLevel {
	return currentLogLevel
}

func Debugf(format string, v ...interface{}) {
	if currentLogLevel <= LevelDebug {
		debugLogger.Printf(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if currentLogLevel >= LevelInfo {
		infoLogger.Printf(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if currentLogLevel >= LevelError {
		errorLogger.Printf(format, v...)
	}
}
