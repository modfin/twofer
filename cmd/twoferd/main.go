package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/modfin/twofer/internal/config"
	"github.com/modfin/twofer/internal/eid/bankid"
	"github.com/modfin/twofer/internal/httpserve"
	"github.com/modfin/twofer/internal/servotp"
	"github.com/modfin/twofer/internal/servpwd"
	"github.com/modfin/twofer/internal/servqr"
	"github.com/modfin/twofer/stream/ndjson"
	"github.com/modfin/twofer/stream/sse"
)

const shutdownGracePeriod = time.Second * 200 // A BankID auth order is only valid for 30 seconds, unless it's scanned, then it's valid for 180 seconds.

func main() {
	if len(os.Args) > 1 && os.Args[1] == "prestophook" {
		prestophook()
		return
	}

	cfg := config.Get()
	e := echo.New()
	e.Use(middleware.Logger())

	fmt.Println("Starting server")

	if cfg.OTP.Enabled {
		fmt.Println("- Enabling OTP")
		otpserv, err := servotp.New(servotp.OTPConfig{
			SkewCounter: config.Get().OTP.SkewCounter,
			SkewTime:    config.Get().OTP.SkewTime,
			RateLimit:   config.Get().OTP.RateLimit,
		}, config.Get().OTP.EncryptionKey)
		if err == nil {
			fmt.Println("  - Serving OTP via HTTP")
			httpserve.RegisterOTPServer(e, otpserv)
		} else {
			fmt.Println("Could not enable OTP", err)
		}
	}

	if cfg.QREnabled {
		fmt.Println("- Enabling QR")
		_servqr := servqr.New()
		fmt.Println("  - Serving QR via HTTP")
		httpserve.RegisterQRServer(e, _servqr)
	}

	if cfg.EIDEnabled() {
		startEid(e)
	}

	if cfg.WebAuthn.Enabled {
		/*
			fmt.Println("- Enabling WebAuthn")
			authn, err := servw6n.New(cfg.WebAuthn)
			if err != nil {
				fmt.Println("WebAuthn", err)
			} else {
				// TODO: gw6n.RegisterWebAuthnServer(grpcServer, authn)
			}
		*/
	}

	if cfg.PWD.Enabled {
		fmt.Println("- Enabling PWD")
		_servpwd, err := servpwd.New(servpwd.PWDConfig{
			DefaultAlg:          servpwd.Alg(cfg.PWD.DefaultAlg),
			DefaultHashCount:    cfg.PWD.DefaultHashCount,
			DefaultBCryptCost:   cfg.PWD.DefaultBCryptCost,
			DefaultSCryptN:      cfg.PWD.DefaultSCryptN,
			DefaultSCryptR:      cfg.PWD.DefaultSCryptR,
			DefaultSCryptP:      cfg.PWD.DefaultSCryptP,
			DefaultSCryptKeyLen: cfg.PWD.DefaultSCryptKeyLen,
		}, cfg.PWD.EncryptionKey)
		if err == nil {
			fmt.Println("  - Serving PWD via HTTP")
			httpserve.RegisterPWDServer(e, _servpwd)
		} else {
			fmt.Println("Could no start PWD grpc server")
		}
	}

	startServer(e)
}

func startServer(e *echo.Echo) {
	appCtx, appClose := context.WithCancel(context.Background())
	go func() {
		err := e.Start(":8080")
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err)
		}
		appClose()
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)

	select {
	case s := <-signalChannel:
		fmt.Printf("twoferd received signal: %v\n", s)
		appClose() // Cancel 'app context' when we receive SIGTERM
	case <-appCtx.Done():
	}

	fmt.Println("Graceful shutdown initiated...")

	timeout, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer cancel()

	err := e.Shutdown(timeout)
	if err != nil {
		log.Fatalf("failure during Echo's shutdown: %v", err)
	}

	fmt.Println("twoferd stopped")
}

func startEid(e *echo.Echo) {
	fmt.Println("- Enabling EID")
	//serve := serveid.New()
	//httpserve.RegisterEIDServer(e, serve)

	if !config.Get().BankID.Enabled {
		return
	}

	fmt.Println("  - Creating BankId")
	bankid, err := bankid.New(bankid.ClientConfig{
		BaseURL:       config.Get().BankID.URL.String(),
		PemRootCA:     config.Get().BankID.GetRootCA(),
		PemClientCert: config.Get().BankID.GetClientCert(),
		PemClientKey:  config.Get().BankID.GetClientKey(),
		PollInterval:  config.Get().BankID.PollInterval,
	})
	if err != nil {
		fmt.Printf("failed to initate bankId %v", err)
		return
	}

	//err = bankid.APIv51.Ping()
	//if err != nil {
	//	fmt.Printf("  - Err: Could not ping bankid v5.1. %v", err)
	//} else {
	//	fmt.Println("  - Adding BankId v5.1")
	//	fmt.Println("  - BankId Client Cert NotAfter:", bankid.ParsedClientCert().NotAfter)
	//	serve.Add(bankid.APIv51)
	//}

	fmt.Println("  - Adding BankId v6.0")
	fmt.Println("  - BankId Client Cert NotAfter:", bankid.ParsedClientCert().NotAfter)
	httpserve.RegisterBankIDServer(e, bankid.APIv60, getStreamEncoder(config.Get().StreamEncoder))
	err = bankid.APIv60.Ping()
	if err != nil {
		fmt.Printf("  - Err: Could not ping bankid. %v", err)
	}
}

func getStreamEncoder(encoder string) httpserve.NewStreamEncoder {
	switch encoder {
	case "SSE":
		return sse.NewEncoder
	default:
		return ndjson.NewEncoder
	}
}

func prestophook() {
	fmt.Println("twofer - prestophook")

	// Give K8S time to remove POD from service before stutdown is started
	time.Sleep(time.Second)

	// TODO: Check that twofer is actually PID 1 before we try to send signal?
	err := syscall.Kill(1, syscall.SIGINT)
	if err != nil {
		fmt.Printf("prestophook: SIGINT error: %v", err)
		return
	}

	// Wait for graceful shutdown period to end. When PID 1 have exited, we'll be terminated as well
	time.Sleep(shutdownGracePeriod)
}
