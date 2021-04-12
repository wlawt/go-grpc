package main


import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/status"
	"github.com/wlawt/goprojects/blog/blogpb"
	"log"
)

func main() {
	fmt.Println("Blog Client.")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("couldn't connect: %v", err)
	}

	// defer this command to the very end and run it
	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	// create blog
	fmt.Println("Creating the blog")
	blog := &blogpb.Blog{
		AuthorId: "Will",
		Title: "Hello World",
		Content: "This is working",
	}
	createBlogRes, err := c.CreateBlog(context.Background(), in *blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
		return
	}
	fmt.Printf("Blog has been created: %v", createBlogRes)
	blogID := createBlogRes.GetBlog().GetId()

	// read blog
	fmt.Println("Reading the blog")
	/* A bad request */
	_, readErr := c.ReadBlog(context.Background(), in *blogpb.ReadBlogRequest{BlogId: "125sda589cweFewr9"})
	if readErr != nil {
		log.Fatalf("Error happened while reading: %v \n", readErr)
	}
	
	/* A better request */
	readBlogReq := &blogpb.ReadBlogRequest{BlogId: blogID}
	readBlogRes, readBlogErr := c.ReadBlog(context.Background(), readBlogReq)
	if readBlogErr != nil {
		log.Fatalf("Error happened while reading: %v \n", readBlogErr)
	}
	fmt.Printf("Blog was read: %v \n", readBlogRes)


	// update blog
	newBlog := &blogpb.Blog{
		Id: 			blogID,
		AuthorId: "James",
		Title: 		"What's up",
		Content: 	"Content has been changed",
	}

	updateRes, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newBlog})
	if updateErr != nil {
		log.Fatalf("Error happened while updating: %v \n", updateErr)
	}
	fmt.Printf("Blog was updated: %v \n", updateRes)


	// delete blog
	deleteRes, deleteErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogID})
	if deleteErr != nil {
		log.Fatalf("Error happened while deleting: %v \n", deleteRes)
	}
	fmt.Printf("Blog was deleted: %v \n", deleteRes)


	// list blogs
	stream, err := c.ListBlog(context.Background(), &blog.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Error happened while listing blogs: %v \n", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("something bad happened: %v", err)
		}
		fmt.Println(res.GetBlog())
	}
}