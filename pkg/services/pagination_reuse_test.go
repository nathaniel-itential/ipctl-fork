// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"testing"

	"github.com/itential/ipctl/pkg/client"
	"github.com/stretchr/testify/assert"
)

// pagedClient is a client.Client that serves a fixed sequence of response
// bodies, one per request, regardless of method. It exists to exercise
// multi-page pagination, which the shared testlib mux cannot do (it registers a
// single fixed body per path and ignores query parameters).
type pagedClient struct {
	pages [][]byte
	calls int
}

func (c *pagedClient) next() (*client.Response, error) {
	body := c.pages[c.calls]
	c.calls++
	return &client.Response{StatusCode: 200, Body: body}, nil
}

func (c *pagedClient) Get(*client.Request) (*client.Response, error)    { return c.next() }
func (c *pagedClient) Post(*client.Request) (*client.Response, error)   { return c.next() }
func (c *pagedClient) Put(*client.Request) (*client.Response, error)    { return c.next() }
func (c *pagedClient) Delete(*client.Request) (*client.Response, error) { return c.next() }
func (c *pagedClient) Patch(*client.Request) (*client.Response, error)  { return c.next() }
func (c *pagedClient) Trace(*client.Request) (*client.Response, error)  { return c.next() }

// TestTransformationGetAllNoFieldBleedAcrossPages is a regression test for the
// paginated-response-reuse merge bug (issue #197). When GetAll reused a single
// response struct across pagination pages, encoding/json merged map fields and
// reused slice backing arrays, so a transformation on page two inherited the
// reference-type fields (incoming/outgoing/steps) of the transformation at the
// same index on page one. Each page must decode into a fresh struct.
func TestTransformationGetAllNoFieldBleedAcrossPages(t *testing.T) {
	// Two pages, one transformation each, total of 2 forces a second fetch.
	// The page-one transformation carries a marker key ("onlyA") in its
	// incoming object that the page-two transformation does not have.
	page1 := []byte(`{
		"total": 2,
		"results": [
			{"_id": "a", "name": "A", "incoming": [{"$id": "a", "onlyA": "x"}]}
		]
	}`)
	page2 := []byte(`{
		"total": 2,
		"results": [
			{"_id": "b", "name": "B", "incoming": [{"$id": "b"}]}
		]
	}`)

	svc := NewTransformationService(&pagedClient{pages: [][]byte{page1, page2}})

	res, err := svc.GetAll()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))

	// Page-two transformation must contain only its own incoming fields. With
	// the bug, "onlyA" bleeds in from page one's same-index element.
	b := res[1]
	assert.Equal(t, "B", b.Name)
	assert.Equal(t, "b", b.Incoming[0]["$id"])
	_, leaked := b.Incoming[0]["onlyA"]
	assert.False(t, leaked, "page-two transformation must not inherit fields from page one")

	// Page-one transformation must be unchanged after the second page is
	// fetched. With the bug, the shared backing map is mutated in place, so
	// its $id flips to "b".
	a := res[0]
	assert.Equal(t, "A", a.Name)
	assert.Equal(t, "a", a.Incoming[0]["$id"])
}
