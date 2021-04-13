package main

import (
	"fmt"
	"net"
	"log"
	"context"
	"os"
	"os/signal"
	"google.golang.org/grpc"
	"google.golang.org/status"
	"github.com/wlawt/goprojects/blog/blogpb"
	"github.com/mongodb/mongo-go-driver"
)

var collection *mongo.Collection

type server struct {}

type blogItem struct {
	ID 				objectid.ObjectID `bson:"_id,omitempty"`
	AuthorID 	string						`bson:"author_id"`
	Content 	string						`bson:"content"`
	Title 		string						`bson:"title"`
}

// CreateBlog(ctx, req) produces a CreateBlogResponse by consuming
//   a context, ctx and a request, req that contains the details
//   for the metadata, and produces an error otherwise.
func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title: blog.GetTitle(),
		Content: blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	oid, ok := res.InsertedID.(objectid.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Can't convert to OID: %v", err),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id: oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title: blog.GetTitle(),
			Content: blog.GetContent(),
		},
	}, nil
}

// ReadBlog(ctx, req) produces a ReadBlogResponse by consuming
//   a context, ctx and a request, req that contains the details
//   for the metadata, and produces an error otherwise.
func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	blogId = req.GetBlog()
	oid, err := objectid.FromHex(blogId)
	
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("can't parse ID: %v", err),
		)
	}

	// create empty struct
	data := &blogItem{}
	filter := bson.NewDocument(bson.EC.ObjectID("_id", oid)) // filter by id that matches OID
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("can't find blog with that ID: %v", err),
		)
	}

	return &blogpb.ReadBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

// dataToBlogPb(data) helper function that creates a Blog object
//   by consuming a blogItem, data.
// requires: data to be not empty
// time: O(1)
func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id: 			data.ID.Hex(),
		AuthorId: data.AuthorId,
		Content: 	data.Content,
		Title: 		data.Title,
	}
}

// UpdateBlog(ctx, req) produces a UpdateBlogResponse by consuming
//   a context, ctx and a request, req that contains the details
//   for the metadata, and produces an error otherwise.
func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()
	oid, err := objectid.FromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("can't parse ID: %v", err),
		)
	}

	// create empty struct
	data := &blogItem{}
	filter := bson.NewDocument(bson.EC.ObjectID("_id", oid))
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("can't find blog with that ID: %v", err),
		)
	}

	// update our internal struct
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("can't update object in mongo: %v", updateErr),
		)
	}

	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

// DeleteBlog(ctx, req) produces a DeleteBlogResponse by consuming
//   a context, ctx and a request, req that contains the details
//   for the metadata, and produces an error otherwise.
func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	oid, err := objectid.FromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("can't parse ID: %v", err),
		)
	}

	filter := bson.NewDocument(bson.EC.ObjectID("_id", oid))
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("can't delete object in mongo: %v", err),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("can't find blog to delete: %v", err),
		)
	}

	return &blogpb.DeleteBlogResponse{BlogId: req.GetBlogId()}, nil
} 

// ListBlog(ctx, req) produces a ListBlogResponse by consuming
//   a context, ctx and accepts a server streaming request and 
//   produces an error otherwise.
func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	// Get all blogs
	cur, err := collection.Find(context.Background(), nil)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("unknown internal error: %v", err),
		)
	}

	// when function exists, cursor will be closed
	defer cur.Close(context.Background())

	// for each new element we find, we'll decode the data
	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("unknown while decoding data from mongo: %v", err),
			)
		}

		stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlogPb(data)})
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("unknown internal error: %v", err),
		)
	}

	return nil
}

func main() {
	// if crash, we get file name and line num
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// connect to mongodb
	fmt.Println("Connecting to MongoDB...")
	client, err := mongo.NewClient("mongodb://localhost:27017")
	if err != nil { log.Fatal(err) }
	err = client.Connect(context.TODO())
	if err != nil { log.Fatal(err) }

	collection = client.Database("mydb").Collection("blog")

	// start server
	fmt.Println("Blog Service Started")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)

	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is received
	<-ch

	// Stop services
	fmt.Println("Stopping the server...")
	s.Stop()
	fmt.Println("Stopping the listener...")
	lis.Close()
	fmt.Println("Stopping MongoDB...")
	client.Disconnect(context.TODO())
	fmt.Println("Server is closed.")
}