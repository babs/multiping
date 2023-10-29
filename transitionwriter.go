package main

import (
	"bufio"
	"log"
	"os"
	"sync"
	"time"
)

type TransitionWriter struct {
	fh                 *os.File
	writer             *bufio.Writer
	lock               sync.Mutex
	writer_initialized bool
}

func (w *TransitionWriter) Init(filename string, quitFlag *bool) {
	var err error
	w.fh, err = os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	w.writer = bufio.NewWriter(w.fh)
	go func(w *TransitionWriter) {
		for !*quitFlag {
			w.lock.Lock()
			w.writer.Flush()
			w.lock.Unlock()
			time.Sleep(500 * time.Millisecond)
		}
	}(w)
	w.writer_initialized = true
}

func (w *TransitionWriter) WriteString(st string) {
	if w.writer_initialized {
		w.lock.Lock()
		w.writer.WriteString(st)
		w.lock.Unlock()
	}
}

func (w *TransitionWriter) Close() {
	if w.writer_initialized {
		w.writer.Flush()
		w.fh.Close()
	}
}
