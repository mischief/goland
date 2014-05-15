package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/srv"

	"github.com/golang/glog"

	"github.com/mischief/goland/game/gfx"
	"github.com/mischief/goland/game/gid"
	"github.com/mischief/goland/game/gobj"
)

type Dirgen func(uint64, interface{}) *p.Dir

func dirread9p(req *srv.Req, gen Dirgen, i interface{}) {
	var start uint64
	var count uint32

	p.InitRread(req.Rc, req.Tc.Count)

	if req.Tc.Offset == 0 {
		start = 0
	} else {
		start = req.Fid.Diroffset
	}

	b := req.Rc.Data

	for len(b) > 0 {
		d := gen(start, i)
		if d == nil {
			break
		}
		sz := p.PackDir(d, b, req.Conn.Dotu)
		if sz == 0 {
			break
		}

		b = b[sz:]
		count += uint32(sz)
		start++
	}

	//req.Fid.Diroffset += count

	p.SetRreadCount(req.Rc, uint32(count))
	req.Respond()
}

const (
	Qroot = iota
	Qrctl
	Qevt

	// zone/
	Qzonedir

	// obj/
	Qobjdir

	// zone/n/
	Qzone

	Qzname
	Qzmap

	// obj/n/
	Qobj

	Qoctl
	Qobjid
	Qobjname
	Qobjx
	Qobjy
)

var nametab = map[int]string{
	Qroot: "/",
	Qrctl: "ctl",
	Qevt:  "event",

	// zone/
	Qzonedir: "zone",

	// zone/n/
	Qzone: "",

	Qzname: "name",
	Qzmap:  "map",

	// obj/
	Qobjdir: "obj",

	// obj/n/
	Qobj: "",

	Qoctl: "ctl",

	Qobjid:   "id",
	Qobjname: "name",
	Qobjx:    "x",
	Qobjy:    "y",
}

func fsmkqid(level, id uint64) *p.Qid {
	var q p.Qid

	q.Type = p.QTFILE
	q.Version = 0

	switch level {
	case Qroot:
		fallthrough
	case Qobjdir, Qzonedir:
		fallthrough
	case Qobj, Qzone:
		q.Type = p.QTDIR
		fallthrough
	case Qrctl:
		fallthrough
	case Qevt:
		fallthrough
	case Qzmap, Qzname:
		fallthrough
	case Qoctl, Qobjid, Qobjname, Qobjx, Qobjy:
		fallthrough
	default:
		q.Path = uint64((level << 17) | id)
	}

	return &q
}

type GameFid struct {
	level uint64
	qid   *p.Qid

	// XXX
	id    gid.Gid
	owner p.User
}

type GameFs struct {
	*p9sk1auth
	srv.Srv
	gs    *GameServer
	conns map[*srv.Conn]p.User
}

func NewGameFs(gs *GameServer) *GameFs {
	gfs := &GameFs{
		p9sk1auth: NewP9SK1Authenticator(gs),
		gs:        gs,
		conns:     make(map[*srv.Conn]p.User),
	}

	return gfs
}

func (gfs *GameFs) fsmkdir(level, id uint64) *p.Dir {
	var d p.Dir

	sysuser := gfs.gs.DBM.Uname2User("sys")
	sysgroup := sysuser.Groups()[0]

	d.Uid = sysuser.Name()
	d.Uidnum = uint32(sysuser.Id())
	d.Gid = sysgroup.Name()
	d.Gidnum = uint32(sysgroup.Id())
	d.Muid = sysuser.Name()
	d.Muidnum = uint32(sysuser.Id())

	d.Qid = *fsmkqid(level, id)
	d.Mode = 0444

	now := uint32(time.Now().Unix())

	d.Atime = now
	d.Mtime = now

	if d.Qid.Type&p.QTDIR > 0 {
		d.Mode |= p.DMDIR | 0111
	}

	d.Name = nametab[int(level)]

	switch level {
	case Qrctl:
		d.Mode = 0666
	case Qoctl:
		d.Mode = 0660
	case Qobj, Qzone:
		d.Name = fmt.Sprintf("%d", id)
	}

	return &d
}

// FS Ops

