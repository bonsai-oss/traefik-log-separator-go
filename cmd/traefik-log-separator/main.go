package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"traefik-log-separator/internal/model"
	"traefik-log-separator/internal/writer"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/alecthomas/kingpin.v2"
)

type parameters struct {
	logInputFile       string
	logOutputDirectory string
}

func init() {
	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)
}

var params parameters

func init() {
	app := kingpin.New(os.Args[0], "traefik-log-separator")
	app.HelpFlag.Short('h')
	app.Flag("input", "input access log path").Short('i').Envar("INPUT").Required().StringVar(&params.logInputFile)
	app.Flag("output", "output directory path").Short('o').Envar("OUTPUT").Required().StringVar(&params.logOutputDirectory)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func fileWorker(inotifySignal chan any) {
	file, _ := os.OpenFile(params.logInputFile, os.O_RDONLY, 0644)
	defer file.Close()

	reader := bufio.NewReader(file)
	for range inotifySignal {
		for {
			line, _, err := reader.ReadLine()

			if err == io.EOF {
				break
			}

			msg, messageDecodeError := model.LogMessage{}.Decode(string(line))
			if messageDecodeError != nil {
				log.Println(messageDecodeError)
				continue
			}

			if logger, writerOpenError := writer.Open(params.logOutputDirectory, msg.RouterName+".log"); writerOpenError != nil {
				log.Println(writerOpenError)
			} else {
				fmt.Println(string(line))
				logger.Println(string(line))
			}
		}
	}
}

func main() {
	workerSignal := make(chan os.Signal, 1)
	signal.Notify(workerSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	defer writer.CloseAll()

	inotifySignal := make(chan any)
	go fileWorker(inotifySignal)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				switch event.Op {
				case fsnotify.Write:
					inotifySignal <- true
				case fsnotify.Remove | fsnotify.Rename:
					log.Println("file removed/moved, closing")
					watcher.Close()
					return
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(params.logInputFile)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-time.After(1 * time.Second):
			if len(watcher.WatchList()) == 0 {
				watcher.Close()
				log.Println("watcher closed, exiting")
				return
			}
		case <-workerSignal:
			log.Println("closing")
			watcher.Close()
			return
		}
	}
}
