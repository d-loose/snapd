// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package client_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/xerrors"
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/asserts"
)

const happyModelAssertionResponse = `type: model
authority-id: mememe
series: 16
brand-id: mememe
model: test-model
architecture: amd64
base: core18
gadget: pc=18
kernel: pc-kernel=18
required-snaps:
  - core
  - hello-world
timestamp: 2017-07-27T00:00:00.0Z
sign-key-sha3-384: 8B3Wmemeu3H6i4dEV4Q85Q4gIUCHIBCNMHq49e085QeLGHi7v27l3Cqmemer4__t

AcLBcwQAAQoAHRYhBMbX+t6MbKGH5C3nnLZW7+q0g6ELBQJdTdwTAAoJELZW7+q0g6ELEvgQAI3j
jXTqR6kKOqvw94pArwdMDUaZ++tebASAZgso8ejrW2DQGWSc0Q7SQICIR8bvHxqS1GtupQswOzwS
U8hjDTv7WEchH1jylyTj/1W1GernmitTKycecRlEkSOE+EpuqBFgTtj6PdA1Fj3CiCRi1rLMhgF2
luCOitBLaP+E8P3fuATsLqqDLYzt1VY4Y14MU75hMn+CxAQdnOZTI+NzGMasPsldmOYCPNaN/b3N
6/fDLU47RtNlMJ3K0Tz8kj0bqRbegKlD0RdNbAgo9iZwNmrr5E9WCu9f/0rUor/NIxO77H2ExIll
zhmsZ7E6qlxvAgBmzKgAXrn68gGrBkIb0eXKiCaKy/i2ApvjVZ9HkOzA6Ldd+SwNJv/iA8rdiMsq
p2BfKV5f3ju5b6+WktHxAakJ8iqQmj9Yh7piHjsOAUf1PEJd2s2nqQ+pEEn1F0B23gVCY/Fa9YRQ
iKtWVeL3rBw4dSAaK9rpTMqlNcr+yrdXfTK5YzkCC6RU4yzc5MW0hKeseeSiEDSaRYxvftjFfVNa
ZaVXKg8Lu+cHtCJDeYXEkPIDQzXswdBO1M8Mb9D0mYxQwHxwvsWv1DByB+Otq08EYgPh4kyHo7ag
85yK2e/NQ/fxSwQJMhBF74jM1z9arq6RMiE/KOleFAOraKn2hcROKnEeinABW+sOn6vNuMVv
`

// note: this serial assertion was generated by adding print statements to the
// test in api_model_test.go that generate a fake serial assertion
const happySerialAssertionResponse = `type: serial
authority-id: my-brand
brand-id: my-brand
model: my-old-model
serial: serialserial
device-key:
    AcZrBFaFwYABAvCgEOrrLA6FKcreHxCcOoTgBUZ+IRG7Nb8tzmEAklaQPGpv7skapUjwD1luE2go
    mTcoTssVHrfLpBoSDV1aBs44rg3NK40ZKPJP7d2zkds1GxUo1Ea5vfet3SJ4h3aRABEBAAE=
device-key-sha3-384: iqLo9doLzK8De9925UrdUyuvPbBad72OTWVE9YJXqd6nz9dKvwJ_lHP5bVxrl3VO
timestamp: 2019-08-26T16:34:21-05:00
sign-key-sha3-384: anCEGC2NYq7DzDEi6y7OafQCVeVLS90XlLt9PNjrRl9sim5rmRHDDNFNO7ODcWQW

AcJwBAABCgAGBQJdZFBdAADCLALwR6Sy24wm9PffwbvUhOEXneyY3BnxKC0+NgdHu1gU8go9vEP1
i+Flh5uoS70+MBIO+nmF8T+9JWIx2QWFDDxvcuFosnIhvUajCEQohauys5FMz/H/WvB0vrbTBpvK
eg==`

const noModelAssertionYetResponse = `
{
	"type": "error",
	"status-code": 404,
	"status": "Not Found",
	"result": {
	  "message": "no model assertion yet",
	  "kind": "assertion-not-found",
	  "value": "model"
	}
}`

const noSerialAssertionYetResponse = `
{
	"type": "error",
	"status-code": 404,
	"status": "Not Found",
	"result": {
	  "message": "no serial assertion yet",
	  "kind": "assertion-not-found",
	  "value": "serial"
	}
}`

func (cs *clientSuite) TestClientRemodelEndpoint(c *C) {
	cs.cli.Remodel([]byte(`{"new-model": "some-model"}`))
	c.Check(cs.req.Method, Equals, "POST")
	c.Check(cs.req.URL.Path, Equals, "/v2/model")
}

func (cs *clientSuite) TestClientRemodel(c *C) {
	cs.status = 202
	cs.rsp = `{
		"type": "async",
		"status-code": 202,
                "result": {},
		"change": "d728"
	}`
	remodelJsonData := []byte(`{"new-model": "some-model"}`)
	id, err := cs.cli.Remodel(remodelJsonData)
	c.Assert(err, IsNil)
	c.Check(id, Equals, "d728")
	c.Assert(cs.req.Header.Get("Content-Type"), Equals, "application/json")

	body, err := ioutil.ReadAll(cs.req.Body)
	c.Assert(err, IsNil)
	jsonBody := make(map[string]string)
	err = json.Unmarshal(body, &jsonBody)
	c.Assert(err, IsNil)
	c.Check(jsonBody, HasLen, 1)
	c.Check(jsonBody["new-model"], Equals, string(remodelJsonData))
}

func (cs *clientSuite) TestClientGetModelHappy(c *C) {
	cs.status = 200
	cs.rsp = happyModelAssertionResponse
	modelAssertion, err := cs.cli.CurrentModelAssertion()
	c.Assert(err, IsNil)
	expectedAssert, err := asserts.Decode([]byte(happyModelAssertionResponse))
	c.Assert(err, IsNil)
	c.Assert(modelAssertion, DeepEquals, expectedAssert)
}

func (cs *clientSuite) TestClientGetModelNoModel(c *C) {
	cs.status = 404
	cs.rsp = noModelAssertionYetResponse
	cs.header = http.Header{}
	cs.header.Add("Content-Type", "application/json")
	_, err := cs.cli.CurrentModelAssertion()
	c.Assert(err, ErrorMatches, "no model assertion yet")
}

func (cs *clientSuite) TestClientGetModelNoSerial(c *C) {
	cs.status = 404
	cs.rsp = noSerialAssertionYetResponse
	cs.header = http.Header{}
	cs.header.Add("Content-Type", "application/json")
	_, err := cs.cli.CurrentSerialAssertion()
	c.Assert(err, ErrorMatches, "no serial assertion yet")
}

func (cs *clientSuite) TestClientGetSerialHappy(c *C) {
	cs.status = 200
	cs.rsp = happySerialAssertionResponse
	serialAssertion, err := cs.cli.CurrentSerialAssertion()
	c.Assert(err, IsNil)
	expectedAssert, err := asserts.Decode([]byte(happySerialAssertionResponse))
	c.Assert(err, IsNil)
	c.Assert(serialAssertion, DeepEquals, expectedAssert)
}

func (cs *clientSuite) TestClientCurrentModelAssertionErrIsWrapped(c *C) {
	cs.err = errors.New("boom")
	_, err := cs.cli.CurrentModelAssertion()
	var e xerrors.Wrapper
	c.Assert(err, Implements, &e)
}
