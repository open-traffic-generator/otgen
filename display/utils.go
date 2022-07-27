package display

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// TODO: would be nice to make it configurable or automatically determined
const NAME_FIELD string = "name"

type DataPoint map[string]interface{}

type DataProcessor interface {
	Layout([]DataPoint) error
	Process([]DataPoint) error
}

func DataProcessorStart(dp DataProcessor) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	needLayout := true

	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			text := scanner.Text()

			var input []DataPoint
			err := json.Unmarshal([]byte(text), &input)
			if err != nil {
				log.Fatal(err)
			}

			if needLayout {
				err := dp.Layout(input)
				if err != nil {
					log.Fatal(err)
				}
				needLayout = false
			}
			err = dp.Process(input)
			if err != nil {
				log.Fatal(err)
			}
		}

		wg.Done()
	}()

	return &wg
}
