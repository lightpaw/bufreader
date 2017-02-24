# buf reader [![Build Status](https://travis-ci.org/lightpaw/bufreader.svg?branch=master)](https://travis-ci.org/lightpaw/bufreader) [![codecov](https://codecov.io/gh/lightpaw/bufreader/branch/master/graph/badge.svg)](https://codecov.io/gh/lightpaw/bufreader)

It's like `bufio.Reader` that reads as much bytes as possible in one read call to reduce the number of `Read` calls to underlying `net.Conn` which results in system calls. 
 
But it returns its internal buffer to the caller to reduce memory copies. Most of the times, the caller unmarshal the slice to something else (like protobuf or just make an int out of 4 bytes). 
 
Internally it uses `sync.Pool` to further reduce gc.
 
```go
    reader := bufreader.NewBufReader(conn, 1024) // use the reader to read the content in conn
    length, err := reader.ReadFull(2) // read 2 bytes for length
    data, err := reader.ReadFull(int(binary.LittleEndian.Uint16(length))) // read data
```
