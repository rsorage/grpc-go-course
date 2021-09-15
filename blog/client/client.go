package main

import (
	"context"
	"io"
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

	createBlog(c)
	// readBlog(c, "6137cbfe24772434d19bc92")
	// updateBlog(c)
	// deleteBlog(c)
	listBlogs(c)
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

func updateBlog(c blogpb.BlogServiceClient) {
	id := "6137d413ca5e9c29f1c44df5"

	log.Printf("id='%s' Reading blog...\n", id)

	data := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       id,
			AuthorId: "rsorage",
			Title:    "My updated blog title",
			Content:  "Updated content...",
		},
	}

	req, err := c.UpdateBlog(context.Background(), data)
	if err != nil {
		log.Printf("Error updating blog item: %v\n", err)
		return
	}

	log.Printf("id='%s' Blog item updated: %v", id, req)
}

func deleteBlog(c blogpb.BlogServiceClient) {
	id := "6137cbfe24772434d19bc92b"

	log.Printf("id='%s' Deleting blog item...\n", id)

	_, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{Id: id})
	if err != nil {
		log.Fatalf("Error deleting blog item: %v", err)
		return
	}

	log.Println("Blog item deleted!")
}

func listBlogs(c blogpb.BlogServiceClient) {
	log.Println("Listing blog items...")

	stream, err := c.ListBlog(context.Background(), &blogpb.Pageable{})
	if err != nil {
		log.Fatalf("Error opening stream: %v", err)
		return
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Server closed stream!")
			break
		}
		if err != nil {
			log.Fatalf("Error receiving blog items: %v", err)
			break
		}

		log.Printf("Receiving blog item: %v\n", res.Blog)
	}

}
