package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"unicode/utf8"
	//"time"

	"github.com/golang/glog"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"

	"github.com/nsf/termbox-go"

	"code.google.com/p/go9p/p"

	"github.com/mischief/goland/game/gfx"
	"github.com/mischief/goland/game/gutil"
)

type gorplog struct {
}

func (gorplog) Printf(format string, v ...interface{}) {
	glog.Infof(format, v...)
}

type DBManager struct {
	*gorp.DbMap
}

func NewDBManager(gs *GameServer) *DBManager {
	var dbdriver, dbspec string
	if driverconf, err := gs.config.Get("db.driver", reflect.String); err != nil {
		glog.Fatalf("missing db.driver in config")
	} else {
		dbdriver = driverconf.(string)
	}

	if specconf, err := gs.config.Get("db.spec", reflect.String); err != nil {
		glog.Fatalf("missing db.spec in config")
	} else {
		dbspec = specconf.(string)
	}

	db, err := sql.Open(dbdriver, dbspec)
	if err != nil {
		panic(err)
	}

	dbm := &DBManager{
		&gorp.DbMap{Db: db},
	}

	switch dbdriver {
	case "sqlite3":
		dbm.Dialect = gorp.SqliteDialect{}
	case "postgres":
		dbm.Dialect = gorp.PostgresDialect{}
	case "mysql":
		dbm.Dialect = gorp.MySQLDialect{}
	}

	dbm.TraceOn("[gorp]", gorplog{})

	dbm.init()

	return dbm
}

func (dbm *DBManager) init() {
	dbm.AddTableWithName(User{}, "users").SetKeys(true, "Dbid")
	dbm.AddTableWithName(Member{}, "members").SetKeys(true, "Id")
	dbm.AddTableWithName(Sprite{}, "sprites").SetKeys(true, "Id")
	dbm.AddTableWithName(ItemProto{}, "itemprotos").SetKeys(true, "Id")
	dbm.AddTableWithName(Item{}, "items").SetKeys(true, "Id")
	dbm.AddTableWithName(Zone{}, "zones").SetKeys(true, "Id")

	if err := dbm.SelectOne(&User{}, "select * from users where uid=-1 limit 1"); err == nil {
		return
	}

	dbm.CreateTables()

	users := []User{
		User{Uid: -1, Username: "adm", Password: ""},
		User{Uid: 0, Username: "none", Password: ""},
		User{Uid: 1, Username: "tor", Password: ""},
		User{Uid: 10000, Username: "sys", Password: ""},
		User{Uid: 10001, Username: "glenda", Password: "noisebridge"},
		User{Uid: 61507, Username: "mischief", Password: "noisebridge"},
	}

	memberships := []Member{
		Member{Uid: -1, Mid: 0},
		Member{Uid: 1, Mid: 1},
		Member{Uid: 10000, Mid: 10000},
		Member{Uid: 61507, Mid: 61507},
		Member{Uid: 10000, Mid: 61507},
	}

	sprites := []Sprite{
		ToSprite("void"),
		ToSprite("floor"),
		ToSprite("wall"),
		//ToSprite("door"),
		ToSprite("flag"),
		ToSprite("human"),
	}

	for _, u := range users {
		if err := dbm.Insert(&u); err != nil {
			panic(err)
		}
	}

	for _, m := range memberships {
		if err := dbm.Insert(&m); err != nil {
			panic(err)
		}
	}

	for _, s := range sprites {
		if err := dbm.Insert(&s); err != nil {
			panic(err)
		}
	}

	var flg Sprite
	if err := dbm.SelectOne(&flg, "select * from sprites where name='flag' limit 1"); err != nil {
		panic(err)
	}

	itemprotos := []ItemProto{
		ItemProto{Name: "flag", SpriteId: flg.Id},
	}

	for _, i := range itemprotos {
		if err := dbm.Insert(&i); err != nil {
			panic(err)
		}
	}

}

func (dbm *DBManager) Uid2User(uid int) p.User {
	var u User
	if err := dbm.SelectOne(&u, "select * from users where uid=? limit 1", uid); err != nil {
		glog.Error(err)
		return nil
	}
	return &u
}

func (dbm *DBManager) Uname2User(uname string) p.User {
	var u User
	if err := dbm.SelectOne(&u, "select * from users where username=? limit 1", uname); err != nil {
		glog.Error(err)
		return nil
	}

	u.dbm = dbm
	return &u
}

func (dbm *DBManager) Gid2Group(gid int) p.Group {
	var u User
	if err := dbm.SelectOne(&u, "select * from users where uid=? limit 1", gid); err != nil {
		glog.Error(err)
		return nil
	}

	u.dbm = dbm
	return &u
}

