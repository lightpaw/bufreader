// It's like bufio.Reader, but it returns its internal buffer to minimize copies.
// It provides better performance than bufio.Reader when the caller only use the result temporarily like unmarshal to protobuf or write to another buffer.
package bufreader

import (
	"errors"
	"github.com/lightpaw/slab"
	"io"
)

var (
	slabPool                  = slab.NewSyncPool(128, 32768, 2)
	ErrBufReaderAlreadyClosed = errors.New("bufreader.Reader already closed")
)

type Reader struct {
	reader    io.Reader
	buf       []byte
	w         int
	r         int
	cleanedUp bool
}

func NewReader(r io.Reader, initialSize int) *Reader {
	buf := slabPool.Alloc(initialSize)
	return &Reader{reader: r, buf: buf}
}

func (r *Reader) ReadByte() (n byte, err error) {
	if r.unreadBytes() > 0 {
		n = r.buf[r.r]
		r.r++
		return
	}

	if r.capLeft() == 0 {
		if r.cleanedUp {
			return 0, ErrBufReaderAlreadyClosed
		}

		// both r and w is at final position
		r.r, r.w = 0, 0
	}

	// enough room to Read
	if err = r.readAtLeast(1); err != nil {
		return
	}
	n = r.buf[r.r]
	r.r++
	return
}

// return a slice with exactly n bytes. It's safe to use the result slice before the next call to any Read method.
func (r *Reader) ReadFull(n int) ([]byte, error) {
	unreadBytes := r.unreadBytes()
	if unreadBytes >= n {
		result := r.buf[r.r : r.r+n]
		r.r += n
		return result, nil
	}

	needToRead := n - unreadBytes
	if r.capLeft() >= needToRead {
		// enough room to Read
		if err := r.readAtLeast(needToRead); err != nil {
			return nil, err
		}

		result := r.buf[r.r : r.r+n]
		r.r += n
		return result, nil
	}

	// not enough room
	// check if buf is large enough
	if n > len(r.buf) {
		if cap(r.buf) == 0 {
			return nil, ErrBufReaderAlreadyClosed
		}

		// make a larger buf
		newBuf := slabPool.Alloc(n + 128)
		r.w = copy(newBuf, r.buf[r.r:r.w])
		r.r = 0
		slabPool.Free(r.buf)
		r.buf = newBuf
	} else {
		// enough room, shift existing data to left
		r.w = copy(r.buf, r.buf[r.r:r.w])
		r.r = 0
	}

	if err := r.readAtLeast(needToRead); err != nil {
		return nil, err
	}

	result := r.buf[r.r : r.r+n]
	r.r += n
	return result, nil
}

func (r *Reader) readAtLeast(bytes int) error {
	if n, err := io.ReadAtLeast(r.reader, r.buf[r.w:], bytes); err != nil {
		return err
	} else {
		r.w += n
		return nil
	}
}

func (r *Reader) unreadBytes() int {
	return r.w - r.r
}

func (r *Reader) capLeft() int {
	return len(r.buf) - r.w
}

func (r *Reader) Close() error {
	if r.cleanedUp {
		return ErrBufReaderAlreadyClosed
	}
	r.cleanedUp = true
	slabPool.Free(r.buf)
	r.w, r.r = 0, 0
	r.buf = nil
	return nil
}
