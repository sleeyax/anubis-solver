package solver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
)

type Result struct {
	Hash       string
	Data       string
	Difficulty int
	Nonce      int64
}

func SolveChallengeNative(challenge *AnubisChallenge, progressCallback func(int64)) (*Result, error) {
	threads := runtime.NumCPU()

	resultChan := make(chan *Result)
	errorChan := make(chan error)
	progressChan := make(chan int64)
	done := make(chan struct{})

	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go startWorker(challenge.Challenge, challenge.Rules.Difficulty, int64(i), int64(threads), resultChan, errorChan, progressChan, done, &wg)
	}

	go func() {
		for {
			select {
			case progress := <-progressChan:
				if progressCallback != nil {
					progressCallback(progress)
				}
			case <-done:
				return
			}
		}
	}()

	var result *Result
	var err error

	select {
	case result = <-resultChan:
		close(done)
	case err = <-errorChan:
		close(done)
	}

	wg.Wait()
	return result, err
}

func startWorker(data string, difficulty int, startNonce, threads int64, resultChan chan<- *Result, errorChan chan<- error, progressChan chan<- int64, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	nonce := startNonce
	threadId := startNonce

	for {
		select {
		case <-done:
			return
		default:
			// Calculate hash
			hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d", data, nonce)))
			valid := true

			// Check if hash meets difficulty requirement
			for j := 0; j < difficulty; j++ {
				byteIndex := j / 2
				nibbleIndex := j % 2

				var nibble byte
				if nibbleIndex == 0 {
					nibble = (hash[byteIndex] >> 4) & 0x0F
				} else {
					nibble = hash[byteIndex] & 0x0F
				}

				if nibble != 0 {
					valid = false
					break
				}
			}

			if valid {
				resultChan <- &Result{
					Hash:       hex.EncodeToString(hash[:]),
					Data:       data,
					Difficulty: difficulty,
					Nonce:      nonce,
				}
				return
			}

			oldNonce := nonce
			nonce += threads

			// Send progress update
			if nonce > oldNonce|1023 && (nonce>>10)%threads == threadId {
				select {
				case progressChan <- nonce:
				default:
				}
			}
		}
	}
}
