package writer

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	selectors = map[string]io.Writer{}
)

func register(name string, writer io.Writer) {
	if selectors == nil {
		selectors = make(map[string]io.Writer)
	}
	selectors[name] = writer
}

func Open(outputDirectory, filename string) (logger *log.Logger, err error) {
	var w io.Writer
	if writer, present := selectors[filename]; !present {
		file, fileOpenError := os.OpenFile(filepath.Join(outputDirectory, filename), os.O_CREATE|os.O_WRONLY, 0644)
		if fileOpenError != nil {
			return nil, fileOpenError
		}
		register(filename, file)
		w = file
	} else {
		w = writer
	}
	return log.New(w, "", 0), nil
}

func CloseAll() {
	for name := range selectors {
		Close(name)
	}
}

func Close(name string) {
	file := selectors[name].(*os.File)
	if file != nil {
		file.Close()
	}
}
