package main

import (
	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/srv"
	"fmt"
	"github.com/mischief/goland/game/gobj"
	"time"
	//"strconv"
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

	req.Fid.Diroffset = start

	p.SetRreadCount(req.Rc, uint32(count))
	req.Respond()
}

const (
	Qroot = iota
	Qctl
	Qevt
	Qobjdir
	Qobj
	Qobjid
)

var nametab = map[int]string{
	Qroot:   "/",
	Qctl:    "ctl",
	Qevt:    "event",
	Qobjdir: "obj",
	/* skip Qobj */
	Qobjid: "id",
}

func fsmkqid(level, id uint64) *p.Qid {
	var q p.Qid

	q.Type = p.QTFILE
	q.Version = 0

	switch level {
	case Qobjdir:
		fallthrough
	case Qroot:
		q.Type = p.QTDIR
	case Qctl:
		fallthrough
	case Qevt:
		fallthrough
	case Qobj:
		q.Path = uint64(level << 24)
	case Qobjid:
		q.Path = uint64((level << 24) | id)
	}

	return &q
}

func fsmkdir(level, id uint64) *p.Dir {
	var d p.Dir

	d.Qid = *fsmkqid(level, id)
	d.Mode = 0444
	d.Uid = "sys"
	d.Gid = "sys"
	d.Muid = "sys"

	now := uint32(time.Now().Unix())

	d.Atime = now
	d.Mtime = now

	if d.Qid.Type&p.QTDIR > 0 {
		d.Mode |= p.DMDIR | 0111
	}

	d.Name = nametab[int(level)]

	switch level {
	case Qctl:
		d.Mode = 0666
	case Qobj:
		d.Name = fmt.Sprintf("%d", id)
	}

	return &d
}

type GameFid struct {
	level uint64
	qid   *p.Qid
}

type GameFs struct {
	srv.Srv
	gs *GameServer
}

func NewGameFs(gs *GameServer) *GameFs {
	gfs := &GameFs{
		gs: gs,
	}

	return gfs
}

func (gfs *GameFs) Attach(req *srv.Req) {
	if req.Afid != nil {
		req.RespondError(srv.Enoauth)
		return
	}

	if len(req.Tc.Aname) > 0 {
		req.RespondError(srv.Enotimpl)
		return
	}

	fid := new(GameFid)
	fid.level = Qroot
	fid.qid = fsmkqid(Qroot, 0)

	req.Fid.Aux = fid

	req.RespondRattach(fid.qid)
}

func (gfs *GameFs) Flush(req *srv.Req) {}

func (gfs *GameFs) Walk1(fid *srv.Fid, name string, qid *p.Qid) error {
	gfid := fid.Aux.(*GameFid)

	if name == ".." {
		switch gfid.level {
		case Qroot, Qctl, Qevt, Qobjdir:
			break
		default:
			return fmt.Errorf("not implemented: %s", name)
		}
	} else {
		var i int
		for i = 0; i < len(nametab); i++ {
			if nametab[i] != "" {
				if nametab[i] == name {
					break
				}
				if i == Qobj {
					//id, err := strconv.Atoi(name)
					//if err != nil {
					return fmt.Errorf("bad id: %s", name)
					//}
				}
			}
		}
		if i >= len(nametab) {
			return fmt.Errorf("directory entry not found")
		}
		gfid.level = uint64(i)
	}

	q := fsmkqid(gfid.level, 0)
	qid = q
	gfid.qid = q
	return nil
}

func (gfs *GameFs) Clone(oldf, newf *srv.Fid) error {
	newfid := new(GameFid)

	*newfid = *oldf.Aux.(*GameFid)

	newf.Aux = newfid
	return nil
}

func (gfs *GameFs) Walk(req *srv.Req) {
	if req.Fid != req.Newfid {
		oldf := req.Fid.Aux.(*GameFid)
		req.Newfid.Aux = oldf.qid
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
		req.RespondRwalk(wqids[:nqids])
		return
	}

	req.RespondError("Walk bug")
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
	if d := fsmkqid(fid.level, 0); d != nil {
		req.RespondRopen(d, 8192)
		return
	}
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Create(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Read(req *srv.Req) {
	fid := req.Fid.Aux.(*GameFid)
	tc := req.Tc

	rootgen := func(off uint64, i interface{}) *p.Dir {
		off += Qroot + 1

		if off < Qobj {
			return fsmkdir(off, 0)
		}
		//off -= Qobj
		return nil
	}

	objgen := func(off uint64, i interface{}) *p.Dir {
		id := int(off)
		//off += Qobj + 1
		objs := i.([]gobj.Object)

		if id < len(objs) {
			return fsmkdir(Qobj, uint64(objs[id].GetID()))
		}

		return nil
	}

	switch fid.level {
	case Qroot:
		dirread9p(req, rootgen, nil)
		return
	case Qobjdir:
		dirread9p(req, objgen, gfs.gs.Objects.GetSlice())
		return
	case Qctl, Qevt:
		if tc.Offset > 0 {
			req.RespondRread([]byte(nil))
			return
		}
		req.RespondRread([]byte("no"))
		return
	}

	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Write(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Clunk(req *srv.Req) {
	req.RespondRclunk()
}

func (gfs *GameFs) Remove(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Stat(req *srv.Req) {
	fid := req.Fid.Aux.(*GameFid)
	if d := fsmkdir(fid.level, 0); d != nil {
		req.RespondRstat(d)
		return
	}
	//req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Wstat(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (gfs *GameFs) Run() error {
	gfs.Id = "gamefs"
	gfs.Debuglevel = 1
	gfs.Start(gfs)

	return gfs.StartNetListener("tcp", ":61508")
}
