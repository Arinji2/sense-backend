package cronjobs

import (
	"fmt"
	"sync"

	"github.com/Arinji2/sense-backend/api"
	"github.com/Arinji2/sense-backend/pocketbase"
)

func ResetWords() {
	token := pocketbase.PocketbaseAdminLogin()
	var wg sync.WaitGroup
	tables := []string{"fake_words", "real_words"}
	client := api.NewApiClient()

	for _, table := range tables {
		address := fmt.Sprintf("/api/collections/%s/records", table)

		levels := []int{1, 2, 3}

		for _, level := range levels {
			wg.Add(1)
			fmt.Println("Deleting words for level", level)
			go levelDeletion(level, address, client, token, &wg, table)
		}

	}

	wg.Wait()

}
