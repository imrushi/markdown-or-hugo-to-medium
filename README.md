# Markdown Or Hugo to Medium

The "Markdown Or Hugo To Medium" action automates the process of pushing Hugo markdown posts or regular markdown posts to Medium. It simplifies the publishing workflow by providing options for converting and formatting your content for Medium.

## Triggering the Action

The action will push to Medium when your Git commit message contains the "PUBLISH" keyword. For example, if you want to push a Hugo or Markdown post to Medium, use a commit message like this:

- Single post: `PUBLISH: file-name.md`
- Multiple posts: `PUBLISH: file1.md, file2.md, ... fileN.md`
- Publish All posts: `PUBLISH: .` or `PUBLISH: all`

## Inputs

- **markdownOrHugo** (required)
  - Specify whether the content is in Markdown or Hugo Markdown format.
  - Default: "markdown"
- **shortcodes**

  - JSON config file location for shortcodes. The config should contain an array of objects, where each object defines a shortcode and its replacement. Config file should be present at your root directory of your project.
  - Default: config.json
  - Example shortcode config JSON:

  ```json
  [
    {
      "name": "alert",
      "regex": "\\{\\{< alert type=\"(.*?)\" >\\}\\}(.*?)\\{\\{< /alert >\\}\\}",
      "replace": "<div class=\"$1\">$2</div>"
    },
    {
      "name": "figure",
      "regex": "\\{\\{< figure src=\"(.*?)\" alt=\"(.*?)\" >\\}\\}",
      "replace": "<figure><img src=\"$1\" alt=\"$2\"></figure>"
    }
  ]
  ```

  You can also check my projects shortcode [config file](https://github.com/imrushi/imrushi.github.io/blob/main/shortcodes.json)

- **replaceHyperlinkToLink**
  - Replace hyperlinks with links for Medium cards.
  - Default: false
- **frontmatterFormat**
  - Select the frontmatter format (yaml, toml).
  - Default: "yaml"
- **draft**
  - Publish the post as a draft on Medium.
  - Default: false
- **canonicalRootUrl**
  - Canonical link for specifying original article root URL with path of folder to prioritize for search engines (use for cross-posted content). eg. https://example.com/posts/
  - Your post name url will be appended to the above root url.
    eg. postName = this-is-your-post-name , canonicalRootUrl = https://example.com/posts/
    result :- https://example.com/posts/this-is-your-post-name
  - Default: ""

## Environment Variables

The action uses the following environment variables:

- **POST_DIR**:
  -Set this variable to specify from which directory the action should take post contents.
- **ACCESS_TOKEN**:
- Set this variable to your Medium access token. You can generate an access token from your Medium settings at "Security and apps" -> "Integration token".

**Note**: Store the Access Token securely in your GitHub repository secrets and then use it in your workflow YAML

## Example Usage

```yaml
on:
  push:
    branches:
      - main

jobs:
  publish-to-medium:
    runs-on: ubuntu-latest
    env:
      POST_DIR: "path/to/your/post/directory"
      ACCESS_TOKEN: ${{ secrets.MEDIUM_ACCESS_TOKEN }}
    steps:
      - name: Checkout Code
        uses: actions/checkout@v2

      - name: Markdown Or Hugo To Medium
        uses: your_username/markdown-or-hugo-to-medium@v1
        with:
          markdownOrHugo: "hugo"
          shortcodes: "path/to/shortcodes.json"
          replaceHyperlinkToLink: false
          frontmatterFormat: "yaml"
          draft: true
          canonicalRootUrl: "https://example.com/posts" #or https://example.com
```