func (gfs *GameFs) Attach(req *srv.Req) {
	if req.Afid == nil {
		req.RespondError(srv.Eperm)
		return
	}

	/*
		if len(req.Tc.Aname) == 0 {
			req.RespondError(srv.Enotimpl)
			return
		}
	*/

	fid := new(GameFid)
	fid.level = Qroot
	fid.qid = fsmkqid(Qroot, 0)
	fid.owner = gfs.gs.DBM.Uname2User(req.Tc.Uname)

	req.Fid.Aux = fid

	id := gid.Gen()
	fid.id = id
	newplayer := gobj.NewGameObject(id, req.Tc.Uname)
	newplayer.SetTag("player", true)
	newplayer.SetTag("visible", true)

	// setting this lets players pick up other players, lol
	//newplayer.SetTag("gettable", true)
	newplayer.SetSprite(gfx.Get("human"))
	newplayer.SetPos(256/2, 256/2)

	// set the session's object
	//cp.Client.Player = newplayer

	// put player object in world
	gfs.gs.AddObject(newplayer, fid.owner)
	gfs.conns[req.Conn] = fid.owner
	req.RespondRattach(fid.qid)
}

func (gfs *GameFs) Flush(req *srv.Req) {}

func (gfs *GameFs) Walk1(fid *srv.Fid, name string, qid *p.Qid) (err error) {
	var i, id int
	gfid := fid.Aux.(*GameFid)

	if gfid.qid.Type&p.QTDIR == 1 {
		return fmt.Errorf("walk in non-directory")
	}

	glog.Infof("walk1 %d %s", gfid.level, name)
	if name == ".." {
		switch gfid.level {
		case Qroot:
			break
		default:
			if gfid.level >= Qobj {
				gfid.level = Qobjdir
			} else if gfid.level > Qzone && gfid.level < Qobj {
				gfid.level = Qzonedir
			} else {
				gfid.level = Qroot
			}
		}
	} else {

	loop:
		for i = int(gfid.level + 1); i < len(nametab); i++ {
			if nametab[i] != "" {
				if nametab[i] == name {
					id = int(gfid.id)
					glog.Infof("matched %s (%d)", name, id)
					break loop
				}
			}

			if i == Qzone {
				/* zone 0 */
				id, err = strconv.Atoi(name)
				if err == nil && id == 0 {
					break loop
				}
			}

			if i == Qobj {
				id, err = strconv.Atoi(name)
				if err == nil {
					if obj := gfs.gs.Objects.FindObjectByID(gid.Gid(id)); obj != nil {
						break loop
					} else {
						glog.Infof("query for missing object %d", id)
						return fmt.Errorf("no such object %d", id)
					}
				}
			}
		}
		if i >= len(nametab) {
			return fmt.Errorf("directory entry not found")
		}
		gfid.level = uint64(i)
	}

	q := fsmkqid(uint64(gfid.level), uint64(id))
	*qid = *q
	gfid.qid = q
	gfid.id = gid.Gid(id)
	glog.Infof("level %d id %d qid %s", gfid.level, gfid.id, q)
	return nil
}

func (gfs *GameFs) Clone(oldf, newf *srv.Fid) error {
	newfid := new(GameFid)
	*newfid = *oldf.Aux.(*GameFid)
	newf.Aux = newfid
	glog.Infof("newfid now level %d id %d qid %s", newfid.level, newfid.id, newfid.qid)
	return nil
}

