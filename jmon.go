package main

import ( "bufio"; "bytes"; "fmt"; "io"; "log"; "net/http"; "net/smtp"; "os"; "strings"; "strconv"; "time"; "encoding/json"; "io/ioutil"
)

var pgm_name    string = "jmon"
var http_port   string = ":80"
var systemID	string = "?"

// ====================================
var pgm_version string = "0.90.0"
// ====================================


var et_start	int
var testName	string = "? / ?"
var alert_email	string = "?"
var linksFileName = "vm_links.json"
var notesFileName = "notes.json"

var cnt_debug	int32
var cnt_pass	int32
var cnt_fail	int32
var cnt_min		int32
var now_pass	int32
var now_fail	int32

var m_last[2]	int32
var h_last[2]	int32
var d_last[2]	int32

var m_next		int32
var h_next		int32
var d_next		int32

var data[6][60]	int32

var threshold_cnt	int = 5

var p0 string   = `<!DOCTYPE html><html><head><title>Page Title</title></head><body>`
var p2 string   = `</body></html>`

var vm_links interface{}

var notes interface{}

func emailThresholdAlert() {
	fmt.Printf("Email alert sent to [%s]\n", alert_email)
    c, err := smtp.Dial("mail.novell.com:25")
    if err != nil { log.Fatal(err) }
    defer c.Close()
    // Set the sender and recipient.
    c.Mail("system.test@novell.com")
    c.Rcpt(alert_email)
    // Send the email body.
    wc, err := c.Data()
    if err != nil { log.Fatal(err) }
    defer wc.Close()
	ebody := fmt.Sprintf("Threshold failure on [%s] for [%s]\n", systemID, testName)
    buf := bytes.NewBufferString(ebody)
    if _, err = buf.WriteTo(wc); err != nil { log.Fatal(err) }
}

func HumanFormat(n int64) string {
    in := strconv.FormatInt(n, 10)
    out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
    if in[0] == '-' {
        in, out[0] = in[1:], '-'
    }

    for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
        out[j] = in[i]
        if i == 0 {
            return string(out)
        }
        if k++; k == 3 {
            j, k = j-1, 0
            out[j] = ','
        }
    }
}

func elapsedTime(getHMS bool) string {
	var et string
	var et_now int
	var et_sec int

	var dd int
	var hh int
	var mm int
	var ss int
	var x int

	et_now = int(time.Now().Unix())
	et_sec = et_now - et_start
	x = et_sec

	dd = x / 86400
	if dd > 0 { x = x - (dd * 86400) }

	if getHMS == false {
		et = fmt.Sprintf("%d", dd)
		return et
	}

	hh = x / 3600
	if hh > 0 { x = x - (hh * 3600) }

	mm = x / 60
	if mm > 0 { x = x - (mm * 60) }

	ss = x

	et = fmt.Sprintf("%02d:%02d:%02d", hh, mm, ss)
	
	return et
}


func handlerUpdateLoginCount(w http.ResponseWriter, r *http.Request) {
	
//	fmt.Printf("[%s]\n", r.URL.RawQuery)
	if r.URL.RawQuery == "debug" { cnt_debug++; return }
	if r.URL.RawQuery == "pass"  { cnt_pass++;  return }
	cnt_fail++
}

func handlerResetLoginCount(w http.ResponseWriter, r *http.Request) {
	var resp string
	var p1 string = `<h1>Passed = 0</h1><p></p><h1>Failed = 0</h1>`
	
	cnt_debug	= 0
	cnt_pass	= 0
	cnt_fail	= 0
	resp 		= p0 + p1 + p2
	
	io.WriteString(w, resp)
}

func handlerDebugLoginCount(w http.ResponseWriter, r *http.Request) {
	var resp string
	var p1 string
	var rec string = `<h1>Debug = %d</h1><p></p><h1>Passed = %d</h1><p></p><h1>Failed = %d</h1>`
		
	p1 = fmt.Sprintf(rec, cnt_debug, cnt_pass, cnt_fail)
	resp = p0 + p1 + p2
	
	io.WriteString(w, resp)
}


func handlerStatusLoginCount(w http.ResponseWriter, r *http.Request) {
	var resp string
	var p1 string
	var rec string = `<h1>Passed = %d</h1><p></p><h1>Failed = %d</h1>`
	
	p1 = fmt.Sprintf(rec, cnt_pass, cnt_fail)
	resp = p0 + p1 + p2
	
	io.WriteString(w, resp)
}


