package main

import (
	"fmt"
	"github.com/modfin/twofer/eid/bankid"
	"github.com/modfin/twofer/eid/freja"
	"github.com/modfin/twofer/grpc/geid"
	"github.com/modfin/twofer/grpc/gotp"
	"github.com/modfin/twofer/grpc/gpwd"
	"github.com/modfin/twofer/grpc/gqr"
	"github.com/modfin/twofer/grpc/gw6n"
	"github.com/modfin/twofer/internal/config"
	"github.com/modfin/twofer/internal/httpserve"
	"github.com/modfin/twofer/internal/serveid"
	"github.com/modfin/twofer/internal/servotp"
	"github.com/modfin/twofer/internal/servpwd"
	"github.com/modfin/twofer/internal/servqr"
	"github.com/modfin/twofer/internal/servw6n"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

func main() {

	cfg := config.Get()
	e := echo.New()

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
			if cfg.EnableHttp {
				fmt.Println("  - Serving OTP via HTTP")
				httpserve.RegisterOTPServer(e, otpserv)
			}
		} else {
			fmt.Println("Could not enable OTP", err)
		}
	}

	if cfg.QREnabled {
		fmt.Println("- Enabling QR")
		_servqr := servqr.New()
		gqr.RegisterQRServer(grpcServer, _servqr)
		if cfg.EnableHttp {
			fmt.Println("  - Serving QR via HTTP")
			httpserve.RegisterQRServer(e, _servqr)
		}
	}

	if cfg.EIDEnabled() {
		startEid(grpcServer, e)
	}

	if cfg.WebAuthn.Enabled {
		fmt.Println("- Enabling WebAuthn")
		authn, err := servw6n.New(cfg.WebAuthn)
		if err != nil {
			fmt.Println("WebAuthn", err)
		} else {
			gw6n.RegisterWebAuthnServer(grpcServer, authn)
		}
	}

	if cfg.PWD.Enabled {
		fmt.Println("- Enabling PWD")
		_servpwd, err := servpwd.New(servpwd.PWDConfig{
			DefaultAlg:          gpwd.Alg(cfg.PWD.DefaultAlg),
			DefaultHashCount:    cfg.PWD.DefaultHashCount,
			DefaultBCryptCost:   cfg.PWD.DefaultBCryptCost,
			DefaultSCryptN:      cfg.PWD.DefaultSCryptN,
			DefaultSCryptR:      cfg.PWD.DefaultSCryptR,
			DefaultSCryptP:      cfg.PWD.DefaultSCryptP,
			DefaultSCryptKeyLen: cfg.PWD.DefaultSCryptKeyLen,
		}, cfg.PWD.EncryptionKey)
		if err == nil {
			gpwd.RegisterPWDServer(grpcServer, _servpwd)
			if cfg.EnableHttp {
				fmt.Println("  - Serving PWD via HTTP")
				httpserve.RegisterPWDServer(e, _servpwd)
			}
		} else {
			fmt.Println("Could no start PWD grpc server")
		}
	}

	startServer(grpcServer, e)

}

func startServer(grpcServer *grpc.Server, e *echo.Echo) {

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
			fmt.Println(e.Start(":8080"))
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

func startEid(grpcServer *grpc.Server, e *echo.Echo) {
	fmt.Println("- Enabling EID")
	serve := serveid.New()
	geid.RegisterEIDServer(grpcServer, serve)
	httpserve.RegisterEIDServer(e, serve)

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
