package main

import (
	//"bufio"
	"eaciit/mfg"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	//"time"

	// "github.com/eaciit/dbox"
	_ "github.com/eaciit/dbox/dbc/csvs"
	_ "github.com/eaciit/dbox/dbc/mongo"
	"github.com/eaciit/toolkit"
	"github.com/eaciit/uklam"
)

var (
    e error
	log    *toolkit.LogEngine
	cdone  chan bool
	status string
	path   = func()string{
        wd, _ := os.Getwd()
        return filepath.Join(wd, "..","data")
    }()
)

func main() {
	toolkit.Println("MFG Deamon v0.5 (c) EACIIT")
	toolkit.Println("Run http://localhost:8888/stop to stop the daemon")
	toolkit.Println("")
	log, _ = toolkit.NewLog(true, false, "", "", "")
	defer func() {
		log.Info("Closing daemon")
	}()

	winbox := prepareInbox()
	winbox.Start()
	defer winbox.Stop()

	wrun := prepareRunning()
	wrun.Start()
	defer wrun.Stop()
    
	cdone = make(chan bool)
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		cdone <- true
		w.Write([]byte("Daemon will be stopped"))
	})

	go func() {
		e = http.ListenAndServe(":8888", nil)
		if e != nil {
			log.Error("Can not start daemon REST server. " + e.Error())
			cdone <- true
		}
	}()

	for {
		select {
		case <-cdone:
			status = "Closing"
			return

		default:
			//-- do nothing
		}
	}
}

func prepareInbox() *uklam.FSWalker {
	w := uklam.NewFS(filepath.Join(path, "inbox"))
	w.EachFn = func(dw uklam.IDataWalker, in toolkit.M, info os.FileInfo, r *toolkit.Result) {
		name := info.Name()
        sourcename := filepath.Join(path, "inbox", name)
		dstname := filepath.Join(path, "running", name)
		log.Info(toolkit.Sprintf("Processing " + sourcename))
		e := uklam.FSCopy(sourcename, dstname, true)
		if e != nil {
			log.Error("Processing " + sourcename + " NOK " + e.Error())
		} else {
			log.Info("Processing " + sourcename + " OK ")
		}
	}
	return w
}

func prepareRunning() *uklam.FSWalker {
    w2 := uklam.NewFS(filepath.Join(path, "running"))
	w2.EachFn = func(dw uklam.IDataWalker, in toolkit.M, info os.FileInfo, r *toolkit.Result) {
		name := info.Name()
        tablename  := strings.Replace(name, ".csv", "", -1)
        folder := filepath.Join(path, "running")
        sourcename := filepath.Join(path, "running", name)
		dstnameOK := filepath.Join(path, "success", name)
		dstnameNOK := filepath.Join(path, "fail", name)
		log.Info(toolkit.Sprintf("Reading %s", sourcename))
		e := streamDo(folder, tablename)
		if e == nil {
			uklam.FSCopy(sourcename, dstnameOK, true)
			log.Info(toolkit.Sprintf("%s OK", sourcename))
		} else {
			uklam.FSCopy(sourcename, dstnameNOK, true)
			log.Error(toolkit.Sprintf("%s NOK: %s", sourcename, e.Error()))
		}
	}
	return w2
}

func streamDo(src, tablename string) error {
    if tablename=="opdata"{
        mfg.ProcessOP(log, src)
    } else if tablename=="costdata"{
        mfg.ProcessCost(log, src)
    }
	mfg.Calc(log)
	return nil
}
