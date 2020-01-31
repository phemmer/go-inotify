package inotify

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInotify(t *testing.T) {
	td, _ := ioutil.TempDir("","")
	defer os.RemoveAll(td)

	in, err := New()
	require.NoError(t, err)
	w, err := in.AddWatch(td, IN_CREATE|IN_DELETE)
	require.NoError(t, err)

	tf, err := ioutil.TempFile(td, "")
	require.NoError(t, err)
	tf.Close()
	os.Remove(tf.Name())

	event1, err := in.Read()
	require.NoError(t, err)
	assert.Equal(t, IN_CREATE, event1.Mask & IN_CREATE)
	assert.Equal(t, path.Base(tf.Name()), event1.Name)

	event2, err := in.Read()
	require.NoError(t, err)
	assert.Equal(t, IN_DELETE, event2.Mask & IN_DELETE)
	assert.Equal(t, path.Base(tf.Name()), event2.Name)

	err = in.RemoveWatch(w)
	require.NoError(t, err)

	event3, err := in.Read()
	require.NoError(t, err)
	assert.Equal(t, IN_IGNORED, event3.Mask & IN_IGNORED)
	assert.Equal(t, int32(w), event3.Wd)

	tf, err = ioutil.TempFile(td, "")
	require.NoError(t, err)
	tf.Close()
	os.Remove(path.Join(td,tf.Name()))

	in.SetReadDeadline(time.Now().Add(time.Second))
	event4, err := in.Read()
	assert.True(t, os.IsTimeout(err))
	assert.Equal(t, Event{}, event4)
}
