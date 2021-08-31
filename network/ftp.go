package network

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/adragomir/linuxcncgo/util"
	"github.com/jlaffaye/ftp"
)

type FileEntry struct {
	Name     string
	Path     string
	Type     ftp.EntryType
	Children []FileEntry
}

func CopyFileEntries(files []FileEntry) []FileEntry {
	var internal func([]FileEntry) []FileEntry
	internal = func(fs []FileEntry) []FileEntry {
		out := make([]FileEntry, 0)
		for _, f := range fs {
			tmp := FileEntry{
				Name: f.Name,
				Path: f.Path,
				Type: f.Type,
			}
			if f.Type == ftp.EntryTypeFolder {
				tmp.Children = internal(f.Children)
			} else {
				tmp.Children = nil
			}
			out = append(out, tmp)
		}
		return out
	}
	return internal(files)
}

type FtpConn struct {
	*util.Callbacker

	endpoint string
	conn     *ftp.ServerConn

	listTicker *time.Ticker
	listDone   chan bool
}

func NewFtpConn(endpoint string) *FtpConn {
	tmp := &FtpConn{
		Callbacker: util.NewCallbacker(),
		endpoint:   endpoint,
	}
	tmp.Ensure(true)
	tmp.startTicker()
	return tmp
}

func (c *FtpConn) startTicker() {
	c.listTicker = time.NewTicker(5 * time.Second)
	c.listDone = make(chan bool)

	go func(c *FtpConn) {
		for {
			select {
			case <-c.listDone:
				return
			case _ = <-c.listTicker.C:
				c.refreshFiles()
			}
		}
	}(c)
}

func (c *FtpConn) refreshFiles() {
	c.Ensure(false)

	var internalRefresh func(path string) []FileEntry
	internalRefresh = func(path string) []FileEntry {
		entries, err := c.conn.List(path)
		out := make([]FileEntry, 0)
		if err != nil {
			log.Printf("error getting files from remote path %s: %+v", path, err)
		}
		for _, entry := range entries {
			tmp := FileEntry{
				Name: entry.Name,
				Path: path + entry.Name,
				Type: entry.Type,
			}
			if entry.Type == ftp.EntryTypeFolder {
				tmp.Children = make([]FileEntry, 0)
				tmp.Children = append(tmp.Children, internalRefresh(path+"/"+entry.Name+"/")...)
			}
			out = append(out, tmp)
		}
		return out
	}
	c.RunCbs("filesReady", internalRefresh("/"))
}

func (c *FtpConn) Ensure(force bool) error {
	if c.conn == nil || force {
		if tmpConn, err := ftp.Dial(c.endpoint); err == nil {
			c.conn = tmpConn
			err := c.conn.Login("anonymous", "anonymous@")
			if err != nil {
				log.Printf("Error logging in: %+v", err)
				return errors.New(fmt.Sprintf("Error logging in: %+v", err))
			}
		} else {
			log.Printf("Error initializing ftp service: %+v (dns: %s)", err, c.endpoint)
			return errors.New(fmt.Sprintf("Error initializing ftp service: %+v (dns: %s)", err, c.endpoint))
		}
	}
	return nil
}

func (c *FtpConn) Retr(p string) ([]byte, error) {
	c.Ensure(false)
	resp, err := c.conn.Retr(p)
	if err == nil {
		defer resp.Close()
		buf, errRead := ioutil.ReadAll(resp)
		if errRead != nil {
			return buf, errRead
		}
		return buf, nil
	} else {
		return []byte{}, err
	}
}
