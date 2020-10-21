package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	basepath    string = "/home/latona/lepus/bitbucket"
	config_path string = "/home/latona/lepus/bitbucket-watchdog/watcher/config.json"
)

type MessageCommit struct {
	time       time.Time
	repository string
	branch     string
	commit_id  string
}

func Watcher() error {
	c := GetConfigInstance()
	target := c.LoadConfig(config_path)
	err := InitController(target)
	if err != nil {
		return fmt.Errorf("[watcher] failed to init (error: %v)", err)
	}
	ch := make(chan MessageCommit)
	for _, t := range target {
		for b, _ := range t.Branch {
			go CheckCommitIDController(ch, t.Repository, b)
		}
	}
	for {
		select {
		case msg := <-ch:
			path := filepath.Join(basepath, msg.repository)
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				if err := os.RemoveAll(path); err != nil {
					log.Fatalf("[watcher] faild to delete repository")
				}
			}
			if err := CloneRepository(basepath, msg.repository); err != nil {
				log.Fatalf("[watcher] failed to clone repository (error: %v)", err)
			}
			if err := CheckOut(path, msg.commit_id); err != nil {
				log.Fatalf("[watcher] failed to checkout (repository: %s, commit id: %s, error: %v)", msg.repository, msg.commit_id, err)
			}
			if err := c.UpdateCommitId(msg.repository, msg.branch, msg.commit_id, time.Now()); err != nil {
				log.Fatalf("[watcher] failed to update commit id  (repository: %s, commit id: %s)", msg.repository, msg.commit_id)
			}
			if err := WriteJson(basepath, c.targets); err != nil {
				log.Fatalf("[watcher] failed to write json file (error: %v)", err)
			}
		}
	}
}

func InitController(target []Target) error {
	for _, t := range target {
		for branch := range t.Branch {
			commit, err := GetCommitInstance().InitCommit(t.Repository, branch)
			if err != nil {
				log.Printf("[watcher] failed to get commit_id (repository: %s, branch: %s)", t.Repository, branch)
			} else {
				log.Printf("init commit branch: %v, commit: %v", branch, commit.commit_id)
				t.Branch[branch] = commit.commit_id
			}
		}
		t.Timestamp = time.Now().String()
	}
	err := WriteJson(basepath, target)
	if err != nil {
		return err
	}
	return nil
}

func CheckCommitIDController(ch chan MessageCommit, repository string, branch string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	commit, err := GetCommitInstance().InitCommit(repository, branch)
	if err != nil {
		log.Printf("[watcher] failed to get commit id (repository: %s, branch: %s)", repository, branch)
	} else {
		for {
			select {
			case t := <-ticker.C:
				res := commit.DetectCommit()
				if res == true {
					m := &MessageCommit{
						time:       t,
						repository: commit.repository,
						branch:     commit.branch,
						commit_id:  commit.commit_id,
					}
					ch <- *m
				}
			}
		}
	}
}
