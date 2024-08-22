package cronjobs

import (
	"fmt"
	"log"

	"os"

	"sync"

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

	fakeWords := make(map[int]generatedWord, 3)
	realWords := make(map[int]generatedWord, 3)

	fakeWordChannel := make(chan generatedWord, 3)
	realWordChannel := make(chan generatedWord, 3)

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			generateWord(fakeLevel.difficulty, true, accessKey, fakeWordChannel, 3)
		}(i)
	}

	for i := 0; i < 3; i++ {
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

	for i := 0; i < 3; i++ {
		word := <-fakeWordChannel
		fakeWords[i] = word
	}

	for i := 0; i < 3; i++ {
		word := <-realWordChannel
		realWords[i] = word
	}

	fmt.Println("Fake Words:", fakeWords)
	fmt.Println("Real Words:", realWords)
}
