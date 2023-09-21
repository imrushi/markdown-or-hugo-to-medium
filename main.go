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
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/yaml.v2"
)

var (
	mediumURL       string = "https://api.medium.com/v1/users/"
	authorID        string
	githubWorkspace string
	postDir         string
	postPath        string
	accessToken     string
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

type Frontmatter struct {
	Title string   `yaml:"title" toml:"title" json:"title"`
	Tags  []string `yaml:"tags" toml:"tags" json:"tags"`
}

// Post to medium
func postToMedium(payload []byte) {
	fmt.Println("payload: ", string(payload))
	bearer := "Bearer " + accessToken

	// create new request using HTTP
	req, err := http.NewRequest("POST", mediumURL+authorID+"/posts", bytes.NewBuffer(payload))
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

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error while reading the response bytes:", err)
		}
		log.Println(string([]byte(body))+"\n Status-Code: ", resp.StatusCode)
		return
	}
}

// Returns Last Git Commit Message
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

	return commit.Message
}

// Reads shortcodes json file
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

func extractPostName(commitMsg string) (string, string) {
	postNameWithExt := strings.TrimSpace(strings.SplitAfter(commitMsg, "PUBLISH:")[1])
	postNameWithDash := strings.Split(postNameWithExt, ".")[0]
	c := cases.Title(language.Und)
	postName := c.String(strings.Join(strings.Split(postNameWithDash, "-"), " "))
	return postNameWithExt, postName
}

func parseHeader(mdContent string) (string, string, []string) {
	yamlFrontmatterPattern := regexp.MustCompile(`(?s)---\n(.+?)\n---`)
	tomlFrontmatterPattern := regexp.MustCompile(`(?s)\+\+\+\n(.+?)\n\+\+\+`)

	var (
		frontmatterContent string
		frontmatterFormat  string
		match              []string
	)

	if yamlFrontmatterPattern.MatchString(mdContent) {
		match = yamlFrontmatterPattern.FindStringSubmatch(mdContent)
		if len(match) > 1 {
			frontmatterContent = match[1]
			frontmatterFormat = "yaml"
		}
	} else if tomlFrontmatterPattern.MatchString(mdContent) {
		match = tomlFrontmatterPattern.FindStringSubmatch(mdContent)
		if len(match) > 1 {
			frontmatterContent = match[1]
			frontmatterFormat = "toml"
		}
	}

	if frontmatterContent == "" {
		return mdContent, "", nil
	}

	// Remove the frontmatter from the Markdown content
	mdContent = strings.Replace(mdContent, match[0], "", 1)

	// Extract frontmatter fields based on the format
	var fm Frontmatter
	switch frontmatterFormat {
	case "yaml":
		err := yaml.Unmarshal([]byte(frontmatterContent), &fm)
		if err != nil {
			log.Fatal(err)
		}
	case "toml":
		err := toml.Unmarshal([]byte(frontmatterContent), &fm)
		if err != nil {
			log.Fatal(err)
		}
	}
	return mdContent, fm.Title, fm.Tags
}

func init() {

	githubWorkspace = os.Getenv("GITHUB_WORKSPACE")
	if githubWorkspace == "" {
		log.Fatalf("GITHUB_WORKSPACE environment variable is not set!")
	}

	postDir = os.Getenv("POST_DIR")
	if postDir == "" {
		log.Fatalf("POST_DIR environment variable is not set!")
	}

	authorID = os.Getenv("AUTHOR_ID")
	if authorID == "" {
		log.Fatalf("AUTHOR_ID environment variable is not set!")
	}

	accessToken = os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatalf("ACCESS_TOKEN environment variable is not set!")
	}

	postPath = filepath.Join(githubWorkspace, postDir)
}

func main() {
	var shortCodesFileName string
	var markdownOrHugo string
	var replaceHyperlinkToLink bool
	var frontMatterFormat string

	flag.StringVar(&markdownOrHugo, "markdownOrHugo", "markdown", "Set the flag for parsing hugo markdown or simple markdown. [hugo, markdown]")
	flag.StringVar(&shortCodesFileName, "shortCodesConfigFile", "", "Pass JSON config file for parsing shortcode to markdown")
	flag.StringVar(&frontMatterFormat, "frontmatter", "yaml", "select frontmatter format [yaml, toml, json]")
	flag.BoolVar(&replaceHyperlinkToLink, "replaceHyperlinkToLink", false, "replace markdown hyperlink syntax with just link")
	flag.Parse()

	commitMsg := getLastCommitMessage()
	// commitMsg := "PUBLISH: json.md"
	switch markdownOrHugo {
	case "markdown":
		if strings.Contains(commitMsg, "PUBLISH") {
			// Extract Post Name from Commit
			postNameWithExt, postName := extractPostName(commitMsg)
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
		if strings.Contains(commitMsg, "PUBLISH") {
			postNameWithExt, postName := extractPostName(commitMsg)

			// Extract/Get Content
			walkFunc := func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Fatalf("Error walking directory: %v", err)
					return err
				}

				fmt.Println(path)
				if strings.Contains(path, postNameWithExt) {
					// 2. Read markdown file
					content, err := os.ReadFile(path)
					if err != nil {
						log.Fatalf("Error reading file: %v", err)
					}
					data := string(content[:])
					// 3. Remove frontmatter part
					data, title, tags := parseHeader(data)

					fmt.Println("Title: ", title)
					fmt.Println("Tags: ", tags)

					payloadData := MediumPostPayload{Title: postName, ContentFormat: "markdown", Content: string(data), PublishStatus: "draft"}
					_, err = json.Marshal(payloadData)
					if err != nil {
						log.Fatal(err)
					}
					// postToMedium(marshalData)
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
		}

		// 4. replace all shortcodes from post.
		// 5. call API to post on Medium
	}
	config := readJsonConfig(shortCodesFileName)

	fmt.Println(config, markdownOrHugo, replaceHyperlinkToLink)
}
