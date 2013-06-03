// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package apiserver_test

import (
	"io"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/rpc"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/state/apiserver"
	coretesting "launchpad.net/juju-core/testing"
)

func (s *suite) TestStop(c *C) {
	// Start our own instance of the server so we have
	// a handle on it to stop it.
	srv, err := apiserver.NewServer(s.State, "localhost:0", []byte(coretesting.ServerCert), []byte(coretesting.ServerKey))
	c.Assert(err, IsNil)

	stm, err := s.State.AddMachine("series", state.JobHostUnits)
	c.Assert(err, IsNil)
	err = stm.SetProvisioned("foo", "fake_nonce")
	c.Assert(err, IsNil)
	err = stm.SetPassword("password")
	c.Assert(err, IsNil)

	// Note we can't use openAs because we're not connecting to
	// s.APIConn.
	st, err := api.Open(&api.Info{
		Tag:      stm.Tag(),
		Password: "password",
		Addrs:    []string{srv.Addr()},
		CACert:   []byte(coretesting.CACert),
	})
	c.Assert(err, IsNil)
	defer st.Close()

	machiner, err := st.Machiner()
	m, err := machiner.Machine(stm.Id())
	c.Assert(err, IsNil)
	c.Assert(m.Id(), Equals, stm.Id())

	err = srv.Stop()
	c.Assert(err, IsNil)

	_, err = st.Machiner()
	// The client has not necessarily seen the server shutdown yet,
	// so there are two possible errors.
	if err != rpc.ErrShutdown && err != io.ErrUnexpectedEOF {
		c.Fatalf("unexpected error from request: %v", err)
	}

	// Check it can be stopped twice.
	err = srv.Stop()
	c.Assert(err, IsNil)
}
