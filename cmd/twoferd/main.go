package main

import (
	"fmt"
	"log"
	"net"
	"twofer/eid/bankid"
	"twofer/eid/freja"
	"twofer/internal/config"
	"twofer/internal/eidserver"
	"twofer/internal/otpserver"
	"twofer/internal/qrserver"
	"twofer/internal/webauthnserver"
	rpc "twofer/twoferrpc"

	"google.golang.org/grpc"
)

func main() {

	cfg := config.Get()

	fmt.Printf("%+v", cfg)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	fmt.Println("Starting server")

	if cfg.OTP.Enabled {
		fmt.Println("- Enabling OTP")
		otpserv, err := otpserver.New(otpserver.OTPConfig{
			SkewCounter: config.Get().OTP.SkewCounter,
			SkewTime:    config.Get().OTP.SkewTime,
			RateLimit:   config.Get().OTP.RateLimit,
		}, config.Get().OTP.EncryptionKey)
		if err == nil {
			rpc.RegisterOTPServer(grpcServer, otpserv)
		} else {
			fmt.Println("Could not enable OTP", err)
		}
	}

	if cfg.QREnabled {
		fmt.Println("- Enabling QR")
		rpc.RegisterQRServer(grpcServer, qrserver.New())
	}

	if cfg.EIDEnabled() {
		startEid(grpcServer)
	}

	if cfg.WebAuthn.Enabled {
		rpc.RegisterWebauthnServer(grpcServer, webauthnserver.New([]byte("this is a key")))
	}

	panic(grpcServer.Serve(lis))

}

func startEid(grpcServer *grpc.Server) {
	fmt.Println("- Enabling EID")
	serve := eidserver.New()
	rpc.RegisterEIDServer(grpcServer, serve)

	if config.Get().BankID.Enabled {
		fmt.Println(" - Enabling BankId")
		client, err := bankid.New(bankid.ClientConfig{
			BaseURL:       config.Get().BankID.URL.String(),
			PemRootCA:     config.Get().BankID.GetRootCA(),
			PemClientCert: config.Get().BankID.GetClientCert(),
			PemClientKey:  config.Get().BankID.GetClientKey(),
		})
		if err == nil {
			serve.Add(client)
		} else {
			fmt.Println("ERROR", err)
		}
	}

	if config.Get().FrejaID.Enabled {
		fmt.Println(" - Enabling Freja")
		client, err := freja.New(freja.ClientConfig{
			BaseURL:       config.Get().FrejaID.URL.String(),
			PemRootCA:     config.Get().FrejaID.GetRootCA(),
			PemClientCert: config.Get().FrejaID.GetClientCert(),
			PemClientKey:  config.Get().FrejaID.GetClientKey(),
			PemJWSCert:    config.Get().FrejaID.GetJWSCert(),
		})
		if err == nil {
			serve.Add(client)
		} else {
			fmt.Println("ERROR", err)
		}
	}
}