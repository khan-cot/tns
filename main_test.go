package main

import (
	"database/sql"
	_ "database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"math/big"
	"testing"
	"time"
)
var(

	rUnit0 = new(big.Int).Mul(big.NewInt(200000),big.NewInt(1e8))
	rUnit1 = new(big.Int).Mul(big.NewInt(100000),big.NewInt(1e8))
	rUnit2 = new(big.Int).Mul(big.NewInt(10000),big.NewInt(1e8))
)




func Test02(t *testing.T) {
	addNodeFromOld("1","3",rUnit0)
	addNodeFromOld("2","3",rUnit1)
	addNodeFromOld("3","",rUnit2)
	addNodeFromOld("5","1",rUnit0)
	addNodeFromOld("6","1",rUnit1)

	scanNodesForLevel()
	scanNodesForReward()
}
func Test03(t *testing.T) {
	LoadNodeFromFile()
	fmt.Println("开始第一次扫描，定义级别......")
	scanNodesForLevel()
	fmt.Println("扫描结束，定义级别完成......")
	fmt.Println("第二次扫描，计算奖励......")
	scanNodesForReward()
	fmt.Println("第二次扫描结束，计算奖励完成......")
	exportNodes()
	fmt.Println("统计结果输出到文件......")
	checkNode()
}
func checkNode() {
	//addr := "TMmqNChoX2UzayPhKFxQZcofrroP5ErAeu"
	//node := findNode(addr)
	pos := 2
	node := getNodeByPos(pos)
	subs := getAllSubNodePos(node.pos)
	println("addr",node.addr,"root:",node.pos,"level",node.selfLevelT,"amount",node.selfRedeem.String(),"max",node.maxLevel)

	for _,ipos := range subs {
		node0 := getNodeByPos(ipos)
		println("addr",node0.addr,"root:",node0.pos,"level",node0.selfLevelT,"amount",node0.selfRedeem.String(),"max",node0.maxLevel)
	}
	fmt.Println(subs)
}
func printSubNode(pos int,subs []int) {
	if len(subs) == 0 {
		return
	}

	for _,ipos := range subs {
		subs0 := getAllSubNodePos(ipos)
		printSubNode(ipos,subs0)
	}

	node := getNodeByPos(pos)
	println("pos:",node.pos,"level",node.selfLevelT,"amount",node.selfRedeem)
}
func Test04(t *testing.T) {
	loadNodeFromMysql()
}
func loadNodeFromMysql() bool {
	db, err := sql.Open("mysql",
		"edgar:edgar123456@@tcp(rm-j6cblrsuzq7892wwbto.mysql.rds.aliyuncs.com.com:3306)/crs?charset=utf8&parseTime=True")
	if err != nil {
		log.Fatalf("Open database error: %s\n", err)
	}
	defer db.Close()
	fmt.Println("连接数据库成功..........")
	fmt.Println("尝试ping数据库..........")
	fmt.Println("ping数据库结束,开始查询数据..........")
	db.SetConnMaxLifetime(time.Second * 3)

	//rows, err := db.Query("select address,sup,amounts from tns_user")
	rows,err := db.Query("select COUNT(*) from tns_user ")

	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	//var addr string
	//var up string
	//var redeem int
	//for rows.Next() {
	//	err := rows.Scan(&addr, &up,&redeem)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println(addr, up,redeem)
	//}
	//err = rows.Err()
	//if err != nil {
	//	log.Fatal(err)
	//}
	return true
}