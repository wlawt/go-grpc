# Go gRPC

Creating and making API calls with gRPC and protobuffers. 
This project is basically implementing a CRUD API with MongoDB 
and Docker. 

## Requirements
- go 1.14
- protobuf (https://github.com/golang/protobuf)
- gRPC (https://github.com/grpc/grpc-go)
- MongoDB 4.4
- Docker

## What is gRPC

Remote procedure calls (gRPC) uses HTTP2 to make API requests
compared to making RESTful API calls that relies on HTTP1. 

gRPC also uses protobuffers, which means that data transferred 
is done using bytes over HTTP2. Compared to RESTful, which sends
things as JSON, this offers some form of speed advantage for
demanding services. 

gRPC also creates all the boiler templates needed for the API, 
which standardizes a lot of the process, whereas in RESTful, you'd
have to implement the different CRUD functions. gRPC also supports 
multiple langauges. My gRPC service in Java can communicate with 
my other gRPC service written in Golang, etc.

One of the biggest advantages of using protobuffers and gRPC is the
support for streaming. You can have unary, server/client streaming, 
and bi-directional streaming. Unary is the simple request and
response call, identical to RESTful calls. Server streaming 
establishes requires the client to send one request to the 
server and the server can send as _many_ times back to client.
Client streaming utilizes similar logic, except vise versa.
And, bi-directional streaming combines both client and server
streaming. 

In my opinion, this addresses lots of scalability and network
problems when dealing with millions of requests per second. 
For data intensive applications, programs can establish a persistant
connection with the client/server, which will save ultimately 
save computation and energy resources.
