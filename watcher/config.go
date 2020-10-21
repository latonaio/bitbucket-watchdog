package watcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

var sharedConfigInstance *Target

type Target struct {
	Repository string            `json:"repository"`
	Branch     map[string]string `json:"branch"`
	Timestamp  string            `json:"timestamp"`
	targets    []Target
}

func GetConfigInstance() *Target {
	if sharedConfigInstance == nil {
		sharedConfigInstance = &Target{}
	}
	return sharedConfigInstance
}

// Read target
func (t *Target) LoadConfig(path string) []Target {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	if err := json.Unmarshal(bytes, &t.targets); err != nil {
		log.Fatal(err)
		return nil
	}
	return t.targets
}

// Update config Json
func (t *Target) UpdateCommitId(repository string, branch string, commitid string, time time.Time) error {
	for _, i := range t.targets {
		if i.Repository == repository {
			i.Branch[branch] = commitid
			i.Timestamp = time.String()
			return nil
		}
	}
	return fmt.Errorf("repository or branch is invalid (repositry: %s, branch: %s)", repository, branch)
}

func WriteJson(path string, data []Target) error {
	file, _ := json.MarshalIndent(data, "", "")
	output_path := filepath.Join(path, "config.json")
	if err := ioutil.WriteFile(output_path, file, 0644); err != nil {
		return err
	}
	return nil
}
