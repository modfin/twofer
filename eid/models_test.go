package eid

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"twofer/twoferrpc"
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

var fromGrpcInterTests = []struct {
	in  *twoferrpc.Inter
	err string
	res Inter
}{
	{
		in: &twoferrpc.Inter{
			Req: &twoferrpc.Req{
				Who: &twoferrpc.User{},
			},
			Mode: twoferrpc.Inter_AUTH,
		},
		err: "",
		res: Inter{
			Req: &Req{
				Provider: nil,
				Who:      &User{},
				Payload:  nil,
			},
			Mode: AUTH,
		},
	},
	{
		in: &twoferrpc.Inter{
			Req: &twoferrpc.Req{
				Who: &twoferrpc.User{},
			},
			Mode: twoferrpc.Inter_SIGN,
		},
		err: "",
		res: Inter{
			Req: &Req{
				Who: &User{},
			},
			Mode: SIGN,
		},
	},
	{
		in: &twoferrpc.Inter{
			Req: &twoferrpc.Req{
				Who: &twoferrpc.User{},
			},
			Mode: twoferrpc.Inter_AUTH,
		},
		err: "",
		res: Inter{
			Req: &Req{
				Who: &User{},
			},
			Mode: AUTH,
		},
	},
}

func TestFromGrpcInter(t *testing.T) {

	for _, test := range fromGrpcInterTests {
		i, err := FromGrpcInter(test.in)

		if test.err != "" {
			assert.EqualError(t, err, test.err)
		}

		if !assert.ObjectsAreEqual(test.res, i) {
			fmt.Println(diffObj(test.res, i))
			t.FailNow()
		}
	}

}
