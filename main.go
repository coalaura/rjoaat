package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: joaat <hash>")

		return
	}

	hash := os.Args[1]

	if strings.HasPrefix(hash, "0x") {
		hash = hash[2:]
	} else if strings.HasPrefix(hash, "hash_") {
		hash = hash[5:]
	} else {
		hash = strings.ToLower(hash)
	}

	ui, err := strconv.ParseUint(hash, 16, 32)
	if err != nil {
		fmt.Println("Invalid hash")

		return
	}

	threads := uint64(runtime.NumCPU())

	// Here we go lmao
	fmt.Printf("Using %d threads.\n", threads)
	fmt.Printf("Searching for 0x%x...\n", ui)

	prefix := fmt.Sprintf("hash_%x_", ui)

	var (
		count uint64
		done  bool
		wg    sync.WaitGroup

		pre   = prehash(prefix)
		found = make(chan bool, 1)
		start = time.Now()
	)

	step := uint64(math.MaxUint64 / threads)

	for chunk := uint64(0); chunk < threads; chunk++ {
		from := chunk * step
		to := (chunk + 1) * step

		wg.Add(1)

		go func() {
			defer wg.Done()

			var str string

			for i := from; i < to; i++ {
				if done {
					break
				}

				str = strconv.FormatUint(i, 36)

				atomic.AddUint64(&count, 1)

				if joaat(pre, str) == ui {
					fmt.Printf("\nFound: %s%s\n", prefix, str)

					found <- true

					break
				}
			}
		}()
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				cnt := atomic.LoadUint64(&count)

				fmt.Printf("%s hashes checked\r", formatNumber(cnt))
			case <-found:
				done = true

				return
			}
		}
	}()

	wg.Wait()

	fmt.Printf("Completed after %s\n", time.Since(start))
}

func prehash(prefix string) uint32 {
	var hash uint32 = 0

	for i := 0; i < len(prefix); i++ {
		charCode := prefix[i]

		hash += uint32(charCode)
		hash += (hash << 10)
		hash ^= (hash >> 6)
	}

	return hash
}

func joaat(hash uint32, input string) uint64 {
	for i := 0; i < len(input); i++ {
		charCode := input[i]

		hash += uint32(charCode)
		hash += (hash << 10)
		hash ^= (hash >> 6)
	}

	hash += (hash << 3)
	hash ^= (hash >> 11)
	hash += (hash << 15)

	// Convert uint32 to uint64 while retaining the bit pattern.
	return uint64(hash)
}

func formatNumber(num uint64) string {
	p := message.NewPrinter(language.English)

	return p.Sprintf("%d", num)
}
