package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type data struct {
	logdetect bool
	strnum    int
	startses  int
	timeses   int
}

func main() {

	if len(os.Args) != 3 {
		fmt.Println("Using: solution.exe вход.txt выход.txt")
		os.Exit(1)
	}

	entpath := os.Args[1]
	exitpath := os.Args[2]

	entfile, err := os.Open(entpath)
	if err != nil {
		fmt.Println("open ERROR!!")
		log.Fatal(err)
	}
	defer entfile.Close()

	scanner := bufio.NewScanner(entfile)
	scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)

	users := make(map[string]data)
	userStack := []string{}

	begindetect := false

	var savedUsers map[string]data
	var savedUserStack []string

	res, err := os.Create(exitpath)
	if err != nil {
		fmt.Println("open ERROR!!")
		log.Fatal(err)
	}
	defer res.Close()

	saveState := func() {
		savedUsers = make(map[string]data)
		for guest, account := range users {
			savedUsers[guest] = account
		}

		savedUserStack = append([]string(nil), userStack...)

	}

	rollback := func() {
		users = savedUsers
		userStack = savedUserStack

		begindetect = false
	}

	commit := func() {
		begindetect = false
	}

	count := 0
	sescounter := 0
	transnum := 0
	peaksession := 0
	parsescounter := 0
	avgsession := 0
	maxparsessions := 0

	start := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		count++
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		} else if len(words) > 3 || len(words) == 2 {
			fmt.Fprintf(res, "ERROR %d bad_event\n", count)
			continue
		}

		command := words[0]

		if len(words) == 1 {
			if command == "BEGIN" {
				if begindetect {
					fmt.Fprintf(res, "ERROR %d nested_begin\n", count)
					continue
				}

				transnum = count
				begindetect = true
				saveState()

			} else if command == "COMMIT" {
				if !begindetect {
					fmt.Fprintf(res, "ERROR %d no_transaction\n", count)
					continue
				}

				commit()
			} else if command == "ROLLBACK" {
				if !begindetect {
					fmt.Fprintf(res, "ERROR %d no_transaction\n", count)
					continue
				}

				rollback()
			} else {
				fmt.Fprintf(res, "ERROR %d bad_event\n", count)
				continue
			}
		} else if len(words) == 3 {

			guest := words[1]
			timestr := words[2]
			time, err := strconv.Atoi(timestr)
			if err != nil {
				fmt.Println("TIME ERROR:", err)
				return
			}

			if command != "LOGIN" && command != "LOGOUT" {
				fmt.Fprintf(res, "ERROR %d bad_event\n", count)
				continue
			}

			account, exist := users[guest]

			if command == "LOGIN" {
				if account.logdetect {
					fmt.Fprintf(res, "ERROR %d re-login\n", count)
					continue
				}

				account.startses = time

				account.logdetect = true
				account.strnum = count
				users[guest] = account

				userStack = append(userStack, guest)

				parsescounter++

			} else if command == "LOGOUT" {
				if !exist || !account.logdetect {
					fmt.Fprintf(res, "ERROR %d logout_without_login\n", count)
					continue
				}

				account.logdetect = false
				account.strnum = count

				account.timeses = time - account.startses

				users[guest] = account

				userStack = userStack[:len(userStack)-1]

				sescounter++

				parsescounter--
			}
		}
		if parsescounter > maxparsessions {
			maxparsessions = parsescounter
			peaksession = count
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("read file ERROR!!")
		log.Fatal(err)
	}

	if begindetect {
		fmt.Fprintf(res, "ERROR %d transaction_not_closed\n", transnum)
	}

	if len(userStack) > 0 {
		for _, user := range userStack {
			fmt.Fprintf(res, "ERROR %d session_not_closed\n", users[user].strnum)
		}
	}

	fmt.Fprintf(res, "SESSIONS %d\n", sescounter)

	for _, user := range users {
		avgsession += user.timeses
	}

	if sescounter != 0 {
		avgsession = avgsession / sescounter
	}

	elapsed := time.Since(start)

	fmt.Fprintf(os.Stderr, "time: %.2f ms\n", float64(elapsed.Microseconds())/1000)

	fmt.Printf("Пиковое значение активных сессий: %d\n", maxparsessions)
	fmt.Printf("Момент достижения пика: %d\n", peaksession)
	fmt.Printf("Среднее время длительности сессий: %d\n", avgsession)

}
