package editor

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type editorState int

const (
	editorWrite editorState = iota + 1
	editorRead
)

type Editor struct {
	TmpDir    string
	TmpPrefix string
	Command   string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	state     editorState
	err       error
	tmpFile   *os.File
}

func (e *Editor) open() error {
	if e.err != nil {
		return e.err
	}
	if e.tmpFile != nil {
		return nil
	}
	file, err := ioutil.TempFile(e.TmpDir, e.TmpPrefix)
	if err != nil {
		e.err = err
		return err
	}
	e.tmpFile = file
	return nil
}

func (e *Editor) Write(b []byte) (int, error) {
	if err := e.open(); err != nil {
		return 0, err
	}
	if e.state != editorWrite {
		return 0, fmt.Errorf("Editor is not writable.")
	}
	if n, err := e.tmpFile.Write(b); err != nil {
		e.err = err
		return n, e.err
	} else {
		return n, nil
	}
}

func (e *Editor) Read(b []byte) (int, error) {
	if err := e.open(); err != nil {
		return 0, err
	}
	if e.state != editorRead {
		return 0, fmt.Errorf("Editor is not readable.")
	}
	if n, err := e.tmpFile.Read(b); err != nil {
		e.err = err
		return n, e.err
	} else {
		return n, nil
	}
}

func (e *Editor) Run() error {
	if err := e.open(); err != nil {
		return err
	}
	if e.Command == "" {
		e.err = fmt.Errorf("No Editor Command set.")
		return e.err
	}
	// @TODO Escape file argument better
	cmd := exec.Command("/bin/sh", "-c", e.Command+" '"+e.tmpFile.Name()+"'")
	cmd.Stdin = e.Stdin
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr
	if err := cmd.Run(); err != nil {
		e.err = err
		return err
	}
	if _, err := e.tmpFile.Seek(0, os.SEEK_SET); err != nil {
		e.err = err
		return err
	}
	e.state = editorRead
	return nil
}

func (e *Editor) Close() error {
	if e.tmpFile == nil {
		return fmt.Errorf("Nothing to close.")
	}
	removeErr := os.Remove(e.tmpFile.Name())
	closeErr := e.tmpFile.Close()
	if removeErr != nil {
		return removeErr
	}
	return closeErr
}

func New() *Editor {
	e := &Editor{
		state:   editorWrite,
		Command: os.Getenv("EDITOR"),
		TmpDir:  os.TempDir(),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
	return e
}