func (gfs *GameFs) Walk(req *srv.Req) {
	if req.Fid != req.Newfid {
		if err := gfs.Clone(req.Fid, req.Newfid); err != nil {
			req.RespondError(err)
			return
		}
	}

	var err error

	wqids := make([]p.Qid, len(req.Tc.Wname))
	nqids := 0
	for i := 0; i < len(req.Tc.Wname); i++ {
		var q p.Qid
		err = gfs.Walk1(req.Newfid, req.Tc.Wname[i], &q)

		if err != nil {
			break
		}

		newf := req.Newfid.Aux.(*GameFid)
		newf.qid = &q
		wqids[i] = q
		nqids++
	}

	if err != nil && nqids == 0 {
		req.RespondError(err)
		return
	} else {
		req.RespondRwalk(wqids[0:nqids])
		return
	}

	/*
		fid := req.Fid.Aux.(*GameFid)
		tc := req.Tc

		if fid.level != Qroot {
			req.RespondError("walk not in root")
			return
		}

		if req.Newfid.Aux == nil {
			req.Newfid.Aux = new(GameFid)
		}

		nfid := req.Newfid.Aux.(*GameFid)

		wqids := make([]p.Qid, len(tc.Wname))
		nqids := 0

		if len(tc.Wname) > 0 {
			switch tc.Wname[0] {
			case "ctl":
				nfid.level = Qctl
			case "event":
				nfid.level = Qevt
			case "obj":
				nfid.level = Qobjdir
			}
			wqids[0] = *fsmkqid(nfid.level, 0)
			nqids = 1
		} else {
			nfid.level = Qroot
		}

		req.RespondRwalk(wqids[:nqids])
	*/
	//req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Open(req *srv.Req) {
	fid := req.Fid.Aux.(*GameFid)
	d := fsmkqid(fid.level, uint64(fid.id))
	glog.Infof("open %d %q %d owner %s %q", fid.level, fid.qid, fid.id, fid.owner, d)
	fid.qid = d
	req.RespondRopen(d, 8192)
	return
}

