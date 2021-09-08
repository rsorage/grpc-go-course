package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/rsorage/grpc-go-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id,omitempty"`
	Content  string             `bson:"content,omitempty"`
	Title    string             `bson:"title,omitempty"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Printf("Create blog request: %v\n", req)
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert to OID: %v", err))
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Content:  blog.GetContent(),
			Title:    blog.Title,
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Printf("Read blog request: %v\n", req)

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		log.Printf("Impossible to convert to ObjectID: %s\n", req.GetId())
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Impossible to convert to ObjectID: %s\n", req.GetId()))
	}

	var blog blogItem
	filter := bson.M{"_id": oid}

	log.Println("Searching the DB...")
	err = collection.FindOne(context.Background(), filter).Decode(&blog)
	if err != nil {
		log.Printf("id='%s' No blog item found!\n", oid.Hex())
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("id='%s' No blog item found!", oid.Hex()))
	}

	log.Println("BlogItem retrieved from DB!")
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       blog.ID.Hex(),
			AuthorId: blog.AuthorID,
			Title:    blog.Title,
			Content:  blog.Content,
		},
	}, nil
}

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Blog Service Started!")

	client := connectMongo()
	collection = client.Database("grpc_blogs").Collection("blogs")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		log.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Ctrl+C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	log.Println("Stopping the server...")
	s.Stop()
	log.Println("Closing the listener...")
	lis.Close()
	log.Println("Closing MongoDB connection...")
	client.Disconnect(context.TODO())
	log.Println("Bye!")
}

func connectMongo() *mongo.Client {
	uri := "mongodb://localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Connecting to MongoDB server...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Erro connecting to MongoDB server: %v", err)
	}

	log.Println("Connected successfully to MongoDB server!")
	return client
}
