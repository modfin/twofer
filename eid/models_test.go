package eid

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"twofer/grpc/geid"
)

func diffObj(expected interface{}, got interface{}) string {
	c := spew.NewDefaultConfig()
	c.DisablePointerAddresses = true

	e := strings.Split(c.Sdump(expected), "\n")
	g := strings.Split(c.Sdump(got), "\n")

	l := len(g)
	for i := range e {
		e[i] = "  " + e[i]
		if l <= i {
			g = append(g, "? ")
			continue
		}
		g[i] = "  " + g[i]
		if e[i] != g[i] {
			g[i] = "!" + g[i][1:]
		}
	}
	return fmt.Sprintf("expected:\n%s\ngot:\n%s\n", strings.Join(e, "\n"), strings.Join(g, "\n"))
}

var testClient = &TestClient{}
var fromGrpcInterTests = []struct {
	in  *geid.Inter
	err string
	res Inter
}{
	{
		in: &geid.Inter{
			Req: &geid.Req{
				Who: &geid.User{},
			},
			Mode: geid.Inter_AUTH,
		},
		err: "",
		res: Inter{
			Req: &Req{
				Provider: testClient,
				Who:      &User{},
				Payload:  nil,
			},
			Mode: AUTH,
		},
	},
	{
		in: &geid.Inter{
			Req: &geid.Req{
				Who: &geid.User{},
			},
			Mode: geid.Inter_SIGN,
		},
		err: "",
		res: Inter{
			Req: &Req{
				Provider: testClient,
				Who:      &User{},
			},
			Mode: SIGN,
		},
	},
	{
		in: &geid.Inter{
			Req: &geid.Req{
				Who: &geid.User{},
			},
			Mode: geid.Inter_AUTH,
		},
		err: "",
		res: Inter{
			Req: &Req{
				Provider: testClient,
				Who:      &User{},
			},
			Mode: AUTH,
		},
	},
}

func TestFromGrpcInter(t *testing.T) {
	for _, test := range fromGrpcInterTests {
		i, err := FromGrpcInter(test.in, testClient)

		if test.err != "" {
			assert.EqualError(t, err, test.err)
		}

		if !assert.ObjectsAreEqual(test.res, i) {
			fmt.Println(diffObj(test.res, i))
			t.FailNow()
		}
	}

}

type TestClient struct {
}

func (c *TestClient) Name() (s string) {
	return "wat"
}

func (c *TestClient) AuthInit(ctx context.Context, req *Req) (*Inter, error) {
	return &Inter{}, nil
}
func (c *TestClient) SignInit(ctx context.Context, req *Req) (*Inter, error) {
	return &Inter{}, nil
}

func (c *TestClient) Peek(ctx context.Context, req *Inter) (*Resp, error) {
	return &Resp{}, nil
}
func (c *TestClient) Collect(ctx context.Context, req *Inter, cancelOnErr bool) (*Resp, error) {
	return &Resp{}, nil
}
func (c *TestClient) Cancel(intermediate *Inter) error {
	return nil
}