func (gfs *GameFs) Create(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Read(req *srv.Req) {
	fid := req.Fid.Aux.(*GameFid)
	tc := req.Tc

	rootgen := func(off uint64, i interface{}) *p.Dir {
		off += Qroot + 1

		if off < Qzone {
			rp := gfs.fsmkdir(off, 0)
			if rp.Mode&p.DMDIR == 0 {
				rp.Uid = req.Fid.User.Name()
				rp.Gid = req.Fid.User.Name()
			}
			return rp
		}
		//off -= Qobj
		return nil
	}

	zonedirgen := func(off uint64, i interface{}) *p.Dir {
		if off == 0 {
			p := gfs.fsmkdir(Qzone, uint64(0))
			return p
		}
		return nil
	}

	zonegen := func(off uint64, i interface{}) *p.Dir {
		off += Qzone + 1
		if off > Qzmap {
			return nil
		}

		dir := gfs.fsmkdir(off, uint64(fid.id))
		return dir
	}

	objdirgen := func(off uint64, i interface{}) *p.Dir {
		objch := i.(<-chan gobj.Object)

		if off+1 > uint64(cap(objch)) {
			return nil
		}
		if obj, ok := <-objch; ok {
			glog.Infof("%d %v %s %v", off, i, obj.GetName(), obj)
			dir := gfs.fsmkdir(Qobj, uint64(obj.GetID()))
			if owner := gfs.gs.Owners[obj.GetID()]; owner != nil {
				dir.Uid = owner.Name()
				dir.Gid = owner.Name()
			}
			return dir
		}

		return nil
	}

	objgen := func(off uint64, i interface{}) *p.Dir {
		off += Qobj + 1
		if off > Qobjy {
			return nil
		}

		dir := gfs.fsmkdir(off, uint64(fid.id))
		if owner := gfs.gs.Owners[fid.id]; owner != nil {
			if owner.Name() == fid.owner.Name() {
				dir.Uid = fid.owner.Name()
				dir.Gid = fid.owner.Name()
			}
		}
		return dir
	}

	glog.Infof("read fid %d level %d", fid.id, fid.level)

	switch fid.level {
	case Qroot:
		dirread9p(req, rootgen, nil)
		return
	case Qrctl, Qevt:
		if tc.Offset > 0 {
			req.RespondRread([]byte(nil))
			return
		}
		req.RespondRread([]byte("no"))
		return
	case Qzonedir:
		dirread9p(req, zonedirgen, nil)
		return
	case Qobjdir:
		dirread9p(req, objdirgen, gfs.gs.Objects.Chan())
		return
	case Qobj:
		dirread9p(req, objgen, nil)
		return
	case Qzone:
		dirread9p(req, zonegen, nil)
		return
	case Qoctl:
	case Qobjid:
		if tc.Offset == 0 {
			obj := gfs.gs.Objects.FindObjectByID(fid.id)
			if obj != nil {
				req.RespondRread([]byte(fmt.Sprintf("%d\n", obj.GetID())))
			} else {
				req.RespondError(srv.Enoent)
			}
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qobjname:
		if tc.Offset == 0 {
			obj := gfs.gs.Objects.FindObjectByID(fid.id)
			if obj != nil {
				req.RespondRread([]byte(fmt.Sprintf("%s\n", obj.GetName())))
			} else {
				req.RespondError(srv.Enoent)
			}
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qobjx:
		if tc.Offset == 0 {
			obj := gfs.gs.Objects.FindObjectByID(fid.id)
			if obj != nil {
				x, _ := obj.GetPos()
				req.RespondRread([]byte(fmt.Sprintf("%d\n", x)))
			} else {
				req.RespondError(srv.Enoent)
			}
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qobjy:
		if tc.Offset == 0 {
			obj := gfs.gs.Objects.FindObjectByID(fid.id)
			if obj != nil {
				_, y := obj.GetPos()
				req.RespondRread([]byte(fmt.Sprintf("%d\n", y)))
			} else {
				req.RespondError(srv.Enoent)
			}
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	}

	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Write(req *srv.Req) {
	fid := req.Fid.Aux.(*GameFid)
	p := gfs.fsmkdir(fid.level, uint64(fid.id))
	if owner := gfs.gs.Owners[fid.id]; owner != nil {
		if owner.Name() == fid.owner.Name() {
			p.Uid = fid.owner.Name()
			p.Gid = fid.owner.Name()
		}
	}
	fidowner := req.Fid.User
	glog.Infof("%s write to %s owned by %s", fidowner.Name(), p.Name, p.Uid)
	if fid.owner.Name() != p.Uid {
		req.RespondError(srv.Eperm)
		return
	}

	buf := bytes.NewBuffer(req.Tc.Data)
	sc := bufio.NewScanner(buf)
	for sc.Scan() {
		spl := strings.Fields(sc.Text())
		if len(spl) < 1 {
			continue
		}

		switch fid.level {
		case Qrctl:
		case Qoctl:
			obj := gfs.gs.Objects.FindObjectByID(fid.id)
			switch spl[0] {
			case "setpos":
				if len(spl) != 3 {
					continue
				}
				x, _ := strconv.Atoi(spl[1])
				y, _ := strconv.Atoi(spl[2])
				obj.SetPos(x, y)
			}

		}
	}
	req.RespondRwrite(uint32(len(req.Tc.Data) - buf.Len()))
	//req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Clunk(req *srv.Req) {
	req.RespondRclunk()
}

func (gfs *GameFs) Remove(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Stat(req *srv.Req) {
	fid := req.Fid.Aux.(*GameFid)
	p := gfs.fsmkdir(fid.level, uint64(fid.id))
	if fid.owner != nil {
		p.Uid = fid.owner.Name()
		p.Gid = fid.owner.Name()
	}

	req.RespondRstat(p)
}

func (gfs *GameFs) Wstat(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) ConnOpened(conn *srv.Conn) {
	if conn.Srv.Debuglevel > 0 {
		glog.Infof("connected %+v", conn)
	}
}

func (gfs *GameFs) ConnClosed(conn *srv.Conn) {
	delete(gfs.conns, conn)
	if conn.Srv.Debuglevel > 0 {
		glog.Infof("disconnected %+v", conn)
	}
}

func (*GameFs) FidDestroy(sfid *srv.Fid) {
}

func (gfs *GameFs) Run() error {
	gfs.Id = "gamefs"
	gfs.Debuglevel = 1
	gfs.Upool = gfs.gs.DBM
	gfs.Start(gfs)

	var listen string
	defaultlistenstr := "127.0.0.1:61507"
	if listenconf, err := gfs.gs.config.Get("listener", reflect.String); err != nil {
		glog.Info("'listener' not found in config. defaulting to ", defaultlistenstr)
		listen = defaultlistenstr
	} else {
		listen = listenconf.(string)
	}

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		glog.Error(err)
		return err
	}

	go func() {
		for {
			c, err := listener.Accept()
			if err != nil {
				glog.Error(err)
			}

			gfs.NewConn(c)
		}
	}()

	return nil //gfs.StartNetListener("tcp", ":61508")
}
