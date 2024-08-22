package cronjobs

import (
	"fmt"
	"log"
	"sync"

	"github.com/Arinji2/sense-backend/api"
	"github.com/Arinji2/sense-backend/pocketbase"
)

type difficultyChannel struct {
	totalItems int
	difficulty int
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

func GetLevel(table string) difficultyChannel {
	token := pocketbase.PocketbaseAdminLogin()
	client := api.NewApiClient("")

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
