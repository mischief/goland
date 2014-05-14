package gauth

// PassToKey converts the string pass into
// a 56-bit Plan 9-style DES key.
func PassToKey(pass string) []byte {
	var t, i, n uint

	buf := make([]byte, ANAMELEN)
	key := make([]byte, DESKEYLEN)

	p := []byte(pass)
	n = uint(len(p))

	if n >= ANAMELEN {
		n = ANAMELEN - 1
	}

	copy(buf, []byte("        "))

	copy(buf, p)
	buf[n] = 0

	for {
		for i = uint(0); i < DESKEYLEN; i++ {
			key[i] = (buf[t+i] >> i) + (buf[t+i+1] << (8 - (i + 1)))
		}
		if n <= 8 {
			return key
		}
		n -= 8
		t += 8
		if n < 8 {
			t -= 8 - n
			n = 8
		}
		DesEncrypt(key, buf[t:t+8])
	}
}
