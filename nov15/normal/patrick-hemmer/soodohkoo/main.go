package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
)

var difficulties = map[string]int{
	"easy":   45,
	"medium": 50,
	"hard":   55,
	"insane": 60,
}

func main() {
	os.Exit(mainMain())
}
func mainMain() int {
	mode := flag.String("mode", "solve", "Operation mode {solve|solveStream|generate}")
	difficulty := flag.String("difficulty", "medium", "Difficulty of generated board {easy|medium|hard|insane|1-70}")
	showStats := flag.Bool("stats", false, "show solver statistics")
	flag.Parse()

	var err error
	switch *mode {
	case "solve":
		err = mainSolveOne(*showStats)
	case "solveStream":
		err = mainSolveStream(*showStats)
	case "generate":
		err = mainGenerate(*difficulty)
	default:
		flag.Usage()
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	return 0
}

func mainSolveReader(input io.Reader, showStats bool) ([]byte, error) {
	b := NewBoard()
	_, err := b.ReadFrom(input)
	if err != nil {
		return nil, err
	}

	if !b.Solve() {
		return nil, fmt.Errorf("invalid board: no solution")
	}

	buf := bytes.NewBuffer(nil)
	buf.Write(b.Art())

	if showStats {
		fmt.Fprintf(buf, "Stats:\n")
		fmt.Fprintf(buf, "  %-30s %8s %8s %14s\n", "Algorithm", "Calls", "Changes", "Duration (ns)")
		for _, a := range b.Algorithms {
			stats := a.Stats()
			fmt.Fprintf(buf, "  %-30s %8d %8d %14d\n", a.Name(), stats.Calls, stats.Changes, stats.Duration)
		}
		stats := b.guessStats
		fmt.Fprintf(buf, "  %-30s %8d %8d %14d\n", "guesser", stats.Calls, stats.Changes, stats.Duration)
	}

	return buf.Bytes(), nil
}

func mainSolveOne(showStats bool) error {
	out, err := mainSolveReader(os.Stdin, showStats)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}

type streamJob struct {
	bs  []byte
	err error
	wg  sync.WaitGroup
}

func mainSolveStream(showStats bool) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	workerCount := runtime.GOMAXPROCS(-1)
	errChan := make(chan error, workerCount)

	workerJobs := make(chan *streamJob, workerCount*4)
	defer close(workerJobs)
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			for streamJob := range workerJobs {
				buf := bytes.NewBuffer(streamJob.bs)
				streamJob.bs, streamJob.err = mainSolveReader(buf, showStats)
				streamJob.wg.Done()
			}
			wg.Done()
		}()
	}

	publisherJobs := make(chan *streamJob, workerCount*8)
	defer close(publisherJobs)
	wg.Add(1)
	go func() {
		for streamJob := range publisherJobs {
			streamJob.wg.Wait()
			if streamJob.err != nil {
				errChan <- streamJob.err
				break
			}
			if _, err := os.Stdout.Write(streamJob.bs); err != nil {
				errChan <- err
				break
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for {
			job := &streamJob{
				bs: make([]byte, 9*9*2),
			}
			_, err := io.ReadFull(os.Stdin, job.bs)
			if err == io.EOF {
				errChan <- nil
				break
			}
			if err != nil {
				errChan <- err
				break
			}

			job.wg.Add(1)
			workerJobs <- job
			publisherJobs <- job
		}
		wg.Done()
	}()
	defer os.Stdin.Close() // unblock the io.ReadFull and get the goroutine to shut down

	return <-errChan
}

func mainGenerate(difficulty string) error {
	lvl := difficulties[difficulty]
	if lvl == 0 {
		// try and parse as an int.
		var err error
		lvl, err = strconv.Atoi(difficulty)
		if err != nil || lvl < 0 {
			return fmt.Errorf("invalid difficulty level")
		}
	}

	b := NewRandomBoard(lvl)
	fmt.Printf("%s", b.Art())
	return nil
}