func handlerPlotData(w http.ResponseWriter, r *http.Request) {
	var label string = `"Passes" `
	if strings.Contains(r.URL.RawQuery, "f") {
		label = `"Failures"`
	}
	var resp string = `{"label":` + label + `, "data": [`
	var temp string
	var sep  string = ","
	var x int32		= m_next
	var p int		= 0
	var f float64	= 60.0
	

	if r.URL.RawQuery == "pm" { x = m_next; p = 0; f = 60.0 }
	if r.URL.RawQuery == "fm" { x = m_next; p = 1; f = 60.0 }

	if r.URL.RawQuery == "ph" { x = h_next; p = 2; f = 3600.0 }
	if r.URL.RawQuery == "fh" { x = h_next; p = 3; f = 3600.0 }

	if r.URL.RawQuery == "pd" { x = d_next; p = 4; f = 86400.0 }
	if r.URL.RawQuery == "fd" { x = d_next; p = 5; f = 86400.0 }

	for i:=1; i<61; i++ {
		if i == 60 { sep = "" }
		temp = fmt.Sprintf("[%d,%.2f]%s",i, float64(data[p][x]) / f, sep)
		resp = resp + temp
		x++
		if x >= 60 { x = 0 }
	}

	resp = resp + "]}"

	io.WriteString(w, resp)
}

func handlerInfo(w http.ResponseWriter, r *http.Request) {
	var strpass string = HumanFormat(int64(cnt_pass))
	var strfail string = HumanFormat(int64(cnt_fail))

	// var p1 string = `<div id="users-contain" class="ui-widget">`
	// var p2 string = `<table id="users" class="ui-button">`
	// var p3 string = `<thead><tr class="ui-widget-header"><th>Product / Test</th><th>Days</th><th>HH:MM:SS</th><th>Logins</th><th>Failed</th></tr></thead>`
	// var p4 string = `<tbody><tr><td>` + testName + `</td><td>` + elapsedTime(false) + `</td>`	
	// var p5 string = `<td>` + elapsedTime(true) + `</td>`
	// var p6 string = `<td>` + strpass + `</td><td>` + strfail + `</td></tr>`

	// var p9 string = `</tbody></table></div>`	

	// resp = p1 + p2 + p3 + p4 + p5 + p6 + p9


	respMap := map[string] string{ 
		"test": testName, 
		"days": elapsedTime(false), 
		"time": elapsedTime(true), 
		"passed": strpass, 
		"failed" : strfail}
	
	respBytes, _:= json.Marshal(respMap)
	io.WriteString(w, string(respBytes))
}

func alertCheck() {
	threshold_cnt--
//	if threshold_cnt == 0 { emailThresholdAlert() }
}


func updateHistory() {

	now_fail = cnt_fail				
	now_pass = cnt_pass
	cnt_min++

	// min
	mpass := now_pass - m_last[0];	m_last[0] = cnt_pass
	mfail := now_fail - m_last[1];	m_last[1] = cnt_fail
	data[0][m_next] = mpass
	data[1][m_next] = mfail
	m_next++
	if m_next >= 60 { m_next = 0 }

	if mpass == 0 { alertCheck() }

	// hour
	if cnt_min % 60 == 0 {
		hpass := now_pass - h_last[0];	h_last[0] = cnt_pass
		hfail := now_fail - h_last[1];	h_last[1] = cnt_fail
		data[2][h_next] = hpass
		data[3][h_next] = hfail
		h_next++
		if h_next >= 60 { h_next = 0 }
	}

	// day
	if cnt_min % 1440 == 0 {
		dpass := now_pass - d_last[0];	d_last[0] = cnt_pass
		dfail := now_fail - d_last[1];	d_last[1] = cnt_fail
		data[4][d_next] = dpass
		data[5][d_next] = dfail
		d_next++
		if d_next >= 60 { d_next = 0 }
	}
}


func parseFile(name string) {
	var sa[] string
	var sa2[] string
	var line string
	var err error
	var temp int
	var min	int

	file, err := os.Open(name)
	if err != nil { fmt.Println("No [recovery.log] file found"); return }
	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line = scanner.Text()
		if len(line) < 20 { fmt.Printf("short line\n"); continue }
		sa = strings.Split(line, " ")
		if len(sa) != 3 { continue }
	
		min++
		sa2 = strings.Split(strings.Trim(sa[2], "[]"), ",")

		temp, err = strconv.Atoi(sa2[0])
		cnt_pass = int32(temp)
		temp, err = strconv.Atoi(sa2[1])
		cnt_fail = int32(temp)
		updateHistory()
	}

	et_start = int(time.Now().Unix()) - (min * 60)
}



func summary() {
	tChan := time.NewTicker(time.Minute).C
	
	for {
		select {
			case <- tChan:
				updateHistory()
				log.Printf("[%d,%d]\n", cnt_pass, cnt_fail)
		}		
	}
}

