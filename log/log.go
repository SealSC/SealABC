/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package log

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
    "fmt"
)

type scLog struct {
	*logrus.Logger
}

type Config struct {
    Level   logrus.Level
    LogFile string
}

var Log *scLog

func SetUpLogger(cfg Config) {
	Log = &scLog{}
	Log.Logger = logrus.New()

	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		Log.Out = io.MultiWriter(os.Stdout, file)
	} else {
		Log.Info("Failed to log to file, using default stderr")
	}

	Log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
		DisableTimestamp: false,
		FullTimestamp:true,
		TimestampFormat: time.StampMilli,
	}

	// Only log the warning severity or above.
	Log.SetLevel(cfg.Level)
}

func (l *scLog) Trace(args ...interface{}) {
	args = getTraceInfo(args)
	l.Debug(args...)
}

func (l *scLog) Tracef(format string, args ...interface{}) {
	args = getTraceInfo(args)
	l.Debugf("%s "+format, args...)
}

func (l *scLog) Error(args ...interface{}) {
    args = getTraceInfo(args)
    l.Logger.Error(args)
}

func (l *scLog) Errorf(format string, args ...interface{}) {
    args = getTraceInfo(args)
    l.Logger.Errorf("%s "+format, args...)
}


func getTraceInfo(args ...interface{}) []interface{} {
	pc := make([]uintptr, 10)

	runtime.Callers(3, pc)
	var stackRows  = []interface{} {"\r\n"}
	for i := 0; i<10; i++ {
        f := runtime.FuncForPC(pc[i])
        if nil == f {
            break
        }
        file, line := f.FileLine(pc[i])
        fileName := filepath.Base(file)
        nameFull := f.Name()

        stackRows = append(stackRows, "func " + nameFull + "@" + fileName + ":" + strconv.Itoa(line - 1) + "\r\n")
    }

    ret :=  append([]interface{}{}, args...)
	return  append(ret, stackRows...)
}

func TracePanic() {
    r := recover()
    if nil != r {
        printPanic(r)
    }
}

func printPanic(r interface{})  {
    if nil == Log {
        trace := getTraceInfo()
        fmt.Println(trace...)
        fmt.Println("got a panic: ", r)
    } else {
        Log.Errorf("got a panic: %#v", r)
    }
}