package webapp

import (
	"os"

	"path/filepath"

	"github.com/eaciit/dbox"
	_ "github.com/eaciit/dbox/dbc/mongo"
	"github.com/eaciit/knot/knot.v1"
	"github.com/eaciit/orm"
)

func App() *knot.App {
	app := knot.NewApp("mfg")
	wd, _ := os.Getwd()
	wd = filepath.Join(wd, "..", "webapp")
	app.ViewsPath = filepath.Join(wd, "views")
	app.LayoutTemplate = "_layout.html"
	app.Static("static", filepath.Join(wd, "assets"))
	app.Register(&Dashboard{})
	app.DefaultOutputType = knot.OutputHtml
	return app
}

func conn() dbox.IConnection {
	conn, _ := dbox.NewConnection("mongo", &dbox.ConnectionInfo{"localhost:27123", "ectest", "", "", nil})
	conn.Connect()
	return conn
}

func DB() *orm.DataContext {
	return orm.New(conn())
}
