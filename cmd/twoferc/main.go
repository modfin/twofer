package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdp/qrterminal/v3"
	"io/ioutil"
	"log"
	"os"
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

			infered := c.Bool("inferred")
			ssn := c.String("ssn")
			contry := c.String("country")
			data := c.String("data")

			client := twoferrpc.NewEIDClient(conn)

			switch action {
			case "auth":
				inter, err := client.AuthInit(context.Background(), &twoferrpc.Req{
					Provider: &twoferrpc.Provider{Name: provider},
					Who: &twoferrpc.User{
						Inferred:   infered,
						Ssn:        ssn,
						SsnCountry: contry,
					},
				})
				if err != nil {
					return err
				}

				if infered {
					config := qrterminal.Config{
						Level:     qrterminal.M,
						Writer:    os.Stdout,
						BlackChar: qrterminal.WHITE,
						WhiteChar: qrterminal.BLACK,
						QuietZone: 2,
					}
					qrterminal.GenerateWithConfig(inter.QrData, config)
				}

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
						SsnCountry: contry,
					},
					Payload: &twoferrpc.Req_Payload{
						Text: data,
						Data: nil,
					},
				})
				if err != nil {
					return err
				}

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
