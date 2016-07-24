package main

import (
	"eaciit/mfg/webapp"

	"github.com/eaciit/knot/knot.v1"
)

func main() {
	app1 := webapp.App()
	//app2 := webapp.App()
	//knot.StartApp(app, "localhost:9100")

	knot.RegisterApp(app1)
	knot.StartContainer(&knot.AppContainerConfig{Address: "localhost:9100"})
}
