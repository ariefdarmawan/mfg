package mfg

import (
    _ "github.com/eaciit/dbox/dbc/csvs"
    _ "github.com/eaciit/dbox/dbc/mongo"
    
    "github.com/eaciit/toolkit"
    "github.com/eaciit/dbox"
    "eaciit/slzr"
    "time"
)

func ProcessOP(l *toolkit.LogEngine, datapath string) {
    config := toolkit.M{"useheader": true, "delimiter": ","}
    //toolkit.M(config).Set("newfile",true)
    df := slzr.NewDataSource("csvs", &dbox.ConnectionInfo{datapath,"","","",config},"opdata")
    dd := slzr.NewDataSource("mongo", &dbox.ConnectionInfo{"localhost:27123","ectest","","",nil},"opdata")
    dm := slzr.NewDataMap(df, dd, nil)

    e := dm.Start()
    if e!=nil {
        l.Error(toolkit.Sprintf("Unable to start: %s", e.Error()))
    }
    e = dm.Wait()
    if e!=nil {
        l.Error(toolkit.Sprintf("Process fail: %s", e.Error()))
    }
}

func ProcessCost(l *toolkit.LogEngine, datapath string) {
    config := toolkit.M{"useheader": true, "delimiter": ","}
    //toolkit.M(config).Set("newfile",true)
   
    df := slzr.NewDataSource("csvs", &dbox.ConnectionInfo{datapath,"","","",config},"costdata")
    dd := slzr.NewDataSource("mongo", &dbox.ConnectionInfo{"localhost:27123","ectest","","",nil},"costdata")
    dm := slzr.NewDataMap(df, dd, nil)

    conn, _ := dd.Connection()
    defer conn.Close()
    qsave := conn.NewQuery().From("costsum").SetConfig("multiexec",true).Save()
    dm.FnPost = func(m *toolkit.M)error{
        wcorsku := ""
        id := ""
        msum := toolkit.M{}
        date := m.Get("transdate").(time.Time)
        wcid := m.GetString("wcid")
        skuid := m.GetString("skuid")
        amount := m.GetFloat64("amount")
        qty := m.GetFloat64("qty")
        avg := float64(amount / qty)
        msum.Set("lastcostperunit", avg)
        msum.Set("wcid", wcid)
        msum.Set("skuid", skuid)
        if m.GetString("wcid")!="" {
            wcorsku="wc"
            id = toolkit.Sprintf("%d_%d_%s_%s", 
                date.Year(), date.Month(),wcorsku, wcid)
        } else {
            wcorsku="sku"
            id = toolkit.Sprintf("%d_%d_%s_%s", 
                date.Year(), date.Month(),wcorsku, skuid)
        }
        msum.Set("_id", id)
        esave := qsave.Exec(toolkit.M{}.Set("data",msum))
        if esave!=nil {
            toolkit.Printfn("Unable to generate sum: %s - %s", 
                id, esave.Error())
        }
        return nil
    }

    e := dm.Start()
    if e!=nil {
        l.Error(toolkit.Sprintf("Unable to start: %s", e.Error()))
    }
    e = dm.Wait()
    if e!=nil {
        l.Error(toolkit.Sprintf("Process fail: %s", e.Error()))
    }
}

func Calc(l *toolkit.LogEngine) {
    conn, _ := dbox.NewConnection("mongo", &dbox.ConnectionInfo{"localhost:27123","ectest","","",nil})
    conn.Connect()
    defer conn.Close()

    sumwc := toolkit.M{}
    sumwca := toolkit.M{}
    sumsku := toolkit.M{}

    totalwc := float64(0)
    csum, _ := conn.NewQuery().From("costsum").Select().Cursor(nil)
    defer csum.Close()
    for{
        m := toolkit.M{}
        esum := csum.Fetch(&m,1,false)
        if esum!=nil{
            break
        }

        wcid := m.GetString("wcid")
        skuid := m.GetString("skuid")
        amt := m.GetFloat64("lastcostperunit")
        if wcid!="" {
            totalwc += amt
            sumwc.Set(wcid, amt)
        } else {
            sumsku.Set(skuid,amt)
        }
    }

    for k, v := range sumwc{
        sumwca.Set(k, float64(v.(float64)/totalwc))
    }

    ctrx, _ := conn.NewQuery().From("opdata").Select().Cursor(nil)
    defer ctrx.Close()
    qsave := conn.NewQuery().From("actualcost").Save().SetConfig("multiexec",true)
    for{
        m := toolkit.M{}
        if ctrx.Fetch(&m,1,false)!=nil {
            break
        }

        wcid := m.GetString("wcid")
        skuid := m.GetString("skuid")
        hours := m.GetFloat64("qty1")
        qty := m.GetFloat64("qty2")
        qtyalloc := qty * sumwca.GetFloat64(wcid)
        m.Set("costhours", hours * sumwc.GetFloat64(wcid))
        m.Set("costqty", qty * sumsku.GetFloat64(skuid) * sumwca.GetFloat64(wcid))
        m.Set("qtyalloc", qtyalloc)
        //m.Set("qty2", qty2)
        qsave.Exec(toolkit.M{}.Set("data",m))
    }
}