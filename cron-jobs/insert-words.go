package cronjobs

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/Arinji2/sense-backend/api"
	"github.com/Arinji2/sense-backend/pocketbase"
	"github.com/joho/godotenv"
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

type generatedWord struct {
	word       string
	definition string
	isFake     bool
}

func generateWord(level int, fake bool, accessKey string, response chan<- generatedWord, retries int) {

	client := api.NewApiClient("https://ai.arinji.com")

	var content string
	if fake {
		content = fmt.Sprintf(
			"Take a word from the english language, modify it in a way so that the new word is a fake made-up word but sounds like a real word. Modify the definition of the old word to make it match the new word and compress it to a maximum of 6 words. Respond with the new word and the definition in a line separated by a semicolon. The amount of modifications must be of a level of difficulty suitable for %s. Seed: %.5f",
			getDifficultyLevel(level),
			rand.Float64(),
		)
	} else {
		content = fmt.Sprintf(
			"Generate a random word with its definition from the English Dictionary. Edit the definition of the word by compressing it to a maximum of 6 words. Respond with the word and the definition in a line separated by a semicolon. The word must be of a level of difficulty suitable for %s. Seed: %.5f",
			getDifficultyLevel(level),
			rand.Float64(),
		)
	}

	body := []map[string]string{
		{
			"role":    "user",
			"content": content,
		},
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessKey,
	}

	res, err := client.SendRequestWithBody("POST", "/completions", body, headers)
	if err != nil {

		if retries > 0 {
			generateWord(level, fake, accessKey, response, retries-1)
		} else {

			response <- generatedWord{word: "FAIL", definition: "FAIL", isFake: fake}
		}
		return
	}

	message, ok := res["message"].(map[string]interface{})
	if !ok {
		if retries > 0 {
			generateWord(level, fake, accessKey, response, retries-1)
		} else {
			response <- generatedWord{word: "FAIL", definition: "FAIL", isFake: fake}
		}
		return
	}

	data, ok := message["content"].(string)
	if !ok {
		if retries > 0 {
			generateWord(level, fake, accessKey, response, retries-1)
		} else {
			response <- generatedWord{word: "FAIL", definition: "FAIL", isFake: fake}
		}
		return
	}

	parts := strings.SplitN(data, ";", 2)
	if len(parts) < 2 {
		if retries > 0 {
			generateWord(level, fake, accessKey, response, retries-1)
		} else {
			response <- generatedWord{word: "FAIL", definition: "FAIL", isFake: fake}
		}
		return
	}

	word := strings.TrimSpace(parts[0])
	definition := strings.TrimSpace(parts[1])

	response <- generatedWord{
		word:       word,
		definition: definition,
		isFake:     fake,
	}
}

func InsertWords() {

	token := pocketbase.PocketbaseAdminLogin()
	fakeLevel := getLevel("fake_words", token)
	realLevel := getLevel("real_words", token)

	fmt.Println(token, fakeLevel, realLevel)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accessKey := os.Getenv("ACCESS_KEY")

	fakeWords := make(map[int]generatedWord, 6)
	realWords := make(map[int]generatedWord, 6)

	fakeWordChannel := make(chan generatedWord, 6)
	realWordChannel := make(chan generatedWord, 6)

	var wg sync.WaitGroup

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			generateWord(fakeLevel.difficulty, true, accessKey, fakeWordChannel, 3)
		}(i)
	}

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			generateWord(realLevel.difficulty, false, accessKey, realWordChannel, 3)
		}(i)
	}

	go func() {
		for i := 0; i < 6; i++ {
			word := <-fakeWordChannel
			fakeWords[i] = word
		}
		close(fakeWordChannel)
	}()

	go func() {
		for i := 0; i < 6; i++ {
			word := <-realWordChannel
			realWords[i] = word
		}
		close(realWordChannel)
	}()

	wg.Wait()

	fmt.Println("Fake Words:", fakeWords)
	fmt.Println("Real Words:", realWords)
}
