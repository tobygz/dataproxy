package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

type exeMain struct {
	_dbop  *Dbagent
	_dburl string
	_port  string
}

func (e *exeMain) init() {
	e._dbop = &Dbagent{}
	e._dbop.init(e._dburl)
}

var g_main *exeMain

func init() {
	g_main = &exeMain{}
	flag.StringVar(&g_main._dburl, "dburl", "server:9MbDiZnsfn3W1SFO!@tcp(127.0.0.1:33061)/engine_v4", "string flag value")
	flag.StringVar(&g_main._port, "port", "8020", "")
	flag.Parse()

	dbip := os.Getenv("SANDMOD_DB_IP")
	dbport := os.Getenv("SANDMOD_DB_PORT")
	dbuname := os.Getenv("SANDMOD_DB_UNAME")
	dbupwd := os.Getenv("SANDMOD_DB_UPWD")
	dbname := os.Getenv("SANDMOD_DB_NAME")
	if dbip != "" && dbport != "" && dbuname != "" && dbupwd != "" && dbname != "" {
		dbstr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbuname, dbupwd, dbip, dbport, dbname)
		g_main._dburl = dbstr
	}
	log.Println("get dbstr:", g_main._dburl)

}

//接受http请求， 将其中的userid转成platid

func test1(w http.ResponseWriter, req *http.Request) {
	bodybuf := strings.Builder{}
	io.Copy(&bodybuf, req.Body)
	fmt.Println("test1 method:", req.Method, " body:", bodybuf.String())
}
func test(w http.ResponseWriter, req *http.Request) {
	nreq, err := http.NewRequest(req.Method, "http://localhost:8097/test1", req.Body)
	if err != nil {
		fmt.Println("http NewRequest failed:", err)
		return
	}
	//rm gameid, url,method
	req.Header.Del("url")
	req.Header.Del("gameid")
	req.Header.Del("method")
	nreq.Header = req.Header
	cli := &http.Client{}
	resp, err := cli.Do(nreq)
	if err != nil {
		fmt.Println("ERROR, cli.do:", err)
		return
	}
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	fmt.Print(buf.String())
}

func hello(w http.ResponseWriter, req *http.Request) {
	seckey := "LDXqWC6O0LMi7nVqxm5ksQcfM2PDk65M"
	url := "http://192.168.40.157:8899/friend/api/v2/inner/friends"
	url = signSand(url, seckey)

	body := `{"friendId":3669264, "msg":"","type":0}`
	nreq, err := http.NewRequest("POST", "http://localhost:8097/friendReqProxy", bytes.NewReader([]byte(body)))
	if err != nil {
		fmt.Println("http NewRequest failed:", err)
		return
	}

	nreq.Header["url"] = []string{url}
	nreq.Header["method"] = []string{"POST"}
	nreq.Header["gameid"] = []string{"109166"}

	//rm gameid, url,method
	nreq.Header["userId"] = []string{"376414"}
	nreq.Header["Content-Type"] = []string{"application/json"}
	{
		dumpstr, _ := httputil.DumpRequest(nreq, true)
		fmt.Println("hello req desc:", string(dumpstr))
	}
	cli := &http.Client{}
	resp, err := cli.Do(nreq)
	if err != nil {
		fmt.Println("ERROR, cli.do:", err)
		return
	}
	bytes, _ := io.ReadAll(resp.Body)
	fmt.Fprintf(w, string(bytes))
}

func rawhello(w http.ResponseWriter, req *http.Request) {
	seckey := "LDXqWC6O0LMi7nVqxm5ksQcfM2PDk65M"
	url := "http://192.168.40.157:8899/friend/api/v2/inner/friends"
	url = signSand(url, seckey)
	body := `{"friendId":3669264, "msg":"","type":0}`
	nreq, err := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))
	if err != nil {
		fmt.Println("http NewRequest failed:", err)
		return
	}
	//rm gameid, url,method
	nreq.Header["userId"] = []string{"376414"}
	nreq.Header["Content-Type"] = []string{"application/json"}
	req.Header.Del("Url")
	req.Header.Del("Gameid")
	req.Header.Del("Method")
	{
		dumpstr, _ := httputil.DumpRequest(nreq, true)
		fmt.Println("rawhello req desc:", string(dumpstr))
	}

	cli := &http.Client{}
	resp, err := cli.Do(nreq)
	if err != nil {
		fmt.Println("ERROR, cli.do:", err)
		return
	}
	bytes, _ := io.ReadAll(resp.Body)
	fmt.Fprintf(w, string(bytes))
}

