package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	pb "github.com/lyft/ratelimit/proto/ratelimit"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type descriptorValue struct {
	descriptor *pb.RateLimitDescriptor
}

func (this *descriptorValue) Set(arg string) error {
	pairs := strings.Split(arg, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return errors.New("invalid descriptor list")
		}

		this.descriptor.Entries = append(
			this.descriptor.Entries, &pb.RateLimitDescriptor_Entry{Key: parts[0], Value: parts[1]})
	}

	return nil
}

func (this *descriptorValue) String() string {
	return this.descriptor.String()
}

func main() {
	dialString := flag.String(
		"dial_string", "localhost:8081", "url of ratelimit server in <host>:<port> form")
	domain := flag.String("domain", "", "rate limit configuration domain to query")
	descriptorValue := descriptorValue{&pb.RateLimitDescriptor{}}
	flag.Var(
		&descriptorValue, "descriptors",
		"descriptor list to query in <key>=<value>,<key>=<value>,... form")
	flag.Parse()

	fmt.Printf("dial string: %s\n", *dialString)
	fmt.Printf("domain: %s\n", *domain)
	fmt.Printf("descriptors: %s\n", &descriptorValue)

	conn, err := grpc.Dial(*dialString, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("error connecting: %s\n", err.Error())
		os.Exit(1)
	}

	defer conn.Close()
	c := pb.NewRateLimitServiceClient(conn)
	response, err := c.ShouldRateLimit(
		context.Background(),
		&pb.RateLimitRequest{*domain, []*pb.RateLimitDescriptor{descriptorValue.descriptor}, 1})
	if err != nil {
		fmt.Printf("request error: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("response: %s\n", response.String())
}
