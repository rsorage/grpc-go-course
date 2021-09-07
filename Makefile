generate:
	protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.
	protoc calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.
	protoc blog/blogpb/blog.proto --go_out=plugins=grpc:.

srv_greet:
	go run greet/server/server.go

srv_calculator:
	go run calculator/server/server.go

srv_blog:
	go run blog/server/server.go
