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

package tupi

import (
	"bytes"
	"testing"
)

func TestTracef(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevelStr("trace")
	oldwriter := traceLogger.Writer()
	defer func() { traceLogger.SetOutput(oldwriter) }()

	var buf bytes.Buffer
	traceLogger.SetOutput(&buf)

	Tracef("oi")
	r := make([]byte, 10)
	buf.Read(r)
	if string(r) != "[TRACE] oi" {
		t.Fatalf("Bad log %s", string(r))
	}
}

func TestDebugf(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevelStr("debug")
	oldwriter := debugLogger.Writer()
	defer func() { debugLogger.SetOutput(oldwriter) }()

	var buf bytes.Buffer
	debugLogger.SetOutput(&buf)

	Debugf("oi")
	r := make([]byte, 10)
	buf.Read(r)
	if string(r) != "[DEBUG] oi" {
		t.Fatalf("Bad log %s", string(r))
	}
}

func TestInfof(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevelStr("info")
	oldwriter := infoLogger.Writer()
	defer func() { infoLogger.SetOutput(oldwriter) }()

	var buf bytes.Buffer
	infoLogger.SetOutput(&buf)

	Infof("oi")
	r := make([]byte, 9)
	buf.Read(r)
	if string(r) != "[INFO] oi" {
		t.Fatalf("Bad log %s", string(r))
	}
}

func TestWarninf(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevelStr("warning")
	oldwriter := warningLogger.Writer()
	defer func() { warningLogger.SetOutput(oldwriter) }()

	var buf bytes.Buffer
	warningLogger.SetOutput(&buf)

	Warningf("oi")
	r := make([]byte, 12)
	buf.Read(r)
	if string(r) != "[WARNING] oi" {
		t.Fatalf("Bad log %s", string(r))
	}
}

func TestErrorf(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevelStr("error")
	oldwriter := errorLogger.Writer()
	defer func() { errorLogger.SetOutput(oldwriter) }()

	var buf bytes.Buffer
	errorLogger.SetOutput(&buf)

	Errorf("oi")
	r := make([]byte, 10)
	buf.Read(r)
	if string(r) != "[ERROR] oi" {
		t.Fatalf("Bad log %s", string(r))
	}
}

func TestSetLogLevelStr_invalid(t *testing.T) {
	r := SetLogLevelStr("bad")
	if r == nil {
		t.Fatalf("No error for bad loglevel")
	}
}
