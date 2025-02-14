package main

import (
	"demo-cosebase/pkg"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"html"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

func init() {
	godotenv.Load("../../.env") // for develop
	godotenv.Load("./.env")     // for production
}

const (
	BaseURL = "https://tangthuvien.net/"
)

func main() {
	app := &cli.App{
		Name: "crawl",
		Commands: []*cli.Command{
			commandCategory(),
			commandNm(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func commandCategory() *cli.Command {
	return &cli.Command{
		Name:  "category",
		Usage: "crawl category",
		Action: func(c *cli.Context) error {
			_, err := pkg.GetDb()
			if err != nil {
				return err
			}

			htmlContent, err := pkg.FetchContent(fmt.Sprintf("%s%s", BaseURL, "tong-hop"))
			if err != nil {
				return err
			}

			content := html.UnescapeString(htmlContent)

			re, err := regexp.Compile(`<a.*?data-name="ctg".*?>([^<]+)</a>`)
			if err != nil {
				return err
			}
			matches := re.FindAllStringSubmatch(content, -1)

			// Lưu kết quả vào slice
			var categories string
			for _, match := range matches {
				//categories = append(categories, match[1]) // match[1] chứa nội dung bên trong thẻ <a>
				categories = fmt.Sprintf("%s%s", categories, match[1])
			}

			err = os.WriteFile("output.txt", []byte(categories), 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func commandNm() *cli.Command {
	return &cli.Command{
		Name:  "story-nominate",
		Usage: "crawl category",
		Action: func(c *cli.Context) error {
			_, err := pkg.GetDb()
			if err != nil {
				return err
			}

			htmlContent, err := pkg.FetchContent(fmt.Sprintf("%s%s", BaseURL, "tong-hop?rank=nm"))
			if err != nil {
				return err
			}

			content := html.UnescapeString(htmlContent)

			re, err := regexp.Compile(`(?s)<div class="book-mid-info">.*?<h4><a href="([^"]+)"`)
			if err != nil {
				return err
			}
			matches := re.FindAllStringSubmatch(content, -1)

			var wg sync.WaitGroup
			limit := make(chan struct{}, 5)

			for _, match := range matches {
				wg.Add(1)
				limit <- struct{}{}

				go func(url string) {
					defer wg.Done()
					defer func() { <-limit }()

					err := crawlStory(url)
					if err != nil {
						log.Println(err)
					}
				}(match[1])
				break
			}

			wg.Wait()
			fmt.Println("craw stories nominates successfully")
			return nil
		},
	}
}

func crawlStory(rawUrl string) error {
	encodedUrl, err := pkg.NormalizeURL(rawUrl)
	if err != nil {
		return err
	}

	//	log.Println("Crawling:", encodedUrl) // Debug

	htmlContent, err := pkg.FetchContent(encodedUrl)
	if err != nil {
		return err
	}
	unescapedContent := html.UnescapeString(htmlContent)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(unescapedContent))
	if err != nil {
		log.Fatal(err)
	}

	content, err := doc.Find("div.book-information").Html()
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(encodedUrl, "/")
	slug := parts[len(parts)-1]

	// img

	imgStoryUrlRegex, err := regexp.Compile(`(?s)class="book-img".*?<img src="(.*?)"`)
	if err != nil {
		return err
	}
	imgStoryContentMatch := imgStoryUrlRegex.FindStringSubmatch(content)
	if len(imgStoryContentMatch) == 2 {
		//	fmt.Println(slug, imgStoryContentMatch[1])
	}

	// title
	titleRegex, err := regexp.Compile(`(?s)class="book-info.*?<h1>(.*?)</h1>`)
	if err != nil {
		return err
	}
	titleContentMatch := titleRegex.FindStringSubmatch(content)

	description := doc.Find("div.book-intro").Text()

	err = os.WriteFile("output.txt", []byte(description), 0644)
	if err != nil {
		return err
	}
	if len(titleContentMatch) == 2 {
		fmt.Println(slug, titleContentMatch[1])
	}

	return nil
}

func crawlChapters(storyID int) error {
	page, limit := 0, 75
	for ; ; page++ {
		urlPageChapter := getChapterUrl(storyID, page, limit)
		resp, err := pkg.FetchContent(urlPageChapter)

	}
	return nil
}

func getChapterUrl(storyID, page, limit int) string {
	return fmt.Sprintf("https://tangthuvien.net/doc-truyen/page/%d?page=%d&limit=%d", storyID, page, limit)
}