func signSand(url string, secKey string) string {
	t := fmt.Sprintf("%v", time.Now().UnixMilli())
	n := fmt.Sprintf("%v", rand.Int31())
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v%v%v", secKey, n, t)
	s := fmt.Sprintf("%x", sha1.Sum(buf.Bytes()))
	p := fmt.Sprintf("timestamp=%v&nonce=%v&signature=%v", t, n, s)
	if !strings.Contains(url, "?") {
		return fmt.Sprintf("%s?%s", url, p)
	}
	return fmt.Sprintf("%s&%s", url, p)
}

func convertDoAllParse(key string, value string, body string, platId2UidPairs map[string]uint, isgame int) string {
	if key == "" {
		for platid, uid := range platId2UidPairs {
			uidstr := fmt.Sprintf("%d", uid)
			body = convertDoAllParse(platid, uidstr, body, platId2UidPairs, isgame)
		}
		return body
	}
	//key: platid, value: uid
	fndstr := fmt.Sprintf("\"userId\":%s,", key)

	pos := strings.Index(body, fndstr)
	if pos == -1 {
		fmt.Printf("find failed: %s", fndstr)
		return body
	}

	tostr := ""
	if isgame != 0 {
		tostr = fmt.Sprintf("\"userId\":%s,\"isgame\":%d,", value, isgame)
	} else {
		tostr = fmt.Sprintf("\"userId\":%s,\"isgame\":%d,", key, isgame)
	}
	tmpstr := strings.ReplaceAll(body, fndstr, tostr)
	return tmpstr
}

func convertPlatID2userIDAll(body string, gameid string) string {
	platids := make([]string, 0)
	platids = convertPlatID2userID(body, platids)
	platId2UidPairs := g_main._dbop.getIDPairsMapByPlatid(platids, gameid) //pl
	mp := make(map[string]bool)
	for _, v := range platids {
		mp[v] = true
	}
	missMp := make(map[string]uint)
	for k, _ := range mp {
		if _, ok := platId2UidPairs[k]; !ok {
			missMp[k] = 0
		}
	}

	{
		//debug
		// jss, _ := json.Marshal(platId2UidPairs)
		// log.Println("jss platId2UidPairs is:", string(jss))

		// jssMsMap, _ := json.Marshal(missMp)
		// log.Println("jss missMp is:", string(jssMsMap))
	}

	str := convertDoAllParse("", "", body, platId2UidPairs, 1)
	return convertDoAllParse("", "", str, missMp, 0)
}

func convertPlatID2userID(body string, platids []string) []string {
	pos := strings.Index(body, "\"userId\":")
	if pos == -1 {
		return platids
	}

	i := pos
	for {
		i++
		if body[i] == ',' {
			break
		}
		if i > len(body) {
			return platids
		}
	}

	targetStr := body[pos:i]
	platid := strings.Replace(targetStr, "\"userId\":", "", 1)
	platids = append(platids, platid)
	platids = convertPlatID2userID(body[i:], platids)
	return platids
}

