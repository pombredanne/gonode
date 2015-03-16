package extra

import (
	"fmt"
	"bufio"
	"net/http"
	"github.com/zenazn/goji/web"
	nc "github.com/rande/gonode/core"
	sq "github.com/lann/squirrel"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"time"
	"container/list"
	"github.com/zenazn/goji/graceful"
	. "github.com/rande/goapp"
	"log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			return
		}
	}
}

func ConfigureGoji(app *App) {

	mux     := app.Get("goji.mux").(*web.Mux)
	logger  := app.Get("logger").(*log.Logger)
	manager := app.Get("gonode.manager").(*nc.PgNodeManager)
	api     := app.Get("gonode.api").(*nc.Api)
	prefix  := ""
	sub     := app.Get("gonode.postgres.subscriber").(*nc.Subscriber)

	var webSocketList = list.New()

	sub.ListenMessage("manager_action", func(notification *pq.Notification) (int, error) {
		logger.Printf("WebSocket: Sending message \n");

		for e := webSocketList.Front(); e != nil; e = e.Next() {
			if err := e.Value.(*websocket.Conn).WriteMessage(websocket.TextMessage, []byte(notification.Extra)); err != nil {
				logger.Printf("Error writing to websocket")
			}
		}

		return nc.PubSubListenContinue, nil
	})

	graceful.PreHook(func() {
		logger.Printf("Closing PostgreSQL subcriber \n");
		sub.Stop()

		logger.Printf("Closing websocket connections \n");
		for e := webSocketList.Front(); e != nil; e = e.Next() {
			e.Value.(*websocket.Conn).Close()
		}
	})

	mux.Get(prefix + "/nodes/stream", func(res http.ResponseWriter, req *http.Request) {
		ws, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			panic(err)
			return
		}

		element := webSocketList.PushBack(ws)

		var closeDefer = func() {
			ws.Close()
			webSocketList.Remove(element)
		}

		defer closeDefer()

		go readLoop(ws)

		// ping remote client, avoid keeping open connection
		for {
			time.Sleep(2 * time.Second);
			if err := ws.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
				return
			}
		}
	})

	mux.Get(prefix + "/nodes/:uuid", func(c web.C, res http.ResponseWriter, req *http.Request) {

		res.Header().Set("X-Generator", "gonode - thomas.rabaix@gmail.com - v" + api.Version)

		values := req.URL.Query()

		if _, raw := values["raw"]; raw {
			node := manager.Find(nc.GetReferenceFromString(c.URLParams["uuid"]))

			if node == nil {
				res.WriteHeader(404);

				return
			}

			handler := manager.GetHandler(node)

			data := handler.GetDownloadData(node)

			res.Header().Set("Content-Type", data.ContentType)

//			if download {
//				res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", data.Filename));
//			}

			data.Stream(node, res)
		} else {
			// send the json value
			res.Header().Set("Content-Type", "application/json")
			api.FindOne(c.URLParams["uuid"], res)
		}
	})

	mux.Post(prefix + "/nodes", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Generator", "gonode - thomas.rabaix@gmail.com - v" + api.Version)

		w := bufio.NewWriter(res)

		err := api.Save(req.Body, w)

		if err == nc.RevisionError {
			res.WriteHeader(http.StatusConflict)
		}

		if err == nc.ValidationError {
			res.WriteHeader(http.StatusPreconditionFailed)
		}

		w.Flush()
	})

	mux.Put(prefix + "/nodes/:uuid", func(c web.C, res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Generator", "gonode - thomas.rabaix@gmail.com - v" + api.Version)

		w := bufio.NewWriter(res)

		err := api.Save(req.Body, w)

		if err == nc.RevisionError {
			res.WriteHeader(http.StatusConflict)
		}

		if err == nc.ValidationError {
			res.WriteHeader(http.StatusPreconditionFailed)
		}

		w.Flush()
	})

	mux.Delete(prefix + "/nodes/:uuid", func(c web.C, res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Generator", "gonode - thomas.rabaix@gmail.com - v" + api.Version)

		api.RemoveOne(c.URLParams["uuid"], res)
	})

	mux.Get(prefix + "/nodes", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Generator", "gonode - thomas.rabaix@gmail.com - v" + api.Version)

		query := api.SelectBuilder()

		req.ParseForm()

		searchForm := nc.GetSearchForm()
		decoder := schema.NewDecoder()
		decoder.Decode(searchForm, req.Form)

		// analyse Meta
		for name, value := range req.Form {
			values := rexMeta.FindStringSubmatch(name)

			if len(values) == 2 {
				searchForm.Meta[values[1]] = value
			}
		}

		// analyse Data
		for name, value := range req.Form {
			values := rexMeta.FindStringSubmatch(name)

			if len(values) == 2 {
				searchForm.Meta[values[1]] = value
			}
		}

		if searchForm.Page < 1 {
			searchForm.Page = 1
		}

		if searchForm.PerPage > 128 {
			searchForm.PerPage = 32
		}

		if searchForm.PerPage < 1 {
			searchForm.PerPage = 32
		}

		if len(searchForm.Source) != 0 {
			query = query.Where(sq.Eq{"source": searchForm.Source})
		}

		if searchForm.Enabled != "" {
			query = query.Where("enabled = ?", searchForm.Enabled)
		}

		if len(searchForm.Type) != 0 {
			query = query.Where(sq.Eq{"type": searchForm.Type})
		}

		if searchForm.Current != "" {
			query = query.Where("current = ?", searchForm.Current)
		}

		if searchForm.Deleted != "" {
			query = query.Where("deleted = ?", searchForm.Deleted)
		}

		if len(searchForm.Uuid) != 0 {
			query = query.Where(sq.Eq{"uuid": searchForm.Uuid})
		}

		if len(searchForm.ParentUuid) != 0 {
			query = query.Where(sq.Eq{"parent_uuid": searchForm.ParentUuid})
		}

		if searchForm.Slug != "" {
			query = query.Where("slug = ?", searchForm.Slug)
		}

		if searchForm.Revision != "" {
			query = query.Where("revision = ?", searchForm.Revision)
		}

		if len(searchForm.Status) != 0 {
			query = query.Where(sq.Eq{"status": searchForm.Status})
		}

		query = query.OrderBy("updated_at DESC")

		// Parse Meta value
		for name, value := range searchForm.Meta {
			//-- SELECT uuid, "data" #> '{tags,1}' as tags FROM nodes WHERE  "data" @> '{"tags": ["sport"]}'
			//-- SELECT uuid, "data" #> '{tags}' AS tags FROM nodes WHERE  "data" -> 'tags' ?| array['sport'];
			if len(value) > 1 {
				query = query.Where(sq.ExprSlice(fmt.Sprintf("meta->'%s' ??| array[" + sq.Placeholders(len(value))+ "]", name), len(value), value),)
			}

			if len(value) == 1 {
				query = query.Where(sq.Expr(fmt.Sprintf("meta->>'%s' = ?", name), value))
			}
		}

		// Parse Data value
		for name, value := range searchForm.Data {
			if len(value) > 1 {
				query = query.Where(sq.ExprSlice(fmt.Sprintf("data->'%s' ??| array[" + sq.Placeholders(len(value))+ "]", name), len(value), value),)
			}

			if len(value) == 1 {
				query = query.Where(sq.Expr(fmt.Sprintf("data->>'%s' = ?", name), value))
			}
		}

		api.Find(res, query, searchForm.Page, searchForm.PerPage)
	})
}

