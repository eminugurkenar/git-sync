/*
Copyright 2021 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
)

func NewLogger(root string, errorFile string) *Logger {
	return &Logger{glogr.New(), root, errorFile}
}

type Logger struct {
	logr.Logger
	Root      string
	ErrorFile string
}

func (l Logger) Error(err error, msg string, kvList ...interface{}) {
	l.Logger.Error(err, msg, kvList...)
	if l.ErrorFile == "" {
		return
	}
	payload := struct {
		Msg  string
		Err  string
		Args map[string]interface{}
	}{
		Msg:  msg,
		Err:  err.Error(),
		Args: map[string]interface{}{},
	}
	if len(kvList)%2 != 0 {
		kvList = append(kvList, "<no-value>")
	}
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			k = fmt.Sprintf("%v", kvList[i])
		}
		payload.Args[k] = kvList[i+1]
	}
	jb, err := json.Marshal(payload)
	if err != nil {
		l.Logger.Error(err, "can't encode error payload")
		content := fmt.Sprintf("%v", err)
		l.writeContent([]byte(content))
	} else {
		l.writeContent(jb)
	}
}

// exportError exports the error to the error file if --export-error is enabled.
func (l *Logger) ExportError(content string) {
	if l.ErrorFile == "" {
		return
	}
	l.writeContent([]byte(content))
}

// writeContent writes the error content to the error file.
func (l *Logger) writeContent(content []byte) {
	if _, err := os.Stat(l.Root); os.IsNotExist(err) {
		fileMode := os.FileMode(0755)
		if err := os.Mkdir(l.Root, fileMode); err != nil {
			l.Logger.Error(err, "can't create the root directory", "root", l.Root)
			return
		}
	}
	tmpFile, err := ioutil.TempFile(l.Root, "tmp-err-")
	if err != nil {
		l.Logger.Error(err, "can't create temporary error-file", "directory", l.Root, "prefix", "tmp-err-")
		return
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			l.Logger.Error(err, "can't close temporary error-file", "filename", tmpFile.Name())
		}
	}()

	if _, err = tmpFile.Write(content); err != nil {
		l.Logger.Error(err, "can't write to temporary error-file", "filename", tmpFile.Name())
		return
	}

	errorFile := filepath.Join(l.Root, l.ErrorFile)
	if err := os.Rename(tmpFile.Name(), errorFile); err != nil {
		l.Logger.Error(err, "can't rename to error-file", "temp-file", tmpFile.Name(), "error-file", errorFile)
		return
	}
	if err := os.Chmod(errorFile, 0644); err != nil {
		l.Logger.Error(err, "can't change permissions on the error-file", "error-file", errorFile)
	}
}

// deleteErrorFile deletes the error file.
func (l *Logger) DeleteErrorFile() {
	if l.ErrorFile == "" {
		return
	}
	errorFile := filepath.Join(l.Root, l.ErrorFile)
	if err := os.Remove(errorFile); err != nil {
		if os.IsNotExist(err) {
			return
		}
		l.Logger.Error(err, "can't delete the error-file", "filename", errorFile)
	}
}
