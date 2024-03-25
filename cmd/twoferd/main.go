package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4/middleware"
	"github.com/modfin/twofer/internal/config"
	"github.com/modfin/twofer/internal/eid/bankid"
	"github.com/modfin/twofer/internal/httpserve"
	"github.com/modfin/twofer/internal/serveid"
	"github.com/modfin/twofer/internal/servotp"
	"github.com/modfin/twofer/internal/servpwd"
	"github.com/modfin/twofer/internal/servqr"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
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
	go func() {
		fmt.Println(e.Start(":8080"))
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)

	<-signalChannel
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		err := e.Shutdown(timeout)
		if err != nil {
			log.Fatalf("failure during Echo's shutdown: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

func startEid(e *echo.Echo) {
	fmt.Println("- Enabling EID")
	serve := serveid.New()
	httpserve.RegisterEIDServer(e, serve)

	if !config.Get().BankID.Enabled {
		return
	}

	fmt.Println("  - Creating BankId")
	bankid, err := bankid.New(bankid.ClientConfig{
		BaseURL:       config.Get().BankID.URL.String(),
		PemRootCA:     config.Get().BankID.GetRootCA(),
		PemClientCert: config.Get().BankID.GetClientCert(),
		PemClientKey:  config.Get().BankID.GetClientKey(),
	})
	if err != nil {
		fmt.Printf("failed to initate bankId %v", err)
		return
	}

	err = bankid.APIv51.Ping()
	if err != nil {
		fmt.Printf("  - Err: Could not ping bankid v5.1. %v", err)
	} else {
		fmt.Println("  - Adding BankId v5.1")
		fmt.Println("  - BankId Client Cert NotAfter:", bankid.ParsedClientCert().NotAfter)
		serve.Add(bankid.APIv51)
	}

	err = bankid.APIv60.Ping()
	if err != nil {
		fmt.Printf("  - Err: Could not ping bankid. %v", err)
	} else {
		fmt.Println("  - Adding BankId v6.0")
		fmt.Println("  - BankId Client Cert NotAfter:", bankid.ParsedClientCert().NotAfter)
		httpserve.RegisterBankIDServer(e, bankid.APIv60)
	}
}
