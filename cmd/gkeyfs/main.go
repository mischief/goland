package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/srv"

	"github.com/golang/glog"

	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"

	"github.com/mischief/goland/game/gauth"
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
	Quser
	Qkey
	Qsecret
	Qlog
	Qstatus
	Qexpire
	Qwarnings
	Qmax
)

type KeyfsUser string

func (u *KeyfsUser) Name() string            { return string(*u) }
func (u *KeyfsUser) Id() int                 { return 61508 }
func (u *KeyfsUser) Groups() []p.Group       { return []p.Group{u} }
func (u *KeyfsUser) IsMember(g p.Group) bool { return true }
func (u *KeyfsUser) Members() []p.User       { return []p.User{u} }

type FakeUsers struct {
}

func (f *FakeUsers) Uid2User(uid int) p.User {
	return &keyfsuser
}

func (f *FakeUsers) Uname2User(uname string) p.User {
	return &keyfsuser
}

func (f *FakeUsers) Gid2Group(gid int) p.Group {
	return &keyfsuser
}

func (f *FakeUsers) Gname2Group(gname string) p.Group {
	return &keyfsuser
}

var (
	keyfsuser = KeyfsUser("auth")

	nametab = map[int]string{
		Qroot:     "/",
		Quser:     "userdir",
		Qkey:      "key",
		Qsecret:   "secret",
		Qlog:      "log",
		Qstatus:   "status",
		Qexpire:   "expire",
		Qwarnings: "warnings",
	}

	flagl  = flag.String("l", ":61509", "listen address")
	flagdb = flag.String("db", "", "database file")
)

func fsmkqid(level, id uint64) *p.Qid {
	var q p.Qid

	q.Type = p.QTFILE
	q.Version = 0

	switch level {
	case Qroot, Quser:
		q.Type = p.QTDIR
		fallthrough
	case Qkey, Qsecret, Qlog, Qstatus, Qexpire, Qwarnings:
		fallthrough
	default:
		q.Path = uint64((level << 17) | id)
	}
	return &q
}

type KeyFid struct {
	qtype uint64
	uid   int
	qid   *p.Qid
}

type KeyFs struct {
	srv.Srv
	dbm *gorp.DbMap

	quit chan struct{}
}

func (k *KeyFs) fsmkdir(qtype, id uint64, name string) *p.Dir {
	var d p.Dir

	d.Uid = keyfsuser.Name()
	d.Uidnum = uint32(keyfsuser.Id())
	d.Gid = keyfsuser.Name()
	d.Gidnum = uint32(keyfsuser.Id())

	d.Qid = *fsmkqid(qtype, id)

	d.Mode = 0666

	if d.Qid.Type&p.QTDIR > 0 {
		d.Mode |= p.DMDIR | 0111
	}

	d.Name = nametab[int(qtype)]

	switch qtype {
	case Quser:
		d.Name = fmt.Sprintf("%s", name)
	}

	return &d
}

func (k *KeyFs) Attach(req *srv.Req) {
	fid := &KeyFid{
		qtype: Qroot,
		uid:   0,
		qid:   fsmkqid(Qroot, 0),
	}

	req.Fid.Aux = fid

	req.RespondRattach(fsmkqid(Qroot, 0))
}

func (k *KeyFs) Flush(req *srv.Req) {
}

func (k *KeyFs) Walk1(fid *srv.Fid, name string, qid *p.Qid) (err error) {
	var i, id int

	kfid := fid.Aux.(*KeyFid)

	glog.Infof("%+v", kfid)
	if kfid.qid.Type&p.QTDIR == 1 {
		return fmt.Errorf("walk in non-directory")
	}

	glog.Infof("walk1 %s %d %s", nametab[int(kfid.qtype)], len(name), name)
	if name == ".." {
		switch kfid.qtype {
		case Qroot:
			break
		default:
			kfid.qtype = Qroot
		}
	} else {

	loop:
		for i = int(kfid.qtype + 1); i < len(nametab); i++ {
			if nametab[i] != "" {
				if nametab[i] == name {
					id = int(kfid.uid)
					glog.Infof("matched %s (%d)", name, id)
					break loop
				}
			}

			if i == Quser {
				var u User
				if err := k.dbm.SelectOne(&u, "select * from users where username=? limit 1", name); err != nil {
					glog.Error(err)
					return err
				}
				id = u.Uid
				break loop
			}
		}
		if i >= len(nametab) {
			return fmt.Errorf("directory entry not found")
		}
		kfid.qtype = uint64(i)
	}

	q := fsmkqid(kfid.qtype, uint64(id))
	*qid = *q
	kfid.qid = q
	kfid.uid = id
	glog.Infof("level %d uid %d qid %s", kfid.qtype, kfid.uid, q)
	return nil
}

func (k *KeyFs) Clone(oldf, newf *srv.Fid) error {
	newfid := new(KeyFid)
	*newfid = *oldf.Aux.(*KeyFid)
	newf.Aux = newfid
	return nil
}

