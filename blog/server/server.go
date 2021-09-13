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
	"google.golang.org/protobuf/types/known/emptypb"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id,omitempty"`
	Content  string             `bson:"content,omitempty"`
	Title    string             `bson:"title,omitempty"`
}

func (blog blogItem) toBlogPb() *blogpb.Blog {
	return &blogpb.Blog{
		Id:       blog.ID.Hex(),
		AuthorId: blog.AuthorID,
		Title:    blog.Title,
		Content:  blog.Content,
	}
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
	return &blogpb.ReadBlogResponse{Blog: blog.toBlogPb()}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()

	log.Printf("Updating blog item: %v\n", blog)

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		log.Printf("Impossible to convert to ObjectId: %s", blog.GetId())
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("id='%s' Cannot parse to ObjectId.", blog.GetId()))
	}

	data := &blogItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		log.Printf("id='%s'\tImpossible to update blog item: %v", oid.Hex(), err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("id='%s' Impossible to find blog item!", err))
	}

	return &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*emptypb.Empty, error) {
	id := req.GetId()

	log.Printf("id='%s' Deleting blog item...", id)

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("id='%s' Cannot convert into ObjectId!\n", id)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Impossible to convert '%s' into ObjectId", id))
	}

	filter := bson.M{"_id": oid}

	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Fatalf("MongoDB error: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error deleting document: %v", err))
	}
	if res.DeletedCount == 0 {
		log.Printf("id='%s' No blog item found!", id)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("No blog item found with id: %v", id))
	}

	log.Printf("id='%s' Blog item deleted!", id)
	return &emptypb.Empty{}, nil
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
