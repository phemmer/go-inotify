// Package inotify provides a no-frills, fully exposed, Linux inotify implementation in go.
//
// There are several other packages available which already provide inotify functionality, but many of them provide
// abstractions which restrict what you can do. This package is meant to fully expose inotify, allowing as much control
// as desired. There are no hidden struct members, and the code is extremely simple.
//
// In addition, the package does not utilize Cgo, and should thus work with the native go compiler, avoiding all the
// little gotchas that Cgo brings.
package inotify

import (
	"bytes"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Mask = uint32
type Event struct {
	// type InotifyEvent struct {
	//    Wd     int32
	//    Mask   uint32
	//    Cookie uint32
	//    Len    uint32
	// }
	unix.InotifyEvent
	Name string
}

// These are duplicated locally and type converted for convenience.
const (
	IN_ACCESS        Mask = unix.IN_ACCESS
	IN_ATTRIB        Mask = unix.IN_ATTRIB
	IN_CLOSE_WRITE   Mask = unix.IN_CLOSE_WRITE
	IN_CLOSE_NOWRITE Mask = unix.IN_CLOSE_NOWRITE
	IN_CREATE        Mask = unix.IN_CREATE
	IN_DELETE        Mask = unix.IN_DELETE
	IN_DELETE_SELF   Mask = unix.IN_DELETE_SELF
	IN_MODIFY        Mask = unix.IN_MODIFY
	IN_MOVE_SELF     Mask = unix.IN_MOVE_SELF
	IN_MOVED_FROM    Mask = unix.IN_MOVED_FROM
	IN_MOVED_TO      Mask = unix.IN_MOVED_TO
	IN_OPEN          Mask = unix.IN_OPEN
	IN_MOVE          Mask = unix.IN_MOVE
	IN_CLOSE         Mask = unix.IN_CLOSE
	IN_ALL_EVENTS    Mask = unix.IN_ALL_EVENTS

	IN_DONT_FOLLOW Mask = unix.IN_DONT_FOLLOW
	IN_MASK_ADD    Mask = unix.IN_MASK_ADD
	IN_ONESHOT     Mask = unix.IN_ONESHOT
	IN_ONLYDIR     Mask = unix.IN_ONLYDIR
	IN_MASK_CREATE Mask = unix.IN_MASK_CREATE

	IN_IGNORED    uint32 = unix.IN_IGNORED
	IN_ISDIR      uint32 = unix.IN_ISDIR
	IN_Q_OVERFLOW uint32 = unix.IN_Q_OVERFLOW
	IN_UNMOUNT    uint32 = unix.IN_UNMOUNT
)

type InotifyWatchDesc int32
type Inotify struct {
	file    *os.File
	rawConn syscall.RawConn
	Buffer  *bytes.Buffer
}

// New constructs a new Inotify watcher.
func New() (*Inotify, error) {
	fd, err := syscall.InotifyInit1(unix.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), "")
	rc, err := f.SyscallConn()
	if err != nil {
		return nil, err
	}
	_ = rc.Control(func(fd uintptr) {
		syscall.CloseOnExec(int(fd))
	})
	return &Inotify{
		file:    f,
		rawConn: rc,
		Buffer:  bytes.NewBuffer(make([]byte, 0, 4096)),
	}, nil
}

func (in *Inotify) Close() error {
	return in.file.Close()
}

// We have to use syscall.RawConn to obtain the file descriptor because file.FD() sets the file descriptor to blocking
// mode. We also can't permanently store the FD as if the file is closed, we won't detect it.
func (in *Inotify) fd() (int, error) {
	var fd uintptr
	err := in.rawConn.Control(func(v uintptr) { fd = v })
	return int(fd), err
}

// AddWatch adds the given path to the watch list and returns a new watch descriptor.
// The watch descriptor can be provided to RemoveWatch() to stop watching.
func (in *Inotify) AddWatch(pathname string, mask Mask) (InotifyWatchDesc, error) {
	fd, err := in.fd()
	if err != nil {
		return 0, err
	}
	wd, err := syscall.InotifyAddWatch(fd, pathname, mask)
	return InotifyWatchDesc(wd), err
}

// Removes the given watch descriptor from the watch list.
func (in *Inotify) RemoveWatch(watchDesc InotifyWatchDesc) error {
	fd, err := in.fd()
	if err != nil {
		return err
	}
	_, err = syscall.InotifyRmWatch(fd, uint32(watchDesc))
	return err
}

// Read reads the next event, blocking if none is available.
func (in *Inotify) Read() (Event, error) {
	if in.Buffer.Len() == 0 {
		in.Buffer.Reset()
		buf := in.Buffer.Bytes()[:in.Buffer.Cap()]
		n, err := in.file.Read(buf)
		if err != nil {
			return Event{}, err
		}
		// Yes, we're writing the contents of the buffer to itself. It just calls copy() under the hood, so it should no-op.
		// This is needed to get the buffer to resize its slice, and be able to provide the content.
		in.Buffer.Write(buf[:n])
	}

	buf := in.Buffer.Next(unix.SizeofInotifyEvent)
	event := Event{}
	event.InotifyEvent = *(*unix.InotifyEvent)(unsafe.Pointer(&buf[0]))
	event.Name = string(bytes.SplitN(in.Buffer.Next(int(event.InotifyEvent.Len)), []byte{0}, 2)[0])
	return event, nil
}
