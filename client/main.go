package main

import (
	"context"
	"log"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/qgymje/grpc_demo/protos/user"
	"google.golang.org/grpc"
)

// User grpc user client
type User struct {
	conn   *grpc.ClientConn
	client pb.UserClient // why not pointer?
}

// NewUser create grpc user client
func NewUser() *User {
	client, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		log.Fatal(err)
	}
	resolver := &etcdnaming.GRPCResolver{Client: client}
	balancer := grpc.RoundRobin(resolver)
	conn, err := grpc.Dial(
		"grpc_demo_user1",
		grpc.WithBalancer(balancer),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	)
	if err != nil {
		log.Fatal("user grpc server cant not connect: ", err)
	}

	u := new(User)
	u.conn = conn
	u.client = pb.NewUserClient(u.conn)

	return u
}

// Close close the client
func (u *User) Close() error {
	return u.conn.Close()
}

// SMSCode get the register code by sms
func (u *User) SMSCode(in *pb.Phone) (*pb.Code, error) {
	defer u.Close()
	return u.client.SMSCode(context.Background(), in)
}

func main() {
	user := NewUser()
	defer user.Close()
	phone := &pb.Phone{Phone: "13817782405"}
	reply, err := user.SMSCode(phone)
	if err != nil {
		log.Println(err)
	}
	log.Println(reply)
}
