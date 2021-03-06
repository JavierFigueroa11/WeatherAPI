package utils

import (
	"TPFinal/pkg/domain"
	"encoding/json"
	"io/ioutil"
)

type Job interface {
	ExitChan() chan error
	Run(locations map[string]domain.Locations) (map[string]domain.Locations, error)
}

func ProcessJobs(jobs chan Job, db string) {
	for {
		j := <-jobs
		locations := make(map[string]domain.Locations, 0)
		content, err := ioutil.ReadFile(db)
		if err == nil {
			if err = json.Unmarshal(content, &locations); err == nil {
				locationsMod, err := j.Run(locations)

				if err == nil && locationsMod != nil {
					b, err := json.Marshal(locationsMod)
					if err == nil {
						err = ioutil.WriteFile(db, b, 0644)
					}
				}
			}
		}
		j.ExitChan() <- err
	}
}
