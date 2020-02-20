package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdp/qrterminal/v3"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"twofer/twoferrpc"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

var serverAddr string = "127.0.0.1:1234"

func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	fmt.Println("Dialing grpc..")
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	eidAction := func(provider string, action string) cli.ActionFunc {
		return func(c *cli.Context) error {

			switch provider {
			case "BankID", "FrejaID":
			default:
				return errors.New("freja or bankid must be selected")
			}

			inferred := c.Bool("inferred")
			ssn := c.String("ssn")
			country := c.String("country")
			data := c.String("data")
			client := twoferrpc.NewEIDClient(conn)

			switch action {
			case "auth":
				inter, err := client.AuthInit(context.Background(), &twoferrpc.Req{
					Provider: &twoferrpc.Provider{Name: provider},
					Who: &twoferrpc.User{
						Inferred:   inferred,
						Ssn:        ssn,
						SsnCountry: country,
					},
				})
				if err != nil {
					return err
				}

				if inferred {
					config := qrterminal.Config{
						Level:     qrterminal.M,
						Writer:    os.Stdout,
						BlackChar: qrterminal.BLACK,
						WhiteChar: qrterminal.WHITE,
						QuietZone: 2,
					}
					qrterminal.GenerateWithConfig(inter.QrData, config)
				}
				fmt.Printf("%+v\n", inter)

				resp, err := client.Collect(context.Background(), inter)
				if err != nil {
					return err
				}

				fmt.Printf("%+v\n", resp.Info)

			case "sign":

				if len(ssn) == 0 {
					return errors.New("an ssn must be provided for signing, this can not be inferred")
				}
				inter, err := client.SignInit(context.Background(), &twoferrpc.Req{
					Provider: &twoferrpc.Provider{Name: provider},
					Who: &twoferrpc.User{
						Inferred:   false,
						Ssn:        ssn,
						SsnCountry: country,
					},
					Payload: &twoferrpc.Req_Payload{
						Text: data,
						Data: nil,
					},
				})
				if err != nil {
					return err
				}
				fmt.Printf("EIDRequest%+v\n", inter)
				resp, err := client.Collect(context.Background(), inter)
				if err != nil {
					return err
				}

				fmt.Printf("%+v\n", resp.Info)

			default:
				return errors.New("sign or auth must be provided")
			}

			return nil
		}
	}

	app := &cli.App{
		Name: "twoferc",
		Commands: []*cli.Command{
			{
				Name:  "qr",
				Usage: "Generates a qr image",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "data",
						Aliases: []string{"d"},
						Usage:   "data to be contained in image",
					},
					&cli.StringFlag{
						Name:    "out",
						Aliases: []string{"o"},
						Usage:   "file output",
					},
				},
				Action: func(c *cli.Context) error {

					data := c.String("data")
					filename := c.String("out")
					if len(data) == 0 {
						return errors.New("data must be provided, --data something ")
					}
					if len(filename) == 0 {
						return errors.New("a output file must be provided")
					}

					qr := twoferrpc.NewQRClient(conn)

					image, err := qr.Generate(context.Background(), &twoferrpc.QRData{
						RecoveryLevel: 2,
						Size:          256,
						Data:          data,
					})

					if err != nil {
						return err
					}

					return ioutil.WriteFile(filename, image.Data, 0660)
				},
			},
			{
				Name:  "otp",
				Usage: "Generates a qr image",
				Subcommands: []*cli.Command{
					{
						Name: "enroll",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "issuer",
								Aliases: []string{"i"},
								Usage:   "issuer name",
							},
							&cli.StringFlag{
								Name:    "user",
								Aliases: []string{"u"},
								Usage:   "username",
							},
							&cli.StringFlag{
								Name:    "alg",
								Aliases: []string{"a"},
								Usage:   "algorithm used",
								Value:   "sha1",
							},
							&cli.UintFlag{
								Name:    "period",
								Aliases: []string{"p"},
								Usage:   "time window for totp",
								Value:   30,
							},
							&cli.UintFlag{
								Name:    "digits",
								Aliases: []string{"d"},
								Usage:   "otp length, 6/8",
								Value:   6,
							},

							&cli.UintFlag{
								Name:    "secret-size",
								Aliases: []string{"s"},
								Usage:   "secret-size in bytes",
								Value:   32,
							},
							&cli.StringFlag{
								Name:    "mode",
								Aliases: []string{"m"},
								Usage:   "otp type, time/counter",
								Value:   "time",
							},
						},
						Action: func(c *cli.Context) error {

							mode := c.String("mode")
							issuer := c.String("issuer")

							user := c.String("user")
							alg := c.String("alg")
							period := c.Uint("period")
							digits := c.Uint("digits")
							ss := c.Int("secret-size")

							var ralg twoferrpc.OTPAlg
							switch strings.ToLower(alg) {
							case "sha1":
								ralg = twoferrpc.OTPAlg_SHA_1
							case "sha512":
								ralg = twoferrpc.OTPAlg_SHA_512
							case "sha256":
								fallthrough
							default:
								ralg = twoferrpc.OTPAlg_SHA_1
							}

							var rmode twoferrpc.OTPMode
							switch mode {
							case "time":
								rmode = twoferrpc.OTPMode_TIME
							case "counter":
								rmode = twoferrpc.OTPMode_COUNTER
							default:
								return errors.New("not a vaild mode")
							}

							var rdigits twoferrpc.OTPDigits
							switch digits {
							case 6:
								rdigits = twoferrpc.OTPDigits_SIX
							case 8:
								rdigits = twoferrpc.OTPDigits_EIGHT
							default:
								return errors.New("digits must be 6 or 8")
							}

							client := twoferrpc.NewOTPClient(conn)

							r, err := client.Enroll(context.Background(), &twoferrpc.OTPEnrollment{
								Issuer:     issuer,
								Account:    user,
								Alg:        ralg,
								Mode:       rmode,
								Digits:     rdigits,
								Period:     uint32(period),
								SecretSize: uint32(ss),
							})

							if err != nil {
								return err
							}
							config := qrterminal.Config{
								Level:     qrterminal.M,
								Writer:    os.Stdout,
								BlackChar: qrterminal.BLACK,
								WhiteChar: qrterminal.WHITE,
								QuietZone: 2,
							}
							qrterminal.GenerateWithConfig(r.Uri, config)

							fmt.Println(r.Uri)
							fmt.Println(r.Secret)

							return nil
						},
					},
					{
						Name: "validate",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "secret",
								Aliases: []string{"s"},
								Usage:   "the secret data",
							},
							&cli.StringFlag{
								Name:    "otp",
								Aliases: []string{"o"},
								Usage:   "the current otp",
							},
						},
						Action: func(c *cli.Context) error {

							secret := c.String("secret")
							otp := c.String("otp")

							client := twoferrpc.NewOTPClient(conn)

							r, err := client.Validate(context.Background(), &twoferrpc.OTPValidate{
								Otp:    otp,
								Secret: secret,
							})

							if err != nil {
								return err
							}

							fmt.Println("Valid:", r.Valid)
							fmt.Println(r.Secret)

							return nil
						},
					},
				},
			},
			{
				Name:  "eid",
				Usage: "eid action",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "inferred",
						Aliases: []string{"i"},
						Usage:   "inferred auth",
					},
					&cli.StringFlag{
						Name:    "ssn",
						Aliases: []string{"s"},
						Usage:   "ssn of person",
					},
					&cli.StringFlag{
						Name:    "country",
						Aliases: []string{"c"},
						Usage:   "country of person, which the ssn relates to",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name: "freja",
						Subcommands: []*cli.Command{
							{
								Name:   "auth",
								Action: eidAction("FrejaID", "auth"),
							},
							{
								Name:   "sign",
								Action: eidAction("FrejaID", "sign"),
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "data",
										Aliases: []string{"d"},
										Usage:   "data to sign",
									},
								},
							},
						},
					},
					{
						Name: "bankid",
						Subcommands: []*cli.Command{
							{
								Name:   "auth",
								Action: eidAction("BankID", "auth"),
							},
							{
								Name:   "sign",
								Action: eidAction("BankID", "sign"),
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "data",
										Aliases: []string{"d"},
										Usage:   "data to sign",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return
}