func (dbm *DBManager) Gname2Group(gname string) p.Group {
	var u User
	if err := dbm.SelectOne(&u, "select * from users where username=? limit 1", gname); err != nil {
		glog.Error(err)
		return nil
	}
	u.dbm = dbm
	return &u
}

type User struct {
	dbm *DBManager `db:"-"`

	Dbid     int    `db:"id"`
	Uid      int    `db:"uid"`
	Username string `db:"username"`
	//	Deskey   string `db:"deskey"`
	Password string `db:"password"`
}

type Member struct {
	Id int `db:"id"`
	// the uid we are a group of
	Uid int `db:"uid"`
	// the member of the group
	Mid int `db:"mid"`
}

func (u User) String() string { return fmt.Sprintf("%s %d", u.Username, u.Uid) }
func (u *User) Name() string  { return u.Username }
func (u *User) Id() int       { return u.Uid }
func (u *User) Groups() []p.Group {
	res, err := u.dbm.Select(User{}, "select users.* from users, (select id as id, uid as uid from members where mid = ?) as tmp1 where users.uid = tmp1.uid order by tmp1.id;", u.Uid)
	//res, err := u.dbm.Select(User{}, "select * from users where uid in (select uid from members where mid = ?)", u.Uid)
	if err != nil {
		glog.Error(err)
		return nil
	}

	groups := make([]p.Group, len(res))
	for i, r := range res {
		groups[i] = r.(p.Group)
	}
	return groups
}

func (u *User) Members() []p.User {
	res, err := u.dbm.Select(User{}, "select * from users where uid in (select mid from members where uid = ?) order by id", u.Uid)
	if err != nil {
		glog.Error(err)
		return nil
	}
	membs := make([]p.User, len(res))
	for i, r := range res {
		us := r.(*User)
		us.dbm = u.dbm
		membs[i] = us
	}
	return membs
}

func (u *User) IsMember(g p.Group) bool {
	for _, og := range u.Groups() {
		if g.Id() == og.Id() {
			return true
		}
	}
	return false
}

type Sprite struct {
	Id      int    `db:"id"`
	Type    int    `db:"type"`
	Name    string `db:"name"`
	Frames  string `db:"frames"`
	FgColor string `db:"fgcolor"`
	BgColor string `db:"bgcolor"`
}

func ToSprite(name string) Sprite {
	sp := gfx.Get(name)
	ss := sp.(*gfx.StaticSprite)
	cell := ss.Cell()
	return Sprite{
		Type:    0,
		Name:    name,
		Frames:  string(cell.Ch),
		FgColor: gutil.TermboxAttrToStr(cell.Fg),
		BgColor: gutil.TermboxAttrToStr(cell.Bg),
	}
}

func (dbm *DBManager) GetSprite(sname string) gfx.Sprite {
	var s Sprite
	if err := dbm.SelectOne(&s, "select * from sprites where name=? limit 1", sname); err != nil {
		glog.Error(err)
		return nil
	}
	if s.Type == 0 {
		r, _ := utf8.DecodeRuneInString(s.Frames)
		cell := termbox.Cell{
			Ch: r,
			Fg: gutil.StrToTermboxAttr(s.FgColor),
			Bg: gutil.StrToTermboxAttr(s.BgColor),
		}
		ss := gfx.NewStaticSprite(cell)
		return ss
	}
	return nil
}

// ItemProto is the prototype of items.
type ItemProto struct {
	Id       int    `db:"id"`
	Class    int    `db:"class"`
	Name     string `db:"name"`
	Damage   int    `db:"damage"`
	Armor    int    `db:"armor"`
	SpriteId int    `db:"spriteid"`
}

func (dbm *DBManager) GetItemProto(iname string) *ItemProto {
	var ip ItemProto
	if err := dbm.SelectOne(&ip, "select * from itemprotos where name=? limit 1", iname); err != nil {
		glog.Error(err)
		return nil
	}
	return &ip
}

// Item is the table of actual instances of items.
type Item struct {
	Id      int `db:"id"`
	ProtoId int `db:"protoid"`
	OwnerId int `db:"ownerid"`
}

type Zone struct {
	Id      int    `db:"id"`
	Name    string `db:"name"`
	Display string `db:"display"`
	Map     string `db:"map"`
	Code    string `db:"code"`
}

func (dbm *DBManager) GetZone(zname string) *Zone {
	var z Zone
	if err := dbm.SelectOne(&z, "select * from zones where name=? limit 1", zname); err != nil {
		glog.Error(err)
		return nil
	}
	return &z
}
