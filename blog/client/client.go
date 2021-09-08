package main

import (
	"context"
	"log"

	"github.com/rsorage/grpc-go-course/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	// createBlog(c)
	readBlog(c, "6137cbfe24772434d19bc92")
}

func createBlog(c blogpb.BlogServiceClient) {
	log.Println("Creating the blog...")
	blog := blogpb.Blog{
		AuthorId: "rsorage",
		Title:    "My first blog",
		Content:  "Content of first blog",
	}

	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: &blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}
	log.Printf("Blog has been created: %v\n", res)
}

func readBlog(c blogpb.BlogServiceClient, id string) {
	log.Printf("id='%s' Reading blog...\n", id)

	req := &blogpb.ReadBlogRequest{Id: id}

	res, err := c.ReadBlog(context.Background(), req)
	if err != nil {
		log.Printf("Error retrieving blog item: %v\n", err)
		return
	}

	log.Printf("Blog item retrieved: %v\n", res)
}