func parseSystemIDFile() {
	file, err := os.Open("systemID.file")
	if err != nil {log.Fatal(err)}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		sa := strings.Fields(s)
		if len(sa) == 2 {
			if strings.HasPrefix(sa[0], "systemID") == true { systemID = sa[1]; fmt.Printf("systemID [%s]\n", systemID) }
		}
	}

	if err := scanner.Err(); err != nil {log.Fatal(err)}
}

func refreshSystemInfo() {
//	var v string

//	v = sqlQueryKV(systemID)
//	if (v != "?") { testName = v; fmt.Printf("product/test = [%s]\n", testName) }

//	v = sqlQueryKV("jmeter-alert")
//	if (v != "?") { alert_email = v; fmt.Printf("alert/email  = [%s]\n", alert_email) }

	threshold_cnt = 5;
}

func handlerSetInfo(w http.ResponseWriter, r *http.Request) {
	var response string = "0"
	
	if len(r.URL.RawQuery) == 0 {
		response = fmt.Sprintf("TestName = %s", testName)
		io.WriteString(w, response)
		return
	}
	
	testName = (strings.Replace(r.URL.RawQuery, "%20", " ", -1))
	log.Printf("--> TestName %s", testName)
	
	response = fmt.Sprintf("TestName = %s", testName)
	io.WriteString(w, response)
}

func handlerVmLinks(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET" {
		respBytes, _:= json.Marshal(vm_links)
		io.WriteString(w, string(respBytes))
	} else if r.Method == "POST" {
		byt, err := ioutil.ReadAll(r.Body)
		writeToLinksFile(byt)
		err1 := json.Unmarshal(byt, &vm_links); 
		
		if err1 != nil || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		respBytes, _:= json.Marshal(vm_links)
		io.WriteString(w, string(respBytes))
	}
}

func writeToLinksFile(byt []byte){
	err := ioutil.WriteFile(linksFileName, byt, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readLinksFile() []byte{
	dat, err := ioutil.ReadFile(linksFileName)
	if err != nil {
		log.Fatal(err)
	}
	return dat
}


// ------------------------------------------------------------------------------------------
func handlerNotes(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET" {
		respBytes, _:= json.Marshal(notes)
		io.WriteString(w, string(respBytes))
	} else if r.Method == "POST" {
		byt, err := ioutil.ReadAll(r.Body)
		writeToNotesFile(byt)
		err1 := json.Unmarshal(byt, &notes); 
		
		if err1 != nil || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		respBytes, _:= json.Marshal(notes)
		io.WriteString(w, string(respBytes))
	}
}

func writeToNotesFile(byt []byte){
	err := ioutil.WriteFile(notesFileName, byt, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readNotesFile() []byte{
	dat, err := ioutil.ReadFile(notesFileName)
	if err != nil {
		log.Fatal(err)
	}
	return dat
}
// ------------------------------------------------------------------------------------------

func handlerRefreshSystemInfo(w http.ResponseWriter, r *http.Request) {
	refreshSystemInfo()
}


func main() {
	if len(os.Args) > 1 {
		testName = os.Args[1]
	}
	logfile, err := os.OpenFile(pgm_name + "_output.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Unable to open/create the log file")
		os.Exit(-1)
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	et_start = int(time.Now().Unix())
	currentTime := time.Now()
	fmt.Println(pgm_name + " [version " + pgm_version + "] port = " + http_port)
	parseSystemIDFile()
	fmt.Println("Started:", currentTime.Format("2006-01-02 3:4:5 pm"))
	
	parseFile("recovery.log")
	refreshSystemInfo()

	b := readLinksFile()
	json.Unmarshal(b, &vm_links)

	c := readNotesFile()
	json.Unmarshal(c, &notes)

	http.Handle("/", http.FileServer(http.Dir("./site")))
	http.HandleFunc("/debugLoginCount", handlerDebugLoginCount)
	http.HandleFunc("/statusLoginCount", handlerStatusLoginCount)
	http.HandleFunc("/request", handlerUpdateLoginCount)
	http.HandleFunc("/refreshSystemInfo", handlerRefreshSystemInfo)
	http.HandleFunc("/resetLoginCount", handlerResetLoginCount)
	http.HandleFunc("/setInfo", handlerSetInfo)
	http.HandleFunc("/system/test/data", handlerPlotData)
	http.HandleFunc("/ajax/getInfo", handlerInfo)
	http.HandleFunc("/vmLinks", handlerVmLinks)
	http.HandleFunc("/notes", handlerNotes)

	go summary()

	err = http.ListenAndServe(http_port, nil)
	if err != nil {log.Fatal("ListenAndServe: ", err)}
}

