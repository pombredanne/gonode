// Copyright © 2014-2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package modules

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rande/goapp"
	"github.com/rande/gonode/modules/base"
	"github.com/rande/gonode/modules/blog"
	"github.com/rande/gonode/test"
	"github.com/stretchr/testify/assert"
)

func Test_Prism_Blog_Archive(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		// WITH
		manager := app.Get("gonode.manager").(*base.PgNodeManager)

		node := app.Get("gonode.handler_collection").(base.HandlerCollection).NewNode("blog.post")
		data := node.Data.(*blog.Post)
		data.Title = "Blog Post 1"

		manager.Save(node, false)

		archive := app.Get("gonode.handler_collection").(base.HandlerCollection).NewNode("core.index")
		archive.Name = "Blog Archive"

		manager.Save(archive, false)

		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/prism/%s", ts.URL, archive.Uuid))

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func Test_Prism_Bad_Request(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		// WITH
		// create a valid user into the database ...
		manager := app.Get("gonode.manager").(*base.PgNodeManager)

		node := app.Get("gonode.handler_collection").(base.HandlerCollection).NewNode("blog.post")
		data := node.Data.(*blog.Post)
		data.Title = "Blog Post 1"

		manager.Save(node, false)

		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/prism/%s.json", ts.URL, node.Uuid))

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func Test_Prism_Format(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		manager := app.Get("gonode.manager").(*base.PgNodeManager)

		home := app.Get("gonode.handler_collection").(base.HandlerCollection).NewNode("core.root")
		home.Name = "Homepage"
		manager.Save(home, false)

		raw := app.Get("gonode.handler_collection").(base.HandlerCollection).NewNode("core.raw")
		raw.Name = "Human.txt"
		raw.Slug = "human.txt"

		manager.Save(raw, false)
		manager.Move(raw.Uuid, home.Uuid)

		raw2 := app.Get("gonode.handler_collection").(base.HandlerCollection).NewNode("core.raw")
		raw2.Name = "Human"
		raw2.Slug = "human"

		manager.Save(raw2, false)
		manager.Move(raw2.Uuid, home.Uuid)

		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/human", ts.URL))

		assert.Equal(t, http.StatusOK, res.StatusCode, "Cannot find /human")

		res, _ = test.RunRequest("GET", fmt.Sprintf("%s/human.txt", ts.URL))

		assert.Equal(t, http.StatusOK, res.StatusCode, "Cannot find /human.txt")
	})
}
