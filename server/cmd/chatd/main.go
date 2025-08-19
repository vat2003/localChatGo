package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/you/chatpoc/server/api"
	"github.com/you/chatpoc/server/internal/chat"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:7070") // đổi 0.0.0.0 nếu muốn liên máy
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	hub := chat.NewHub()
	pb.RegisterChatServiceServer(s, chat.NewService(hub))

	log.Println("chatd listening on 127.0.0.1:7070")
	log.Fatal(s.Serve(lis))
}
