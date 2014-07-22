package gauth

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	AuthTreq   = 1  /* ticket request */
	AuthChal   = 2  /* challenge box request */
	AuthPass   = 3  /* change password */
	AuthOK     = 4  /* fixed length reply follows */
	AuthErr    = 5  /* error follows */
	AuthMod    = 6  /* modify user */
	AuthApop   = 7  /* apop authentication for pop3 */
	AuthOKvar  = 9  /* variable length reply follows */
	AuthChap   = 10 /* chap authentication for ppp */
	AuthMSchap = 11 /* MS chap authentication for ppp */
	AuthCram   = 12 /* CRAM verification for IMAP (RFC2195 & rfc2104) */
	AuthHttp   = 13 /* http domain login */
	AuthVNC    = 14 /* VNC server login (deprecated) */

	AuthTs = 64 /* ticket encrypted with server's key */
	AuthTc = 65 /* ticket encrypted with client's key */
	AuthAs = 66 /* server generated authenticator */
	AuthAc = 67 /* client generated authenticator */
	AuthTp = 68 /* ticket encrypted with client's key for password change */
	AuthHr = 69 /* http reply */
)

var (
	typenametab = map[int]string{
		AuthTreq: "AuthTreq",
		AuthOK:   "AuthOK",
		AuthErr:  "AuthErr",
	}
)

type Ticketreq struct {
	Type    uint8
	Authid  [ANAMELEN]byte
	Authdom [DOMLEN]byte
	Chal    [CHALLEN]byte
	Hostid  [ANAMELEN]byte
	Uid     [ANAMELEN]byte
}

func btos(b []byte) string {
	if i := bytes.IndexByte(b, byte(0)); i != -1 {
		return string(b[:i])
	}
	return string(b[:])
}

func (t Ticketreq) String() string {
	typ := typenametab[int(t.Type)]
	if typ == "" {
		typ = "unknown"
	}
	return typ + " " + btos(t.Authid[:]) + "@" + btos(t.Authdom[:]) +
		" chal " + fmt.Sprintf("%x", t.Chal[:]) +
		" hostid " + btos(t.Hostid[:]) +
		" uid " + btos(t.Uid[:])
}

type Ticket struct {
	Num  uint8
	Chal [CHALLEN]byte
	Cuid [ANAMELEN]byte
	Suid [ANAMELEN]byte
	Key  [DESKEYLEN]byte
}

func (t *Ticket) ToM(key []byte) []byte {
	buf := new(bytes.Buffer)

	copy(t.Key[:], key[:7])
	binary.Write(buf, binary.LittleEndian, t)

	if key != nil {
		DesEncrypt(key, buf.Bytes())
	}

	return buf.Bytes()
}
