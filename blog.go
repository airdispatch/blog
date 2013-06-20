package blog;

import (
	"crypto/ecdsa"
	"html/template"
	"errors"
	clientFramework "airdispat.ch/client/framework"
	"airdispat.ch/airdispatch"
	"code.google.com/p/goprotobuf/proto"
	"github.com/russross/blackfriday"
	"github.com/hoisie/web"
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
	Key *ecdsa.PrivateKey

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

		proto.Unmarshal(byteTypes, dataTypes)

		toFormat := Post{}
		for _, dataObject := range(dataTypes.Payload) {
			if *dataObject.TypeName == "blog/content" {
				toFormat.plainText = string(dataObject.Payload)
			} else if *dataObject.TypeName == "blog/author" {
				toFormat.Author = string(dataObject.Payload)
			} else if *dataObject.TypeName == "blog/date" {
				toFormat.Date = string(dataObject.Payload)
			} else if *dataObject.TypeName == "blog/title" {
				toFormat.Title = string(dataObject.Payload)
			}
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
func (b *Blog) WebGoBlog(template *template.Template) WebGoRouter {
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
		template.Execute(ctx, context)
	}
}