func getPlatidBatchProxy(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("gameid") {
		fmt.Println("parameter gameid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	if !req.URL.Query().Has("userids") {
		fmt.Println("parameter userid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	gameid := req.URL.Query().Get("gameid")
	userids := req.URL.Query()["userids"]
	useridsAry := strings.Split(userids[0], ",")
	retslc := g_main._dbop.getIDPairsByUserid(useridsAry, gameid) //userid: platid
	if retslc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "fail")
		return
	}
	sb := strings.Builder{}
	for i, v := range retslc {
		sb.WriteString(v)
		if i != len(retslc)-1 {
			sb.WriteString(",")
		}
	}
	{
		//log
		reqjss, _ := json.Marshal(userids)
		log.Printf("getPlatidBatchProxy reqjss:", string(reqjss), sb.String())
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", sb.String())
}
func getUseridBatchProxy(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("gameid") {
		fmt.Println("parameter gameid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	if !req.URL.Query().Has("platids") {
		fmt.Println("parameter platids not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	gameid := req.URL.Query().Get("gameid")
	platids := req.URL.Query()["platids"]
	platidsAry := strings.Split(platids[0], ",")
	retslc := g_main._dbop.getIDPairsByPlatid(platidsAry, gameid) //userid: platid
	if retslc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "fail")
		return
	}
	sb := strings.Builder{}
	for i, v := range retslc {
		sb.WriteString(v)
		if i != len(retslc)-1 {
			sb.WriteString(",")
		}
	}
	{
		//log
		reqjss, _ := json.Marshal(platids)
		retjss, _ := json.Marshal(retslc)
		log.Printf("getUseridBatchProxy reqjss: %s  ret:%s", string(reqjss), string(retjss))
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", sb.String())
}

func getPlatidProxy(w http.ResponseWriter, req *http.Request) {
	jss, _ := json.Marshal(req.URL.Query())
	log.Println("getPlatidProxy params:", string(jss))
	if !req.URL.Query().Has("gameid") {
		fmt.Println("parameter gameid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	if !req.URL.Query().Has("userid") {
		fmt.Println("parameter userid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	gameid := req.URL.Query().Get("gameid")
	userid := req.URL.Query().Get("userid")
	platid := g_main._dbop.getPlatIDByUserid(userid, gameid)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", platid)
}

func getUseridProxy(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("gameid") {
		fmt.Println("parameter gameid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	if !req.URL.Query().Has("platid") {
		fmt.Println("parameter platid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	gameid := req.URL.Query().Get("gameid")
	platid := req.URL.Query().Get("platid")
	userid := g_main._dbop.getUserIDByPlatid(platid, gameid)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", userid)
}
func httpProxy(w http.ResponseWriter, req *http.Request) {

	if req.Header.Get("url") == "" {
		fmt.Println("parameter url not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	if req.Header.Get("gameid") == "" {
		fmt.Println("parameter gameid not exists!")
		fmt.Fprintf(w, "failed\n")
		return
	}
	nowurl := req.Header.Get("url")
	if req.Header.Get("seckey") != "" {
		seckey := req.Header.Get("seckey")
		nowurl = signSand(nowurl, seckey)
		nowurl = fmt.Sprintf("%s&%s", nowurl, req.URL.Query().Encode())
	}
	gameid := req.Header.Get("gameid")

	bodybuf := strings.Builder{}
	io.Copy(&bodybuf, req.Body)
	fmt.Println("call realurl: ", nowurl, " method:", req.Method, " body:", bodybuf.String())

	nreq, err := http.NewRequest(req.Method, nowurl, bytes.NewBuffer([]byte(bodybuf.String())))
	if err != nil {
		fmt.Println("http NewRequest failed:", err)
		return
	}
	//rm gameid, url,method
	req.Header.Del("Url")
	req.Header.Del("Gameid")
	req.Header.Del("Method")
	nreq.Header = req.Header
	{
		dumpstr, _ := httputil.DumpRequest(nreq, true)
		fmt.Println("req desc:", string(dumpstr))
	}
	cli := &http.Client{}
	resp, err := cli.Do(nreq)
	if err != nil {
		fmt.Println("ERROR, cli.do:", err)
		return
	}
	w.WriteHeader(resp.StatusCode)
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	result := buf.String()

	fmt.Println("raw result:", result)

	result = convertPlatID2userIDAll(result, gameid)
	fmt.Println("true result:", result)
	w.Write([]byte(result))
}

/*

curl -H 'userId: 311885' -H 'seckey: LDXqWC6O0LMi7nVqxm5ksQcfM2PDk65M' -H 'url: http://192.168.40.157:8899/friend/api/v2/inner/friends/311886' -H 'method: GET'      -X GET 'http://localhost:8097/friendReqProxy?a=1&b=2'

curl -H 'userId: 311885'
-H 'seckey: LDXqWC6O0LMi7nVqxm5ksQcfM2PDk65M'
-H 'url: http://192.168.40.157:8899/friend/api/v2/inner/friends/311886'
-H 'method: GET'
-X GET
'http://localhost:8097/friendReqProxy?a=1&b=2'
*/

func main() {

	fmt.Println("dburl:", g_main._dburl)

	g_main.init()
	{
		cont, err := os.ReadFile("./test.dat")
		if err == nil {
			ret := convertPlatID2userIDAll(string(cont), "109166")
			jss, _ := json.Marshal(ret)
			fmt.Println("jss:", string(jss))
		} else {
			fmt.Println("read fail, start direct work.")
		}
	}

	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/rawhello", rawhello)

	http.HandleFunc("/test", test)
	http.HandleFunc("/test1", test1)

	http.HandleFunc("/friendReqProxy", httpProxy)
	http.HandleFunc("/getPlatIDByUserIDProxy", getPlatidProxy)
	http.HandleFunc("/getUserIDByPlatIDProxy", getUseridProxy)

	http.HandleFunc("/getPlatIDByUserIDBatchProxy", getPlatidBatchProxy)
	http.HandleFunc("/getUserIDByPlatIDBatchProxy", getUseridBatchProxy)

	err := http.ListenAndServe(fmt.Sprintf(":%s", g_main._port), nil)
	panic(err)
}
