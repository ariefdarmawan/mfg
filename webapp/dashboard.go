package webapp

import (
	//"eaciit/mfg"
	"time"

	//"github.com/eaciit/crowd.dev"
	"github.com/eaciit/dbox"
	"github.com/eaciit/knot/knot.v1"
	"github.com/eaciit/toolkit"
)

type Dashboard struct {
}

func (d *Dashboard) Index(ctx *knot.WebContext) interface{} {
	ctx.Config.OutputType = knot.OutputTemplate
	return struct{}{}
}

type timedata struct {
	TimeStamp                             time.Time
	Power, Uptime, Downtime, Speed, Count float32
}

func (d *Dashboard) GetSum(ctx *knot.WebContext)interface{}{
    ctx.Config.OutputType = knot.OutputJson
    result := toolkit.NewResult()
    
    db := DB()
    defer db.Close()
	c, _ := db.Connection.NewQuery().From("actualcost").Select().
		Group("prodlane"). 
		Aggr(dbox.AggrSum,"$costhours","costhour").
		Aggr(dbox.AggrSum,"$costqty","costqty").
		Aggr(dbox.AggrSum,"$qtyalloc","qty").
		Cursor(nil)
	defer c.Close()
	costsum := []*toolkit.M{}
	c.Fetch(&costsum,0,false)
	for _, m := range costsum{
		m.Set("costtotal", m.GetFloat64("costhour") + m.GetFloat64("costqty"))
		m.Set("costperunit", m.GetFloat64("costtotal")/m.GetFloat64("qty"))
	}
    result.Data = toolkit.M{}.
        Set("costsum", costsum)
    return result
}

func (d *Dashboard) GetBySKU(ctx *knot.WebContext)interface{}{
    ctx.Config.OutputType = knot.OutputJson
    result := toolkit.NewResult()
    
    db := DB()
    defer db.Close()
	c, _ := db.Connection.NewQuery().From("actualcost").Select().
		Group("prodlane"). 
		Aggr(dbox.AggrSum,"$costhours","costhour").
		Aggr(dbox.AggrSum,"$costqty","costqty").
		Aggr(dbox.AggrSum,"$qtyalloc","qty").
		Cursor(nil)
	defer c.Close()
	costsum := []*toolkit.M{}
	c.Fetch(&costsum,0,false)
	for _, m := range costsum{
		m.Set("costtotal", m.GetFloat64("costhour") + m.GetFloat64("costqty"))
		m.Set("costperunit", m.GetFloat64("costtotal")/m.GetFloat64("qty"))
	}
    result.Data = toolkit.M{}.
        Set("cost", costsum)
    return result
}