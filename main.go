package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const FPL_BASE_URL = "https://fantasy.premierleague.com/api/bootstrap-static/"

type FplEvent struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	DeadlineTime           string `json:"deadline_time"`
	Finished               bool   `json:"finished"`
	DeadlineTimeEpoch      int64  `json:"deadline_time_epoch"`
	DeadlineTimeGameOffset int    `json:"deadline_time_game_offset"`
	IsPrevious             bool   `json:"is_previous"`
	IsCurrent              bool   `json:"is_current"`
	IsNext                 bool   `json:"is_next"`
}

type FplResponse struct {
	Events []FplEvent `json:"events"`
}

func main() {
	resp, err := http.Get(FPL_BASE_URL)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	var result FplResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalln(err)
	}
	currentTime := time.Now().Unix()
	nextEvent, err := getNextEvent(result.Events)

	if err != nil {
		log.Fatalln(err.Error())
	}

	currentEvent, err := getCurrentEvent(result.Events)

	if err != nil {
		log.Println(err.Error())
		timeUntilDeadline, err := getNextDeadline(nextEvent, currentTime)

		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Time until deadline: %s\n", timeUntilDeadline)
		os.Exit(0)
	}

	ceTimeUntilDeadline, err := getCurrentEventDeadline(currentEvent, currentTime)

	if err != nil {
		log.Println(err.Error())
		timeUntilDeadline, err := getNextDeadline(nextEvent, currentTime)

		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Time until deadline: %s\n", timeUntilDeadline)
		os.Exit(0)
	}

	fmt.Printf("Time until deadline: %s\n", ceTimeUntilDeadline)

	sendSms(fmt.Sprintf("Here is your FPL Deadline reminder!\n Time until deadline: %s\n", ceTimeUntilDeadline))

}

func getCurrentEvent(events []FplEvent) (FplEvent, error) {
	for _, event := range events {
		if event.IsCurrent {
			return event, nil
		}
	}

	return FplEvent{}, errors.New("No current event found")
}

func getNextEvent(events []FplEvent) (FplEvent, error) {
	for _, event := range events {
		if event.IsNext {
			return event, nil
		}
	}

	return FplEvent{}, errors.New("No next event found")
}

func getNextDeadline(event FplEvent, currentTime int64) (time.Duration, error) {
	if event.Finished == false && event.DeadlineTimeEpoch > currentTime {
		deadlineTime := time.Unix(event.DeadlineTimeEpoch, 0)
		timeUntilDeadline := time.Until(deadlineTime)

		return timeUntilDeadline, nil
	}

	return 0, errors.New("Could not get deadline information")
}

func getCurrentEventDeadline(event FplEvent, currentTime int64) (time.Duration, error) {
	if event.Finished == false && event.DeadlineTimeEpoch > currentTime {
		deadlineTime := time.Unix(event.DeadlineTimeEpoch, 0)
		timeUntilDeadline := time.Until(deadlineTime)

		return timeUntilDeadline, nil
	}

	return 0, errors.New("Could not get deadline information")
}
