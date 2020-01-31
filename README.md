[![Documentation](https://godoc.org/github.com/phemmer/go-inotify?status.png)](http://godoc.org/github.com/phemmer/go-inotify)

Package inotify provides a no-frills, fully exposed, Linux inotify implementation in go.

There are several other packages available which already provide inotify functionality, but many of them provide
abstractions which restrict what you can do. This package is meant to fully expose inotify, allowing as much control
as desired. There are no hidden struct members, and the code is extremely simple.

In addition, the package does not utilize Cgo, and should thus work with the native go compiler, avoiding all the
little gotchas that Cgo brings.

```go
import "github.com/phemmer/go-inotify"

func main() {
  in, err := inotify.New()
  // handle error

  w, err := in.AddWatch("/path/to/foo", inotify.IN_CREATE|inotify.IN_DELETE)
  // handle error

  in.SetReadDeadline(time.Now().Add(time.Second))
  event, err := in.Read()
  // handle error
  // do something with event

  err := in.RemoveWatch(w)
  // moar error!

  in.Close()
}
```
