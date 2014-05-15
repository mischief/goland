package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"

	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/srv"

	"github.com/golang/glog"

	"bitbucket.org/mischief/libauth"
)

// AuthOps below

type AuthFid struct {
	rpc    *libauth.AuthRpc
	uname  string
	aname  string
	authok bool
	afd    io.ReadWriteCloser
}

func (af *AuthFid) Read(p []byte) (n int, err error) {
	st, res := af.rpc.Rpc("read", "")
	copy(p, []byte(res))
	switch st {
	case libauth.ARdone:
		if err := af.rpc.GetInfo(); err != nil {
			return 0, err
		}
		af.authok = true

		return len(res), nil
	case libauth.ARok:
		return len(res), nil
	case libauth.ARphase:
		fallthrough
	default:
		return 0, fmt.Errorf("authfid read: phase error: %s", res)
	}

	return 0, nil
}

var authgen uint64 = 1 << 63

type p9sk1auth struct {
	opener func() (io.ReadWriteCloser, error)
}

func NewP9SK1Authenticator(gs *GameServer) *p9sk1auth {
	var fdriver, fspec string
	if driverconf, err := gs.config.Get("factotum.driver", reflect.String); err != nil {
		glog.Fatalf("missing db.driver in config")
	} else {
		fdriver = driverconf.(string)
	}
	if specconf, err := gs.config.Get("factotum.spec", reflect.String); err != nil {
		glog.Fatalf("missing db.driver in config")
	} else {
		fspec = specconf.(string)
	}

	p9sk1 := p9sk1auth{}

	switch fdriver {
	case "tcp":
		p9sk1.opener = func() (io.ReadWriteCloser, error) {
			return net.Dial("tcp", fspec)
		}
	case "rpc":
		p9sk1.opener = libauth.OpenRPC
	}

	return &p9sk1
}

func (*p9sk1auth) AuthCheck(fid *srv.Fid, afid *srv.Fid, aname string) error {
	glog.Infof("AuthCheck: fid %#v afid %#v aname %s", fid, afid, aname)

	if afid == nil {
		return srv.Eperm
	}

	aux := afid.Aux.(*AuthFid)

	if afid.Type&p.QTAUTH == 0 || aux == nil {
		return srv.Ebaduse
	}

	if !aux.authok {
		if _, err := aux.Read(nil); err != nil {
			return err
		}
	}

	if aux.uname != afid.User.Name() {
		return fmt.Errorf("uname mismatch: %s vs %s", afid.User.Name(), aux.uname)
	}

	if aux.aname != aname {
		return fmt.Errorf("aname mismatch: %s vs %s", aname, aux.aname)
	}

	return nil
}

func (*p9sk1auth) AuthDestroy(afid *srv.Fid) {
	aux := afid.Aux.(*AuthFid)
	glog.Infof("AuthDestroy: %v", afid)

	aux.rpc.F.Close()
}

func (p9 *p9sk1auth) AuthInit(afid *srv.Fid, aname string) (*p.Qid, error) {
	f, err := p9.opener()
	if err != nil {
		return nil, errors.New("OpenRPC: " + err.Error())
	}

	rpc := &libauth.AuthRpc{
		F: f,
	}

	aux := new(AuthFid)
	aux.rpc = rpc

	aux.uname = afid.User.Name()
	aux.aname = aname

	if st, err := rpc.Rpc("start", "proto=p9any role=server"); st != libauth.ARok {
		f.Close()
		glog.Errorf("auth rpc error: %s", err)
		return nil, fmt.Errorf("%s", err)
	}

	var aqid p.Qid

	glog.Infof("AuthInit: %+v %s", afid, aname)

	aqid.Type |= p.QTAUTH
	aqid.Path = authgen
	authgen++

	aqid.Version = 0

	afid.Omode = p.ORDWR
	afid.Aux = aux

	return &aqid, nil
}

func (*p9sk1auth) AuthRead(afid *srv.Fid, offset uint64, data []byte) (count int, err error) {
	glog.Infof("AuthRead: %v %d", afid, offset)

	aux := afid.Aux.(*AuthFid)

	if afid == nil {
		return 0, srv.Eperm
	}

	if afid.Type&p.QTAUTH == 0 || aux == nil {
		return 0, srv.Ebaduse
	}

	n, err := aux.Read(data)
	if err != nil {
		return 0, fmt.Errorf("authread: %s", err)
	}

	return n, nil
}

func (*p9sk1auth) AuthWrite(afid *srv.Fid, offset uint64, data []byte) (count int, err error) {
	glog.Infof("AuthWrite: %v %d %q", afid, offset, data)

	aux := afid.Aux.(*AuthFid)

	if afid == nil {
		return 0, srv.Eperm
	}

	if afid.Type&p.QTAUTH == 0 || aux == nil {
		return 0, srv.Ebaduse
	}

	st, res := aux.rpc.Rpc("write", string(data))
	if st != libauth.ARok {
		err = fmt.Errorf("p9sk1: write: %d %s", st, res)
		glog.Error(err)
		return 0, err
	}

	return len(data), nil
}
