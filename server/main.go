package main

import (
	"client_golang/prometheus"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/davecgh/go-spew/spew"
	pb "github.com/qgymje/grpc_demo/protos/user"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

var (
	port = flag.String("port", ":4000", "service port")
)

// UserServer implement pb.UserServer
type UserServer struct {
}

// SMSCode request a sms code before register
func (s *UserServer) SMSCode(ctx context.Context, in *pb.Phone) (*pb.Code, error) {
	var err error
	defer func() {
		if err != nil {
			log.Println(err)
		}
		log.Println("request: ctx:", spew.Sdump(ctx))
		log.Println("request: in:", spew.Sdump(in))
	}()

	return &pb.Code{Code: getCode()}, nil
}

func getCode() string {
	return "1234"
}

func main() {
	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		log.Fatal(err)
	}

	r := &etcdnaming.GRPCResolver{Client: cli}
	addOp := naming.Update{Op: naming.Add, Addr: "127.0.0.1:4000", Metadata: "metadata"}
	err = r.Update(context.TODO(), "grpc_demo_user1", addOp)
	if err != nil {
		log.Fatal("failed to add foo", err)
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	pb.RegisterUserServer(s, &UserServer{})

	grpc_prometheus.Register(s)

	go func() {
		http.Handle("/metrics", prometheus.Handler())
		http.ListenAndServe(":4001", nil)
	}()

	log.Println("server started at prot: ", *port)
	err = s.Serve(lis)
	if err != nil {
		log.Println("server start failed: ", err)
	}
}
