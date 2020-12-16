package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"twofer/eid/bankid"
	"twofer/eid/freja"
	"twofer/grpc/geid"
	"twofer/grpc/gotp"
	"twofer/grpc/gqr"
	"twofer/grpc/gw6n"
	"twofer/internal/config"
	"twofer/internal/serveid"
	"twofer/internal/servotp"
	"twofer/internal/servqr"
	"twofer/internal/servw6n"

	"google.golang.org/grpc"
)

func main() {

	cfg := config.Get()

	var grpcServer *grpc.Server
	var opts []grpc.ServerOption
	grpcServer = grpc.NewServer(opts...)

	fmt.Println("Starting server")

	if cfg.OTP.Enabled {
		fmt.Println("- Enabling OTP")
		otpserv, err := servotp.New(servotp.OTPConfig{
			SkewCounter: config.Get().OTP.SkewCounter,
			SkewTime:    config.Get().OTP.SkewTime,
			RateLimit:   config.Get().OTP.RateLimit,
		}, config.Get().OTP.EncryptionKey)
		if err == nil {
			gotp.RegisterOTPServer(grpcServer, otpserv)
		} else {
			fmt.Println("Could not enable OTP", err)
		}
	}

	if cfg.QREnabled {
		fmt.Println("- Enabling QR")
		gqr.RegisterQRServer(grpcServer, servqr.New())
	}

	if cfg.EIDEnabled() {
		startEid(grpcServer)
	}

	if cfg.WebAuthn.Enabled {
		fmt.Println("- Enablin WebAuthn")
		authn, err := servw6n.New(cfg.WebAuthn)
		if err != nil {
			fmt.Println("WebAuthn", err)
		} else {
			gw6n.RegisterWebAuthnServer(grpcServer, authn)
		}
	}

	startServer(grpcServer)

}

func startServer(grpcServer *grpc.Server) {

	if config.Get().EnableGrpc {
		go func() {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Get().GRPCPort))
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			err = grpcServer.Serve(lis)
			log.Fatalf("grpc server failed: %v", err)
		}()

	}
	if config.Get().EnableHttp {
		go func() {

		}()
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)

	<-signalChannel
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		grpcServer.GracefulStop()
		wg.Done()
	}()

	wg.Wait()
}

func startEid(grpcServer *grpc.Server) {
	fmt.Println("- Enabling EID")
	serve := serveid.New()
	geid.RegisterEIDServer(grpcServer, serve)

	if config.Get().BankID.Enabled {
		fmt.Println("  - Creating BankId")
		client, err := bankid.New(bankid.ClientConfig{
			BaseURL:       config.Get().BankID.URL.String(),
			PemRootCA:     config.Get().BankID.GetRootCA(),
			PemClientCert: config.Get().BankID.GetClientCert(),
			PemClientKey:  config.Get().BankID.GetClientKey(),
		})
		if err != nil {
			fmt.Println("ERROR", err)
		}
		if err == nil {
			err = client.API().Ping()
			if err == nil {
				fmt.Println("  - Adding BankId")
				serve.Add(client)
			}
			if err != nil {
				fmt.Println("  - Err: Could not ping bankid", err)
			}
		}
	}

	if config.Get().FrejaID.Enabled {
		fmt.Println("  - Creating Freja")
		client, err := freja.New(freja.ClientConfig{
			BaseURL:       config.Get().FrejaID.URL.String(),
			PemRootCA:     config.Get().FrejaID.GetRootCA(),
			PemClientCert: config.Get().FrejaID.GetClientCert(),
			PemClientKey:  config.Get().FrejaID.GetClientKey(),
			PemJWSCert:    config.Get().FrejaID.GetJWSCert(),
		})
		if err != nil {
			fmt.Println("ERROR", err)
		}
		if err == nil {
			err = client.Ping()
			if err == nil {
				fmt.Println("  - Adding Freja")
				serve.Add(client)
			}
			if err != nil {
				fmt.Println("  - Err: Could not ping frejaid", err)
			}
		}
	}
}
