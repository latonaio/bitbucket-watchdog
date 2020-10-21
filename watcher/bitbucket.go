package watcher

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	api_url          = "https://api.bitbucket.org/2.0/repositories/latonaio"
	token_url        = "https://bitbucket.org/site/oauth2/access_token"
	app_id           = "PBPbhJqBeBZzjWyqEW"
	secret           = "V8gjrujZFe9wFQsRrWyGQVJqJGXn7bWM"
	get_commit_query = "fields=values.hash,values.date"
)

var sharedCommitInstance *Commit

type Commit struct {
	repository string
	branch     string
	commit_id  string
}

func GetCommitInstance() *Commit {
	if sharedCommitInstance == nil {
		sharedCommitInstance = &Commit{}
	}
	return sharedCommitInstance
}

func CloneRepository(directory string, repository string) error {
	oauth_client, err := GetOauthInstance().NewOAuthClientCredentials(token_url, api_url, app_id, secret)
	if err != nil {
		return fmt.Errorf("[bitbucket]: failed to authenticate to bitbucket server %v", err)
	}
	dir := filepath.Join(directory, repository)
	access_token := oauth_client.token.AccessToken
	git_url := "https://x-token-auth:" + access_token + "@bitbucket.org/latonaio/" + repository + ".git"
	if _, err = git.PlainClone(
		dir, false,
		&git.CloneOptions{
			URL:           git_url,
			ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		}); err != nil {
		log.Printf("[bitbucket] failed to clone repo (url: %v, error: %v)", git_url, err)
	}
	log.Printf("[bitbucket] success to clone repository(directory: %s, repository: %s)", directory, repository)
	return nil
}

func CheckOut(path string, commit string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("[bitbucket]: failed to authenticate to bitbucket server %v", err)
	}
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("[bitbucket]: failed to get the working directory for the repository %v", err)
	}
	if err := w.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(commit)}); err != nil {
		return fmt.Errorf("[bitbucket]: failed to checkout (commit id: %v, err: %v)", commit, err)
	}
	log.Printf("[bitbucket] success to checkout (id: %s)", commit)
	return nil
}

func GetCommits(repository string, branch string) (*CommitResponse, error) {
	oauth_client, err := GetOauthInstance().NewOAuthClientCredentials(token_url, api_url, app_id, secret)
	if err != nil {
		return nil, fmt.Errorf("[bitbucket]: failed to authenticate to bitbucket server %v", err)
	}
	url := oauth_client.GenerateRequestURL("/" + repository + "/commits/" + branch + "?" + get_commit_query)
	res, err := oauth_client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("[bitbucket]: failed to get commits err: %v", err)
	}
	return res, nil
}

func (c *Commit) InitCommit(repository string, branch string) (*Commit, error) {
	log.Printf("[bitbucket] init commit (repository: %v, branch: %v)", repository, branch)
	res, err := GetCommits(repository, branch)
	if err != nil {
		return nil, fmt.Errorf("[bitbucket]: failed to new check commit id conntroller err: %v", err)
	}
	commit := &Commit{
		repository: repository,
		branch:     branch,
		commit_id:  res.Values[0].Hash,
	}
	return commit, nil
}

func (c *Commit) DetectCommit() bool {
	res, err := GetCommits(c.repository, c.branch)
	if err != nil {
		return false
	}
	if res.Values[0].Hash != c.commit_id {
		c.commit_id = res.Values[0].Hash
		log.Printf("[bitbucket] detect new commit (repository: %s, branch: %s, commit_id; %s)", c.repository, c.branch, c.commit_id)
		return true
	} else {
		return false
	}
}