func (k *KeyFs) Walk(req *srv.Req) {
	if req.Fid != req.Newfid {
		if err := k.Clone(req.Fid, req.Newfid); err != nil {
			req.RespondError(err)
			return
		}
	}

	var err error

	wqids := make([]p.Qid, len(req.Tc.Wname))
	nqids := 0
	for i := 0; i < len(req.Tc.Wname); i++ {
		var q p.Qid
		err = k.Walk1(req.Newfid, strings.TrimRight(req.Tc.Wname[i], "\x00"), &q)

		if err != nil {
			break
		}

		newf := req.Newfid.Aux.(*KeyFid)
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
}

func (k *KeyFs) Open(req *srv.Req) {
	fid := req.Fid.Aux.(*KeyFid)
	d := fsmkqid(fid.qtype, uint64(fid.uid))
	glog.Infof("open %d %q %d", fid.qtype, fid.qid, fid.uid)
	fid.qid = d
	req.RespondRopen(d, 8192)
	return
}

func (k *KeyFs) Create(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (k *KeyFs) Read(req *srv.Req) {
	fid := req.Fid.Aux.(*KeyFid)
	tc := req.Tc

	rootgen := func(off uint64, i interface{}) *p.Dir {
		users := i.([]User)

		if off+1 > uint64(len(users)) {
			return nil
		}
		//glog.Infof("%d %v %s %v", off, i, obj.GetName(), obj)
		dir := k.fsmkdir(Quser, uint64(users[off].Dbid), users[off].Username)
		return dir
	}

	keydirgen := func(off uint64, i interface{}) *p.Dir {
		off += Quser + 1
		if off >= Qmax {
			return nil
		}
		var u User
		if err := k.dbm.SelectOne(&u, "select * from users where uid=? limit 1", fid.uid); err != nil {
			glog.Error(err)
			return nil
		}
		dir := k.fsmkdir(off, uint64(fid.uid), u.Username)
		return dir
	}

	glog.Infof("read fid qtype %d uid %d", fid.qtype, fid.uid)

	switch fid.qtype {
	case Qroot:
		// select bla bla
		var users []User
		if _, err := k.dbm.Select(&users, "select * from users order by id"); err != nil {
			req.RespondError(err)
			return
		}
		dirread9p(req, rootgen, users)
		return
	case Quser:
		dirread9p(req, keydirgen, nil)
		return
	case Qkey:
		if tc.Offset == 0 {
			var u User
			if err := k.dbm.SelectOne(&u, "select * from users where uid=? limit 1", fid.uid); err != nil {
				glog.Error(err)
				req.RespondError(err)
				return
			}
			deskey := gauth.PassToKey(u.Password)
			req.RespondRread(deskey)
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qsecret:
		if tc.Offset == 0 {
			var u User
			if err := k.dbm.SelectOne(&u, "select * from users where uid=? limit 1", fid.uid); err != nil {
				glog.Error(err)
				req.RespondError(err)
				return
			}
			req.RespondRread([]byte(u.Password))
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qlog:
		if tc.Offset == 0 {
			req.RespondRread([]byte("0\n"))
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qstatus:
		if tc.Offset == 0 {
			req.RespondRread([]byte("ok\n"))
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qexpire:
		if tc.Offset == 0 {
			req.RespondRread([]byte("never\n"))
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	case Qwarnings:
		if tc.Offset == 0 {
			req.RespondRread([]byte("0\n"))
		} else {
			req.RespondRread([]byte(nil))
		}
		return
	}

	req.RespondError(srv.Enotimpl)
}

func (k *KeyFs) Write(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (k *KeyFs) Clunk(req *srv.Req) {
	req.RespondRclunk()
}

func (k *KeyFs) Remove(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (k *KeyFs) Stat(req *srv.Req) {
	var u User

	fid := req.Fid.Aux.(*KeyFid)

	if fid.uid != 0 {
		if err := k.dbm.SelectOne(&u, "select * from users where uid=? limit 1", fid.uid); err != nil {
			glog.Error(err)
			req.RespondError(err)
			return
		}
	}

	p := k.fsmkdir(fid.qtype, uint64(fid.uid), u.Username)

	req.RespondRstat(p)
}

func (k *KeyFs) Wstat(req *srv.Req) {
	req.RespondError(srv.Enotimpl)
}

func (k *KeyFs) Run() error {
	k.Id = "gkeyfs"
	k.Upool = &FakeUsers{}
	k.Debuglevel = 1
	k.Start(k)

	listener, err := net.Listen("tcp", *flagl)
	if err != nil {
		glog.Error(err)
		return err
	}

	go func() {
	loop:
		for {
			select {
			case <-k.quit:
				break loop
			default:
				c, err := listener.Accept()
				if err != nil {
					glog.Error(err)
				}

				glog.Infof("accept %s", c.RemoteAddr())
				k.NewConn(c)
			}
		}
		k.quit <- struct{}{}
	}()

	return nil
}

type gorplog struct {
}

func (gorplog) Printf(format string, v ...interface{}) {
	glog.Infof(format, v...)
}

type User struct {
	Dbid     int    `db:"id"`
	Uid      int    `db:"uid"`
	Username string `db:"username"`
	//Deskey   string `db:"deskey"`
	Password string `db:"password"`
}

func db(dbname string) (*gorp.DbMap, error) {
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		return nil, err
	}

	dbm := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbm.TraceOn("[gorp]", gorplog{})

	dbm.AddTableWithName(User{}, "users").SetKeys(true, "Dbid")
	if err := dbm.CreateTablesIfNotExists(); err != nil {
		return nil, err
	}
	return dbm, nil
}

func main() {
	flag.Parse()
	dbm, err := db(*flagdb)
	if err != nil {
		glog.Errorf("db: %s", err)
		return
	}
	defer dbm.Db.Close()

	keyfs := &KeyFs{
		dbm:  dbm,
		quit: make(chan struct{}, 1),
	}

	keyfs.Run()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	select {
	case <-sig:
		keyfs.quit <- struct{}{}
	}
	<-keyfs.quit
}
