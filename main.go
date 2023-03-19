package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type js struct {
	Commit comInfo `json:"commit"`
}
type comInfo struct {
	Author  authorInfo `json:"author"`
	Message string     `json:"message"`
}

type authorInfo struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

func (j *js) parseDate() {
	var d *string
	d = &j.Commit.Author.Date
	splDate := strings.Split(*d, "T")
	*d = splDate[0]
}

func (j *js) parseMessage() {
	j.Commit.Message = strings.ReplaceAll(j.Commit.Message, "\n\n", " ")
}

func (j *js) WriteToFile(file *os.File) {
	j.parseDate()
	j.parseMessage()
	dataToWrite := "message:" + j.Commit.Message + "\n" + "author:" + j.Commit.Author.Name + "\n" + "date:" + j.Commit.Author.Date + "\n\n"
	if _, err := file.WriteString(dataToWrite); err != nil {
		panic(err)
	}
}

func dataFromURL(url string) ([]byte, error) {
	url = strings.ReplaceAll(url, "{/sha}", "")
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func findRepo(targetName string, reposList []*github.Repository) (int, error) {
	for ind, r := range reposList {
		if *r.Name == targetName {
			return ind, nil
		}
	}
	return 0, errors.New("can't find such repository")
}
func main() {
	fileName := "info.txt"
	filePath := ".\\" + fileName
	var (
		data     []byte
		j        []js
		token    string
		user     string
		repoName string
	)
	fmt.Println("Write down access token for your github account:")
	fmt.Scan(&token)
	fmt.Println("Write down username of your account:")
	fmt.Scan(&user)
	fmt.Println("Write down repository name:")
	fmt.Scan(&repoName)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	repos, _, err := client.Repositories.List(ctx, user, nil)
	if err != nil {
		panic(err)
	}

	tgIndex, err := findRepo(repoName, repos)
	if err != nil {
		panic(err)
	}
	commitsURL := repos[tgIndex].GetCommitsURL()
	data, err = dataFromURL(commitsURL)

	if err != nil {
		log.Fatal(err)
	}

	if err = json.Unmarshal(data, &j); err != nil {
		log.Fatal(err)
	}

	_, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		os.Remove(filePath)
	}
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Errorf("can't create file %v", err)
	}
	for _, v := range j {
		v.WriteToFile(file)
	}

	defer file.Close()
}
