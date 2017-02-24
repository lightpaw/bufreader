package bufreader

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestBufReader(t *testing.T) {
	RegisterTestingT(t)
	r := &numReader{}

	bufReader := NewReader(r, 10)

	buf, err := bufReader.ReadFull(5)

	Ω(err).ShouldNot(HaveOccurred())
	Ω(buf).Should(HaveLen(5))
	assertSame(buf, err, 0, 5)

	buf, err = bufReader.ReadFull(6)
	assertSame(buf, err, 5, 11)

	buf, err = bufReader.ReadFull(2)
	assertSame(buf, err, 11, 13)

	buf, err = bufReader.ReadFull(200)
	assertSame(buf, err, 13, 213)
}

func TestBufReader_ReadByte(t *testing.T) {
	RegisterTestingT(t)
	r := &numReader{}

	bufReader := NewReader(r, 0)

	Ω(bufReader.buf).Should(HaveCap(128))
	Ω(bufReader.buf).Should(HaveLen(128))

	for i := 0; i < 999999; i++ {
		num, err := bufReader.ReadByte()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(byte(i)).Should(Equal(num))
	}
}

func TestBufReader_CleanUp(t *testing.T) {
	RegisterTestingT(t)
	r := &numReader{}

	bufReader := NewReader(r, 0)
	bufReader.readAtLeast(100)
	bufReader.Close()
	Ω(bufReader.unreadBytes()).Should(Equal(0))
	Ω(bufReader.capLeft()).Should(Equal(0))

	_, err := bufReader.ReadByte()
	Ω(err).Should(HaveOccurred())
	_, err = bufReader.ReadFull(2)
	Ω(err).Should(HaveOccurred())

	bufReader.Close() // call multiple times
}

func assertSame(b []byte, err error, start, end int) {
	Ω(err).ShouldNot(HaveOccurred())
	Ω(b).Should(HaveLen(end - start))
	for i := 0; i < len(b); i++ {
		Ω(b[i]).Should(Equal(byte(start + i)))
	}
}

type numReader struct {
	next int
}

func (r *numReader) Read(p []byte) (int, error) {
	for i := 0; i < len(p); i++ {
		p[i] = byte(r.next)
		r.next++
	}

	return len(p), nil
}
