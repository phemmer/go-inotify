package inotify_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/phemmer/go-inotify"
)

func Example() {
	in, err := inotify.New()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	w, err := in.AddWatch("/tmp", inotify.IN_CREATE|inotify.IN_DELETE)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	// Test IN_CREATE
	tf, _ := ioutil.TempFile("/tmp", "")
	tf.Close()
	event, err := in.Read()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Event: %+v\n", event)

	// Test IN_DELETE
	os.Remove(tf.Name())
	event, err = in.Read()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Event: %+v\n", event)

	// Stop watching the specific watch
	in.RemoveWatch(w)

	// Shut down entirely
	in.Close()
}
