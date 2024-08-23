package cronjobs

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

type generatedWord struct {
	word       string
	definition string
	isFake     bool
	level      int
}

func InsertWords() {

	fmt.Println("STARTING WORDS INSERTION")
	token := pocketbase.PocketbaseAdminLogin()
	fakeLevel := getLevel("fake_words", token)
	realLevel := getLevel("real_words", token)

	godotenv.Load()
	accessKey := os.Getenv("ACCESS_KEY")
	wordsToGenerate := os.Getenv("WORDS_TO_GENERATE")
	if wordsToGenerate == "" {
		wordsToGenerate = "6"
	}

	wordsToGenerateInt, err := strconv.Atoi(wordsToGenerate)
	if err != nil {
		log.Fatal("Error converting WORDS_TO_GENERATE to int")
	}

	fakeWords := []generatedWord{}
	realWords := []generatedWord{}

	fakeWordChannel := make(chan generatedWord, wordsToGenerateInt)
	realWordChannel := make(chan generatedWord, wordsToGenerateInt)

	var wg sync.WaitGroup

	for i := 0; i < wordsToGenerateInt; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			generateWord(fakeLevel.difficulty, true, accessKey, fakeWordChannel, 3)
		}(i)
	}

	for i := 0; i < wordsToGenerateInt; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			generateWord(realLevel.difficulty, false, accessKey, realWordChannel, 3)
		}(i)
	}

	go func() {
		wg.Wait()
		close(fakeWordChannel)
		close(realWordChannel)
	}()

	for word := range fakeWordChannel {
		fakeWords = append(fakeWords, word)
	}

	for word := range realWordChannel {
		realWords = append(realWords, word)
	}

	fakeWords = filterWords(fakeWords)
	realWords = filterWords(realWords)

	fmt.Println("WORDS GENERATED")

	processWords := func(data []generatedWord) {
		for _, wordData := range data {
			wg.Add(1)
			go func(wordData generatedWord) {
				defer wg.Done()

				var tableName string

				if wordData.isFake {
					tableName = "fake_words"
				} else {
					tableName = "real_words"
				}

				exists := wordCheck(wordData, token, tableName)
				fmt.Println(wordData.word, "exists", exists)
				if exists {
					return
				}

				client := api.NewApiClient()
				_, err := client.SendRequestWithBody("POST", fmt.Sprintf("/api/collections/%s/records", tableName), map[string]string{
					"word":       strings.ToLower(wordData.word),
					"definition": strings.ToLower(wordData.definition),
					"level":      strconv.Itoa(wordData.level),
				}, map[string]string{
					"Content-Type":  "application/json",
					"Authorization": token,
				})

				if err != nil {
					fmt.Println("Error posting wordData", wordData.word)
				}
			}(wordData)
		}
	}

	processWords(fakeWords)
	processWords(realWords)

	wg.Wait()

	fmt.Printf("WORDS INSERTED SUCCESSFULLY. %d fake words and %d real words\n", len(fakeWords), len(realWords))
}
