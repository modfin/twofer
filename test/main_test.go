package test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/bankid"
	"github.com/modfin/twofer/internal/httpserve"
	"github.com/modfin/twofer/internal/ordertoken"
	"github.com/modfin/twofer/stream/sse"
	"github.com/modfin/twofer/test/fakes"
	"github.com/stretchr/testify/suite"
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
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Error starting twofer", err.Error())
		}
	}()

	// Arbitrary wait for servers to start
	time.Sleep(1 * time.Second)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	slog.Info("Tearing down test suite")
	d := 2 * time.Second
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

	key, err := generateKey(16)
	if err != nil {
		return nil, err
	}
	ec, ecPub, err := generateEcKey()
	if err != nil {
		return nil, fmt.Errorf("error generating ec key for test: %v", err)
	}
	otm, err := ordertoken.NewManager(ec, ecPub, []string{fmt.Sprintf("1:aes:%s", key)})
	if err != nil {
		return nil, fmt.Errorf("error creating ordertoken manager: %v", err)
	}

	twoferBankIDAPI := bankid.NewAPI(client, bankIDV6URL, time.Second)
	httpserve.RegisterBankIDServer(e, twoferBankIDAPI, otm, sse.NewEncoder)

	return e, nil
}

func generateEcKey() (string, string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	priv, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", "", err
	}
	pub, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return "", "", err
	}
	var privOut bytes.Buffer
	var pubOut bytes.Buffer
	err = pem.Encode(&privOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: priv})
	if err != nil {
		return "", "", err
	}
	err = pem.Encode(&pubOut, &pem.Block{Type: "PUBLIC KEY", Bytes: pub})
	if err != nil {
		return "", "", err
	}
	return privOut.String(), pubOut.String(), nil
}

func generateKey(len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}
