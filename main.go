package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/yaml.v2"
)

var (
	mediumURL       string = "https://api.medium.com/v1/"
	authorID        string
	githubWorkspace string
	postDir         string
	postPath        string
	accessToken     string
	log             *logrus.Logger
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

// User defines a Medium user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
}

// payload defines a struct to represent payloads that are returned from Medium.
type Envelope struct {
	Data   User    `json:"data"`
	Errors []Error `json:"errors,omitempty"`
}

// Error defines an error received when making a request to the API.
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Post to medium
func postToMedium(payload []byte) int {
	// fmt.Println("payload: ", string(payload))
	bearer := fmt.Sprintf("Bearer %s", accessToken)
	client := &http.Client{}

	// create new request using HTTP
	req, err := http.NewRequest("POST", mediumURL+"users/"+authorID+"/posts", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("request err: %v", err)
	}

	// Add authorization header to the req
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	// log.Println(req.Body)

	// Send req using http client
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error on response.\n[ERROR] -", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error while reading the response bytes:", err)
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 600 {
			log.Error(string([]byte(body))+"\n Status-Code: ", resp.StatusCode)
		}
		return resp.StatusCode
	}
	return resp.StatusCode
}

// get user info from medium
func getUser() User {
	bearer := fmt.Sprintf("Bearer %s", accessToken)
	client := &http.Client{}

	getUserReq, err := http.NewRequest("GET", mediumURL+"me", nil)
	if err != nil {
		log.Fatalf("get request err: %v", err)
	}

	getUserReq.Header.Add("Authorization", bearer)

	getResp, err := client.Do(getUserReq)
	if err != nil {
		log.Error("Error on response.\n[ERROR] -", err)
	}

	defer getResp.Body.Close()

	getBody, err := io.ReadAll(getResp.Body)
	if err != nil {
		log.Error("Error while reading the response bytes:", err)
	}

	if getResp.StatusCode > 400 && getResp.StatusCode < 600 {
		log.Error(string([]byte(getBody))+"\n Status-Code: ", getResp.StatusCode)
	}

	var env Envelope
	if err := json.Unmarshal(getBody, &env); err != nil {
		log.Fatalf("Could not parse response: %s", err)
	}

	// log.Infof("get response body : %v", env.Data.ID)
	return env.Data
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

func extractPostName(commitMsg string) ([]string, []string) {
	var (
		postNameSlice        []string
		postNameWithExtSlice []string
	)

	if commitMsg == "" || !strings.Contains(commitMsg, "PUBLISH") {
		return postNameSlice, postNameWithExtSlice
	}

	f := func(c rune) bool {
		return c == ','
	}
	postNames := strings.SplitAfter(commitMsg, "PUBLISH:")[1]
	sliceOfPostName := strings.FieldsFunc(postNames, f)
	for _, val := range sliceOfPostName {
		postNameWithExt := strings.TrimSpace(val)
		postNameWithExtSlice = append(postNameWithExtSlice, postNameWithExt)
		postNameWithDash := strings.Split(postNameWithExt, ".")[0]
		c := cases.Title(language.Und)
		postName := c.String(strings.Join(strings.Split(postNameWithDash, "-"), " "))
		postNameSlice = append(postNameSlice, postName)
	}
	return postNameWithExtSlice, postNameSlice
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

func replaceShortCodes(s ShortCodes, mdContent string) string {
	shortCodeStart := regexp.MustCompile(fmt.Sprintf("{{< %s[^>]*>}}", s.Name))
	shortCodeEnd := regexp.MustCompile(fmt.Sprintf("{{< /%s >}}", s.Name))
	if s.Regex != "" {
		shortCodeStart = regexp.MustCompile(s.Regex)
	}

	replaceMdContent := shortCodeStart.ReplaceAllString(mdContent, s.Replace)
	replaceMdContent = shortCodeEnd.ReplaceAllString(replaceMdContent, s.Replace)

	return replaceMdContent
}

func init() {
	log = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		},
		Level: logrus.InfoLevel,
	}

	githubWorkspace = os.Getenv("GITHUB_WORKSPACE")
	if githubWorkspace == "" {
		log.Fatalf("GITHUB_WORKSPACE environment variable is not set!")
	}

	postDir = os.Getenv("POST_DIR")
	if postDir == "" {
		log.Fatalf("POST_DIR environment variable is not set!")
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
	var draft bool

	flag.StringVar(&markdownOrHugo, "markdown-or-hugo", "markdown", "Set the flag for parsing hugo markdown or simple markdown. [hugo, markdown]")
	flag.StringVar(&shortCodesFileName, "shortcodes-config-file", "", "Pass JSON config file for parsing shortcode to markdown")
	flag.StringVar(&frontMatterFormat, "frontmatter", "yaml", "select frontmatter format [yaml, toml, json]")
	flag.BoolVar(&replaceHyperlinkToLink, "replace-hyperlink-to-link", false, "replace markdown hyperlink syntax with just link")
	flag.BoolVar(&draft, "draft", false, "publish as a draft on medium")
	flag.Parse()

	draftPub := "draft"
	if draft {
		draftPub = "public"
	}

	user := getUser()
	authorID = user.ID
	// commitMsg := getLastCommitMessage()
	commitMsg := "PUBLISH: go-basics-and-a-dash-of-clean-code.md, lets-go.md"
	switch markdownOrHugo {
	case "markdown":
		var postRespCode int
		if strings.Contains(commitMsg, "PUBLISH") {
			// Extract Post Name from Commit
			postNameWithExt, postName := extractPostName(commitMsg)
			// fmt.Println("post name: ", postName)
			for i := 0; i < len(postName); i++ {
				// Extract/Get Content
				walkFunc := func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						log.Fatalf("Error walking directory: %v", err)
						return err
					}

					// fmt.Println(path)
					if strings.Contains(path, postNameWithExt[i]) {
						content, err := os.ReadFile(path)
						if err != nil {
							log.Fatalf("Error reading file: %v", err)
						}
						payloadData := MediumPostPayload{Title: postName[i], ContentFormat: "markdown", Content: string(content), PublishStatus: draftPub}
						marshalData, err := json.Marshal(payloadData)
						if err != nil {
							log.Fatal(err)
						}
						postRespCode = postToMedium(marshalData)
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

				if postRespCode >= 200 && postRespCode < 300 {
					log.Infof("%v: Post successful!", postName[i])
				}
			}
		}
	default:
		var postRespCode int
		// Hugo
		// Set environment variable for default post path: POST_PATH
		// 1. Read the blog post file - details of file name and push will be in commit like - Publish: filename.md
		if strings.Contains(commitMsg, "PUBLISH") {
			postNameWithExt, postName := extractPostName(commitMsg)

			for i := 0; i < len(postName); i++ {
				// Extract/Get Content
				walkFunc := func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						log.Fatalf("Error walking directory: %v", err)
						return err
					}

					// fmt.Println(path)
					if strings.Contains(path, postNameWithExt[i]) {
						// 2. Read markdown file
						content, err := os.ReadFile(path)
						if err != nil {
							log.Fatalf("Error reading file: %v", err)
						}
						data := string(content[:])
						// 3. Remove frontmatter part
						data, title, tags := parseHeader(data)

						log.Info("Title: ", title)
						log.Info("Tags: ", tags)

						if title == "" {
							title = postName[i]
						}

						// 4. replace all shortcodes from post.
						config := readJsonConfig(shortCodesFileName)

						for _, s := range config {
							data = replaceShortCodes(s, data)
						}

						payloadData := MediumPostPayload{Title: title, ContentFormat: "markdown", Content: string(data), PublishStatus: draftPub, Tags: tags}
						marshalData, err := json.Marshal(payloadData)
						if err != nil {
							log.Fatal(err)
						}
						// 5. call API to post on Medium
						postRespCode = postToMedium(marshalData)
						// fmt.Println(string(data))
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

				if postRespCode >= 200 && postRespCode < 300 {
					log.Infof("%v: Post successful!", postName[i])
				}
			}
		}
	}
}
