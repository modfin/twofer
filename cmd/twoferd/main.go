package main

import (
	"fmt"
	"log"
	"net"
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

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

	panic(grpcServer.Serve(lis))

}

func startEid(grpcServer *grpc.Server) {
	fmt.Println("- Enabling EID")
	serve := serveid.New()
	geid.RegisterEIDServer(grpcServer, serve)

	if config.Get().BankID.Enabled {
		fmt.Println("  - Enabling BankId")
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
		fmt.Println("  - Enabling Freja")
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
