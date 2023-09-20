package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/src-d/go-git.v4"
)

var (
	mediumURL string = "https://api.medium.com"
	authorID  string = os.Getenv("AUTHOR_ID")
	postPath  string = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), os.Getenv("POST_DIR"))
)

type ShortCodes struct {
	Name    string `json:"name"`
	Regex   string `json:"regex,omitempty"`
	Replace string `json:"replace,omitempty"`
}

type MediumPostPayload struct {
	Title         string   `json:"title"`
	ContentFormat string   `json:"contentFormat"`
	Content       string   `json:"content,omitempty"`
	PublishStatus string   `json:"publishStatus,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

func postToMedium(payload []byte) {
	fmt.Println("payload: ", string(payload))
	bearer := "Bearer " + os.Getenv("ACCESS_TOKEN")

	// create new request using HTTP
	req, err := http.NewRequest("POST", mediumURL+"/v1/users/"+authorID+"/posts", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("request err: %v", err)
	}

	// Add authorization header to the req
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	log.Println(req.Body)
	// Send req using http client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}
	log.Println(string([]byte(body)))
}

func getLastCommitMessage() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatal(err)
	}

	ref, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println("Last Git Commit Message: ")
	// fmt.Println(commit.Message)
	return commit.Message
}

func readJsonConfig(shortCodesFileName string) []ShortCodes {
	data, err := os.ReadFile(shortCodesFileName)
	if err != nil {
		log.Fatalf("Error while json config: %v", err)
	}

	var config []ShortCodes
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error while converting JSON: %v", err)
	}

	return config
}

func main() {
	var shortCodesFileName string
	var markdownOrHugo string
	var replaceHyperlinkToLink bool

	flag.StringVar(&markdownOrHugo, "markdownOrHugo", "markdown", "Set the flag for parsing hugo markdown or simple markdown.")
	flag.StringVar(&shortCodesFileName, "shortCodesConfigFile", "", "Pass JSON config file for parsing shortcode to markdown")
	flag.BoolVar(&replaceHyperlinkToLink, "replaceHyperlinkToLink", false, "replace markdown hyperlink syntax with just link")
	flag.Parse()

	switch markdownOrHugo {
	case "markdown":
		commitMsg := getLastCommitMessage()
		// commitMsg := "PUBLISH: test-tl.md"
		if strings.Contains(commitMsg, "PUBLISH") {
			// Extract Post Name from Commit
			postNameWithExt := strings.TrimSpace(strings.SplitAfter(commitMsg, "PUBLISH:")[1])
			postNameWithDash := strings.Split(postNameWithExt, ".")[0]
			c := cases.Title(language.Und)
			postName := c.String(strings.Join(strings.Split(postNameWithDash, "-"), " "))
			fmt.Println("post name: ", postName)

			// Extract/Get Content
			walkFunc := func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Fatalf("Error walking directory: %v", err)
					return err
				}

				fmt.Println(path)
				if strings.Contains(path, postNameWithExt) {
					content, err := os.ReadFile(path)
					if err != nil {
						log.Fatalf("Error reading file: %v", err)
					}
					payloadData := MediumPostPayload{Title: postName, ContentFormat: "markdown", Content: string(content), PublishStatus: "draft"}
					marshalData, err := json.Marshal(payloadData)
					if err != nil {
						log.Fatal(err)
					}
					postToMedium(marshalData)
					return nil
				}

				// Check if it's a directory
				// if d.IsDir() {
				// 	fmt.Println(" (directory)")
				// }

				return nil
			}

			// Start walking the directory and its subdirectories
			err := filepath.WalkDir(postPath, walkFunc)
			if err != nil {
				log.Fatalf("Error during directory traversal: %v", err)
			}

			log.Println("Post successful!")
		}
	default:
		// Hugo
		// Set environment variable for default post path: POST_PATH
		// 1. Read the blog post file - details of file name and push will be in commit like - Publish: filename.md
		// 2. Read markdown file
		// 3. Remove frontmatter part
		// 4. replace all shortcodes from post.
		// 5. call API to post on Medium
		break
	}
	config := readJsonConfig(shortCodesFileName)

	fmt.Println(config, markdownOrHugo, replaceHyperlinkToLink)
}
