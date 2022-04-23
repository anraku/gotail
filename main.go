package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	follow    bool
	lineCount int
)

func init() {
	flag.BoolVar(&follow, "f", false, "output appended data as the file grows")
	flag.IntVar(&lineCount, "n", 10, "output the last NUM lines, insted of the last 10")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("invalid args count %d", len(args))
	}
	if err := run(args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}

	lines := tail(f)
	fmt.Println(strings.Join(lines, "\n"))

	if follow {
		tailf(f)
	}

	return nil
}

func tail(r io.Reader) []string {
	scanner := bufio.NewScanner(r)
	var queue []string
	for scanner.Scan() {
		queue = append(queue, scanner.Text())
		if lineCount <= len(queue)-1 {
			queue = queue[1:]
		}
	}

	return queue
}

func tailf(r io.Reader) {
	ch := make(chan string, 1)
	// チャネルからテキストを受け取って標準出力する
	go func() {
		for {
			t, ok := <-ch
			if !ok {
				break
			}
			fmt.Println(t)
		}
	}()

	// readerからデータを読み取ってチャネルに送る
	go func() {
		for {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				ch <- scanner.Text()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)
	go func() {
		<-sigs
		done <- true
	}()
	<-done
}
