package blog;

import (
	"html/template"
	"errors"
	clientFramework "airdispat.ch/client/framework"
	"airdispat.ch/airdispatch"
	"airdispat.ch/common"
	"code.google.com/p/goprotobuf/proto"
	"github.com/russross/blackfriday"
	"github.com/hoisie/web"
	"time"
)

type Post struct {
	Title string
	Author string
	URL string
	Date string
	Content template.HTML
	plainText string
}

type Blog struct {
	Address string
	Trackers []string
	Key *common.ADKey

	BlogId string

	allPosts map[string]Post
}

func (b *Blog) Initialize() {
	b.allPosts = make(map[string]Post)
}

func (b *Blog) GetPost(url string) ([]Post, error) {
	thePost, ok := b.allPosts[url]
	if !ok {
		return nil, errors.New("Unable to Find Post with that URL")
	}
	return []Post{thePost}, nil
}

func (b *Blog) GetPosts() ([]Post, error) {
	c := clientFramework.Client{}
	c.Populate(b.Key)
	allPosts, err := c.DownloadPublicMail(b.Trackers, b.Address, 0)
	if err != nil {
		return nil, err
	}

	formattedPosts := []Post{}

	for _, value := range(allPosts) {
		byteTypes := value.Data
		dataTypes := &airdispatch.MailData{}

		skipped := false

		proto.Unmarshal(byteTypes, dataTypes)

		toFormat := Post{}
		for _, dataObject := range(dataTypes.Payload) {
			if *dataObject.TypeName == "airdispat.ch/blog/content" {
				toFormat.plainText = string(dataObject.Payload)
			} else if *dataObject.TypeName == "airdispat.ch/blog/author" {
				toFormat.Author = string(dataObject.Payload)
			} else if *dataObject.TypeName == "airdispat.ch/blog/title" {
				toFormat.Title = string(dataObject.Payload)
			} else if *dataObject.TypeName == "airdispat.ch/blog/id" {
				if string(dataObject.Payload) != b.BlogId {
					skipped = true
					break
				}
			}
		}

		dateObject := time.Unix(int64(*value.Timestamp), 0)
		localTZ, _ := time.LoadLocation("Local")

		toFormat.Date = dateObject.In(localTZ).Format("Jan 2, 2006 at 3:04pm")

		if toFormat.Title == "" || skipped {
			continue
		}

		formattedPosts = append(formattedPosts, b.CreatePost(toFormat))
	}

	return formattedPosts, nil
}

func (b *Blog) CreatePost(toFormat Post) Post {
	theContent := template.HTML(string(blackfriday.MarkdownCommon([]byte(toFormat.plainText))))
	thePost := Post{
		Title: toFormat.Title,
		Author: toFormat.Author, 
		URL: web.Slug(toFormat.Title, "-"),
		Date: toFormat.Date,
		Content: theContent}
	b.allPosts[thePost.URL] = thePost
	return thePost
}

type WebGoRouter func(ctx *web.Context, val string)
func (b *Blog) WebGoBlog(tmp *template.Template, templateName string) WebGoRouter {
	return func(ctx *web.Context, val string) {
		var err error
		context := make(map[string]interface{})
		if val == "/" || val == "" {
			context["Posts"], err = b.GetPosts()
		} else {
			context["Posts"], err = b.GetPost(val[1:])
		}
		if err != nil {
			ctx.Write([]byte(err.Error()))
			return
		}

		if templateName != "" {
			tmp.ExecuteTemplate(ctx, templateName, context)
			return
		}

		tmp.Execute(ctx, context)
	}
}