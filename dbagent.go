package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Dbagent struct {
	_db           *sql.DB
	_cacheMap     map[string]uint
	_cacheUserMap map[string]string
}

func (d *Dbagent) init(dburl string) {
	var err error
	d._db, err = sql.Open("mysql", dburl)
	if err != nil {
		panic(err.Error())
	}
	d._cacheMap = make(map[string]uint)
	d._cacheUserMap = make(map[string]string)
}

func (d *Dbagent) getIDPairsByUserid(userids []string, gameid string) []string {
	ret := make(map[string]string)
	sb := strings.Builder{}

	sb.WriteString("(")
	bfind := false
	for i := 0; i < len(userids); i++ {
		val, ok := d._cacheUserMap[userids[i]]
		if ok {
			ret[userids[i]] = val
			continue
		}
		bfind = true
		sb.WriteString(fmt.Sprintf("%s", userids[i]))
		sb.WriteString(",")
	}
	sb.WriteString(")")

	instr := strings.Replace(sb.String(), ",)", ")", 1)

	if bfind {
		sqlstr := fmt.Sprintf("SELECT UserID,OriginIdentity FROM Game_User_%s WHERE UserID in %s", gameid, instr)
		rows, err := d._db.Query(sqlstr)
		if err != nil {
			fmt.Println("exec sql err:", err.Error(), " sql:", sqlstr)
			return nil
		}
		defer rows.Close()

		for rows.Next() {
			var uid uint
			var platid string
			err := rows.Scan(&uid, &platid)
			if err != nil {
				fmt.Println("rows.Scan err:", err.Error())
				return nil
			}
			uidstr := fmt.Sprintf("%d", uid)
			ret[uidstr] = platid
			d._cacheUserMap[uidstr] = platid
		}
	}

	retSlc := make([]string, 0, len(ret))
	for _, uid := range userids {
		v, ok := ret[uid]
		if ok {
			retSlc = append(retSlc, v)
		} else {
			retSlc = append(retSlc, "0")
		}
	}
	return retSlc
}

func (d *Dbagent) getIDPairsByPlatid(platids []string, gameid string) []string {
	ret := d.getIDPairsMapByPlatid(platids, gameid)
	if ret == nil {
		return nil
	}
	{
		jss, _ := json.Marshal(ret)
		req, _ := json.Marshal(platids)
		log.Printf("getIDPairsByPlatid req:%s ret: %s lenof: ", string(req), string(jss), len(platids))
	}
	retSlc := make([]string, 0, len(ret))
	for _, uid := range platids {
		v, ok := ret[uid]
		if ok {
			vstr := fmt.Sprintf("%d", v)
			retSlc = append(retSlc, vstr)
			log.Printf("got uid: %s val： %d", uid, v)
		} else {
			retSlc = append(retSlc, "0")
			log.Printf("else got uid: %s val： %d", uid, v)
		}
	}
	return retSlc
}

func (d *Dbagent) getIDPairsMapByPlatid(platids []string, gameid string) map[string]uint {
	ret := make(map[string]uint)
	sb := strings.Builder{}

	sb.WriteString("(")
	bfind := false
	for i := 0; i < len(platids); i++ {
		val, ok := d._cacheMap[platids[i]]
		if ok {
			ret[platids[i]] = val
			continue
		}
		bfind = true
		sb.WriteString(platids[i])
		sb.WriteString(",")
	}
	sb.WriteString(")")

	instr := strings.Replace(sb.String(), ",)", ")", 1)

	if bfind {
		sqlstr := fmt.Sprintf("SELECT UserID,OriginIdentity FROM Game_User_%s WHERE OriginIdentity in %s", gameid, instr)
		rows, err := d._db.Query(sqlstr)
		if err != nil {
			fmt.Println("exec sql err:", err.Error(), " sql:", sqlstr)
			return nil
		}
		defer rows.Close()

		for rows.Next() {
			var uid uint
			var platid string
			err := rows.Scan(&uid, &platid)
			if err != nil {
				fmt.Println("rows.Scan err:", err.Error())
				return nil
			}
			ret[platid] = uid
			d._cacheMap[platid] = uid
		}
	}
	return ret
}

func (d *Dbagent) getPlatIDByUserid(userid string, gameid string) string {
	rows, err := d._db.Query(fmt.Sprintf("SELECT OriginIdentity FROM Game_User_%s WHERE UserID='%s'", gameid, userid))
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var uid string
		err := rows.Scan(&uid)
		if err != nil {
			panic(err.Error())
		}
		return uid
	}
	return ""
}
func (d *Dbagent) getUserIDByPlatid(platid string, gameid string) uint {
	rows, err := d._db.Query(fmt.Sprintf("SELECT UserID FROM Game_User_%s WHERE OriginIdentity='%s'", gameid, platid))
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var uid uint
		err := rows.Scan(&uid)
		if err != nil {
			panic(err.Error())
		}
		return uid
	}
	return 0
}
