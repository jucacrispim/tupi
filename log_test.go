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
	"bytes"
	"testing"
)

func TestDebugf(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevel(LevelDebug)
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
	SetLogLevel(LevelInfo)
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

func TestErrorf(t *testing.T) {
	oldlevel := GetLogLevel()
	defer func() { SetLogLevel(oldlevel) }()
	SetLogLevel(LevelError)
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
