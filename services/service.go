package services

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/the-web3/eth-wallet/proto/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func startRpcServer() {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(Interceptor))
	defer grpcServer.GracefulStop()

	wallet.RegisterWalletServiceServer(grpcServer, nil)

	listen, err := net.Listen("tcp", ":"+"8989")
	if err != nil {
		log.Error("net listen failed", "err", err)
		panic(err)
	}
	reflection.Register(grpcServer)

	log.Info("savour dao start success", "port", "8989")

	if err := grpcServer.Serve(listen); err != nil {
		log.Error("grpc server serve failed", "err", err)
		panic(err)
	}

}
