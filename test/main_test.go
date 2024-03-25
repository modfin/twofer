package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/bankid"
	"github.com/modfin/twofer/internal/httpserve"
	"github.com/modfin/twofer/test/fakes"
	"github.com/stretchr/testify/suite"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite
	twofer    *echo.Echo
	twoferURL string

	bankidv6 *fakes.BankIDV6Fake
}

func (s *IntegrationTestSuite) SetupSuite() {
	slog.Info("Setting up test suite")

	// BANKID v6
	s.bankidv6 = fakes.CreateBankIDV6Fake()
	go func() {
		slog.Info("Starting bankid v6 fake")
		err := s.bankidv6.Start()
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			fmt.Println("Error starting bankid v6 server.", err.Error())
		}
	}()

	//TWOFER
	app, err := InitApplication(s.bankidv6.URL)
	if err != nil {
		fmt.Println("Error setting up twofer in SetupSuite", err)
	}
	s.twofer = app
	s.twoferURL = "http://127.0.0.1:8999"

	go func() {
		slog.Info("Starting twofer")
		err := s.twofer.Start(":8999")
		if err != nil {
			fmt.Println("Error starting twofer", err.Error())
		}
	}()

	// Arbitrary wait for servers to start
	time.Sleep(1 * time.Second)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	slog.Info("Tearing down test suite")
	d := 1 * time.Second
	go func() {
		err := s.bankidv6.Stop(d)
		if err != nil {
			fmt.Println("Error stopping bankid v6 server.", err.Error())
		}
	}()

	go func() {
		parentCtx := context.Background()
		ctx, _ := context.WithTimeout(parentCtx, d)
		err := s.twofer.Shutdown(ctx)
		if err != nil {
			fmt.Println("Error stopping twofer server.", err.Error())
		}
	}()
}

func TestInit(t *testing.T) {
	slog.Info("TestInit")
	suite.Run(t, new(IntegrationTestSuite))
}

func InitApplication(bankIDV6URL string) (*echo.Echo, error) {
	e := echo.New()

	client := &http.Client{}

	twoferBankIDAPI := bankid.NewAPI(client, bankIDV6URL)
	httpserve.RegisterBankIDServer(e, twoferBankIDAPI)

	return e, nil
}
