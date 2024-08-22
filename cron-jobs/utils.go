package cronjobs

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/Arinji2/sense-backend/api"
)

func getRandomLetter() string {

	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomIndex := rand.Intn(len(letters))
	return string(letters[randomIndex])
}

func getDifficultyLevel(level int) string {
	switch level {
	case 3:
		return "PHD Researchers"
	case 2:
		return "University Professors"
	default:
		return "Secondary Education Teachers"
	}
}

func difficultyAmount(difficulty int, address string, client *api.ApiClient, token string, response chan difficultyChannel, wg *sync.WaitGroup) {
	defer wg.Done()

	result, err := client.SendRequestWithQuery("GET", address, map[string]string{
		"perPage": "1",
		"filter":  fmt.Sprintf("level='%d'", difficulty)}, map[string]string{
		"AUTHORIZATION": token})

	if err != nil {
		log.Printf("error in fetching for difficulty %d: %v", difficulty, err)
		return
	}

	totalItems, ok := result["totalItems"].(float64)
	if !ok {
		log.Println("Error in parsing totalItems")
		return
	}

	difficultyData := difficultyChannel{
		totalItems: int(totalItems),
		difficulty: difficulty,
	}

	response <- difficultyData
}

func getLevel(table string, token string) difficultyChannel {

	client := api.NewApiClient()

	address := fmt.Sprintf("/api/collections/%s/records", table)

	response := make(chan difficultyChannel)
	var wg sync.WaitGroup

	levels := []int{1, 2, 3}

	for _, level := range levels {
		wg.Add(1)
		go difficultyAmount(level, address, client, token, response, &wg)
	}

	go func() {
		wg.Wait()
		close(response)
	}()

	var selectedDifficulty difficultyChannel
	var isFirst bool

	for data := range response {
		if !isFirst {
			selectedDifficulty = data
			isFirst = true
			continue
		}
		if selectedDifficulty.totalItems > data.totalItems {
			selectedDifficulty = data
		}
	}

	return selectedDifficulty
}
