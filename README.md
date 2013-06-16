# The Airdispatch Blogging Project

This github project aims to show off the extensibility of the Airdispatch protocol by using it to create a blog. Any public messages sent by the specified address are converted into posts using this code that is less than 200 lines long.

## Usage

You must first import the code:

    import github.com/airdispatch/blog

We currently provide bindings for [web.go](https://github.com/hoisie/web), the following example is intended to be used in a web.go project.

    // Create the Blog Instance
    theBlog =  &blog.Blog{
    	Address: "e7da159a65cb19a37c86b56f789e96c410a6a5b74a8a570f",   // The airdispatch address to display
		Trackers: []string{"localhost:1024"},                          // The tracker used for that address
		Key: serverKey,                                                // The key that the server uses to authenticate
	}
    
    // Initialize It
    theBlog.Initialize()
    
    // Use it in your routes
    server.Get("/blog(.*)", theBlog.WebGoBlog(&blogTemplate))
    
##### Template Information

The template that you pass to the blog must contain something like this to display the posts:

    {{ range .Posts }}
      <div class="row blogpost">
        <div class="span7 offset1">
          <h3><a href="/blog/{{.URL}}">{{.Title}}</a></h3>
          <h5>by {{.Author}} on {{.Date}}</h5>
          <div class="body">
            {{.Content}}
          </div>
        </div>
      </div>
      <hr/>
    {{ else }}
      <h2>No posts here.</h2>
    {{ end }}
    
That's it! It automatically creates slugs for all of your posts, and responds to those specific pages.
