package gapi

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient[T any](newClient func(grpc.ClientConnInterface) T, addr string) (T, func() error, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		var t T
		return t, nil, nil
	}

	client := newClient(conn)
	return client, conn.Close, nil
}
