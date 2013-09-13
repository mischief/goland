package gutil

import (
	"github.com/stevedonovan/luar"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var testfile = `
config = {
  foo = "bar",

  sub = {
    one = 1,
    two = "two",
  }
}

return config
`

func TestLuaConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "config.lua")
	if err != nil {
		t.Fatal(err)
	}

	name := f.Name()
	f.WriteString(testfile)
	f.Close()

	defer os.Remove(name)

	L := luar.Init()

	config, err := NewLuaConfig(L, name)

	if err != nil {
		t.Fatal(err)
	}

	twoconf, err := config.Get("sub.two", reflect.String)
	if err != nil {
		t.Fatalf("sub.two: %s", err)
	}

	two, _ := twoconf.(string)

	if two != "two" {
		t.Fatalf("two was not two, was %q", two)
	}

}
