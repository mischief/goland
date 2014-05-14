package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"net"

	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/clnt"

	"github.com/golang/glog"

	"github.com/mischief/goland/game/gauth"
)

type FakeUser string

func (u *FakeUser) Name() string            { return string(*u) }
func (u *FakeUser) Id() int                 { return 61508 }
func (u *FakeUser) Groups() []p.Group       { return []p.Group{u} }
func (u *FakeUser) IsMember(g p.Group) bool { return true }
func (u *FakeUser) Members() []p.User       { return []p.User{u} }

var (
	myuser  = FakeUser("authsrv")
	keyaddr = flag.String("k", "127.0.0.1:61509", "key server")
)

func mkkey() []byte {
	k := make([]byte, gauth.DESKEYLEN)
	if _, err := rand.Read(k); err != nil {
		panic(err)
	}
	return k
}

func findkey(cl *clnt.Clnt, user string) ([]byte, error) {
	f, err := cl.FOpen(fmt.Sprintf("%s/key", user), p.OREAD)
	if err != nil {
		return nil, fmt.Errorf("findkey: %s", err)
	}
	defer f.Close()

	ret := make([]byte, gauth.DESKEYLEN)
	if n, err := f.ReadAt(ret, 0); n != gauth.DESKEYLEN {
		return nil, fmt.Errorf("findkey: short key: %d", n)
	} else if err != nil {
		return nil, fmt.Errorf("findkey: %s", err)
	}

	return ret, nil
}

type authcli struct {
	con net.Conn
	p9  *clnt.Clnt
}

func (a *authcli) terr(err error) {
	a.con.Write([]byte{gauth.AuthErr})
	fmt.Fprintf(a.con, "%s", err)
}

func (a *authcli) serve() {
	defer a.con.Close()

	p9, err := clnt.Mount("tcp", *keyaddr, "", &myuser)
	if err != nil {
		a.terr(err)
		glog.Error(err)
		return
	} else {
		a.p9 = p9
	}

	defer a.p9.Unmount()

	tr := &gauth.Ticketreq{}
	err = binary.Read(a.con, binary.LittleEndian, tr)
	if err != nil {
		a.terr(err)
		glog.Errorf("tr-fail %s", err)
		return
	}

	glog.Infof("tr %s", tr)

	switch tr.Type {
	case gauth.AuthTreq:
		a.ticketrequest(tr)
	default:
		glog.Error("unhandled Ticketreq %d", tr.Type)
	}
}

func (a *authcli) ticketrequest(tr *gauth.Ticketreq) {
	var akey, hkey, m []byte
	var err error
	var t gauth.Ticket

	if akey, err = findkey(a.p9, string(tr.Authid[:])); err != nil {
		goto fail
	}

	if hkey, err = findkey(a.p9, string(tr.Authid[:])); err != nil {
		goto fail
	}

	copy(t.Chal[:], tr.Chal[:])
	copy(t.Cuid[:], tr.Uid[:])

	/* speaksfor(tr.Hostid, tr.Uid) */
	copy(t.Suid[:], tr.Uid[:])

	copy(t.Key[:], mkkey())

	a.con.Write([]byte{gauth.AuthOK})

	t.Num = gauth.AuthTc
	m = t.ToM(hkey)
	a.con.Write(m)

	t.Num = gauth.AuthTs
	m = t.ToM(akey)
	a.con.Write(m)

	return

fail:
	glog.Error(err)
	a.terr(err)
	return
}

func main() {
	flag.Parse()

	//listener, err := net.Listen("tcp", ":61510")
	listener, err := net.Listen("tcp", ":567")
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		c, err := listener.Accept()
		if err != nil {
			glog.Error(err)
			break
		}
		glog.Infof("accept from %s", c.RemoteAddr())
		cli := &authcli{con: c}
		go cli.serve()
	}
}
