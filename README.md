# Hugo to Medium

This github action will publish or create draft on your Medium account.

sample running command:
`GITHUB_WORKSPACE=. POST_DIR=posts go run main.go -shortCodesConfigFile ./config.json`

````
[
    {
      "name": "code",
      "replace": "```"
    },
    {
      "name": "image",
      "regex": "{{< image src=\"https://media0.giphy.com/media/3orieQcuSiWouzdHq0/giphy.webp\" alt=\"names\" position=\"center\" style=\"border-radius: 8px; width: 320px; height: 230px;\" >}}",
      "replace": "https://media0.giphy.com/media/3orieQcuSiWouzdHq0/giphy.webp"
    }
]
````
