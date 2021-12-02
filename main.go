
package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)
type Stack []interface {}

func (stack Stack) Len() int {
	return len(stack)
}
func (stack Stack) IsEmpty() bool  {
	return len(stack) == 0
}
func (stack Stack) Cap() int {
	return cap(stack)
}
func (stack *Stack) Push(value interface{})  {
	*stack = append(*stack, value)
}
func (stack Stack) Top() (interface{}, error)  {
	if len(stack) == 0 {
		return nil, errors.New("Out of index, len is 0")
	}
	return stack[len(stack) - 1], nil
}
func (stack *Stack) Pop() (interface{}, error)  {
	theStack := *stack
	if len(theStack) == 0 {
		return nil, errors.New("Out of index, len is 0")
	}
	value := theStack[len(theStack) - 1]
	*stack = theStack[:len(theStack) - 1]
	return value, nil
}

//===================================================================
var(
	MAXLEVEL = 100
	unit0 = big.NewInt(1e8)
	T1Redeem0 = new(big.Int).Mul(big.NewInt(400000),unit0)
	T1Redeem1 = new(big.Int).Mul(big.NewInt(120000),unit0)
	RedeemRate = 20
)


func levelT2Rate(l int) int {
	if l == 0 {
		return 0
	}
	switch l {
	case 1:
		return 5
	case 2:
		return 7
	case 3:
		return 9
	case 4:
		return 10
	case 5:
		return 10
	default:
		panic(fmt.Errorf("invalid levelT,%d",l))
	}
}
func levelT2Redeem(l int) *big.Int {
	if l == 0 {
		return big.NewInt(0)
	}
	switch l {
	case 1:
		return new(big.Int).Mul(big.NewInt(10000),unit0)
	case 2:
		return new(big.Int).Mul(big.NewInt(20000),unit0)
	case 3:
		return new(big.Int).Mul(big.NewInt(30000),unit0)
	case 4:
		return new(big.Int).Mul(big.NewInt(30000),unit0)
	case 5:
		return new(big.Int).Mul(big.NewInt(30000),unit0)
	default:
		panic(fmt.Errorf("invalid levelT,%d",l))
	}
}
func layer2Rate(l int) int {
	switch l {
	case 1:
		return 300
	case 2:
		return 60
	case 3,4,5:
		return 25
	case 6,7,8,9,10:
		return 10
	default:
		return 0
	}
}
//===================================================================

type Node struct {
	pos 				int
	up					int
	down 				[]int
	addr 				string

	done				bool
	calcRewardDone	 	bool

	selfRedeem			*big.Int
	allSubRedeem		*big.Int
	maxSubRedeem		*big.Int

	selfLevelT			int

	allRedeemT			[]*TRewardUnit 				// 6 count
	rewardT				*big.Int
	rewardL				*big.Int
	maxLevel			int
}
type TRewardUnit struct {
	level			int
	allRedeemT 		*big.Int
}
func addRedeemT(a []*TRewardUnit,b []*TRewardUnit) []*TRewardUnit {
	if len(a) != len(b) || len(a) != 6 {
		panic(fmt.Errorf("invalid allRedeemT count,%d,%d",len(a),len(b)))
	}
	left := make([]*TRewardUnit,6,6)
	for i:=0;i<6;i++ {
		if a[i].level != b[i].level || a[i].level != i {
			panic(fmt.Errorf("0 invalid allRedeemT count,%d,%d",a[i].level,b[i].level))
		}
		left[i] = &TRewardUnit{
			level: i,
			allRedeemT: new(big.Int).Add(a[i].allRedeemT,b[i].allRedeemT),
		}
	}
	return left
}
//===================================================================
func calcAndSetRewardForT(pos int) {
	node := getNodeByPos(pos)
	subs := getAllSubNodePos(pos)
	for _,ipos := range subs {
		level := getNodeByPos(ipos).selfLevelT
		t0 := getNodeByPos(ipos).allRedeemT
		if level == 0 {
			node.allRedeemT= addRedeemT(node.allRedeemT,t0)
		} else {
			if node.selfLevelT > level {
				downRedeem := big.NewInt(0)
				for i:=0;i<level;i++ {
					downRedeem = new(big.Int).Add(downRedeem,t0[i].allRedeemT)
				}
				node.allRedeemT[level].allRedeemT = new(big.Int).Add(node.allRedeemT[level].allRedeemT,downRedeem)
			}
		}
	}
}
func baseMakeNodeToCommon1(pos int) {
	allSubRedeem,maxSubRedeem := calcRedeemInfo(pos)
	node := getNodeByPos(pos)
	node.allSubRedeem = new(big.Int).Add(node.selfRedeem,allSubRedeem)
	node.maxSubRedeem = new(big.Int).Set(maxSubRedeem)
	node.selfLevelT = 0
	node.allRedeemT = make([]*TRewardUnit,6,6)
	node.rewardT = big.NewInt(0)
	node.rewardL = big.NewInt(0)
	node.maxLevel = getSubMaxLevel(pos)
	for i,_ := range node.allRedeemT {
		node.allRedeemT[i] = &TRewardUnit{
			level: i,
			allRedeemT: big.NewInt(0),
		}
	}
}
func makeLeafNode(pos int) {
	node := getNodeByPos(pos)
	if !node.done {
		panic(fmt.Errorf("node not done,pos=%d",pos))
	}
	node.allSubRedeem = new(big.Int).Set(node.selfRedeem)
	node.maxSubRedeem = new(big.Int).Set(node.selfRedeem)
	node.selfLevelT = 0
	node.allRedeemT = make([]*TRewardUnit,6,6)
	node.rewardT = big.NewInt(0)
	node.rewardL = big.NewInt(0)
	node.maxLevel = 0
	for i,_ := range node.allRedeemT {
		if i == 0{
			node.allRedeemT[i] = &TRewardUnit{
				level: 0,
				allRedeemT: new(big.Int).Set(node.selfRedeem),
			}
		} else {
			node.allRedeemT[i] = &TRewardUnit{
				level: i,
				allRedeemT: big.NewInt(0),
			}
		}
	}
}
func makeNodeToT5(pos int,calc bool) {
	if !calc {
		node := getNodeByPos(pos)
		baseMakeNodeToCommon1(pos)
		node.selfLevelT = 5
		if node.selfLevelT > getSubMaxLevel(pos) {
			node.maxLevel = node.selfLevelT
		}
	} else {
		calcAndSetRewardForT(pos)
	}
}
func makeNodeToT4(pos int,calc bool) {
	if !calc {
		node := getNodeByPos(pos)
		baseMakeNodeToCommon1(pos)
		node.selfLevelT = 4
		if node.selfLevelT > getSubMaxLevel(pos) {
			node.maxLevel = node.selfLevelT
		}
	} else {
		calcAndSetRewardForT(pos)
	}
}
func makeNodeToT3(pos int,calc bool) {
	if !calc {
		node := getNodeByPos(pos)
		baseMakeNodeToCommon1(pos)
		node.selfLevelT = 3
		if node.selfLevelT > getSubMaxLevel(pos) {
			node.maxLevel = node.selfLevelT
		}
	} else {
		calcAndSetRewardForT(pos)
	}
}
func makeNodeToT2(pos int,calc bool) {
	if !calc {
		node := getNodeByPos(pos)
		baseMakeNodeToCommon1(pos)
		node.selfLevelT = 2
		if node.selfLevelT > getSubMaxLevel(pos) {
			node.maxLevel = node.selfLevelT
		}
	} else {
		calcAndSetRewardForT(pos)
	}
}
func makeNodeToT1(pos int,calc bool) {
	if !calc {
		node := getNodeByPos(pos)
		baseMakeNodeToCommon1(pos)
		node.selfLevelT = 1
		if node.selfLevelT > getSubMaxLevel(pos) {
			node.maxLevel = node.selfLevelT
		}
	} else {
		calcAndSetRewardForT(pos)
	}
}
func makeNodeToCommon(pos int,calc bool) {
	// level must be 0
	if !calc {
		baseMakeNodeToCommon1(pos)
	} else {
		node := getNodeByPos(pos)
		subs := getAllSubNodePos(pos)
		for _,ipos := range subs {
			level := getNodeByPos(ipos).selfLevelT
			t0 := getNodeByPos(ipos).allRedeemT
			if level > node.selfLevelT {
				// the Nearest ancestor
				uplevel := getNearestAncestorLevel(pos)
				if uplevel > level {
					downRedeem := big.NewInt(0)
					for i:=0;i<level;i++ {
						downRedeem = new(big.Int).Add(downRedeem,t0[i].allRedeemT)
					}
					node.allRedeemT[level].allRedeemT = new(big.Int).Add(node.allRedeemT[level].allRedeemT,downRedeem)
				}
			}
			if node.selfLevelT == level {
				node.allRedeemT= addRedeemT(node.allRedeemT,t0)
			}
		}
	}
}

//===================================================================
var allNodes []*Node
var mapNodes map[string]int
func init() {
	allNodes = make([]*Node,0,0)
	mapNodes = make(map[string]int)
}

func main() {
	LoadNodeFromFile()
	fmt.Println("开始第一次扫描，定义级别......")
	scanNodesForLevel()
	fmt.Println("扫描结束，定义级别完成......")
	fmt.Println("第二次扫描，计算奖励......")
	//scanNodesForReward()
	fmt.Println("第二次扫描结束，计算奖励完成......")
	exportNodes()
	fmt.Println("统计结果输出到文件......")
	time.Sleep(time.Second*30)
}

//===================================================================

func addNode(n *Node) {
	allNodes = append(allNodes,n)
	mapNodes[n.addr] = n.pos
}
func makeNode(addr string,pos int,redeem *big.Int) *Node {
	return &Node{
		addr: addr,
		pos: pos,
		selfRedeem: new(big.Int).Set(redeem),
		up: -1,
		down: make([]int,0,0),
		done: false,
		calcRewardDone: false,
	}
}
func findNode(addr string) *Node {
	pos,ok := mapNodes[addr]
	if ok {
		return getNodeByPos(pos)
	}
	return nil
}
func isSubNode(parent int,sub int) bool {
	subs := getAllSubNodePos(parent)
	for _,ipos := range subs {
		if ipos == sub {
			return true
		}
	}
	return false
}
func addNodeFromOld(addr string,up string,redeem *big.Int) {
	addr = strings.Trim(addr," ")
	up = strings.Trim(up," ")
	if len(addr) == 0 {
		panic("address is null")
	}
	node0,node1 := findNode(addr),findNode(up)
	pos := len(allNodes)
	oldup := -1
	if node0 == nil {
		node0 = makeNode(addr,pos,redeem)
		addNode(node0)
	} else {
		// update the redeem
		node0.selfRedeem = new(big.Int).Set(redeem)
		oldup = node0.up
	}
	if len(up) > 0 {
		pos = len(allNodes)
		if node1 == nil {
			node1 = makeNode(up,pos,big.NewInt(0))
			addNode(node1)
		}
		if oldup == -1 {
			node0.up = node1.pos
		}
		// check
		if node0.up != node1.pos || node0.up == node0.pos {
			panic(fmt.Errorf("addNodeFromOld error,node0.up:%d,node1.pos:%d",node0.up,node1.pos))
		}
		checkLoop(node0.pos)
		if isSubNode(node1.pos, node0.pos) {
			panic(fmt.Errorf("node1 has same sub node"))
		}
		node1.down = append(node1.down,node0.pos)
	}
}
func checkLoop(pos int) {
	ups := make(map[int]int)
	fix := pos
	for {
		if _,ok := ups[fix]; ok {
			fmt.Println("[",printkey(ups),"]","key",fix)
			panic("loops.....")
		}
		node := getNodeByPos(pos)
		if node.up == -1 {
			break
		}
		pos = node.up
		ups[pos] = 0
	}
}
func printkey(keys map[int]int) string {
	str := ""
	for k,_ := range keys {
		str = fmt.Sprintf("%s,%d",str,k)
	}
	return str
}
//===================================================================
func hasSubNode(pos int) bool {
	sub := getAllSubNodePos(pos)
	return  len(sub) > 0
}
func getNodeByPos(pos int) *Node {
	if pos >= 0 && pos < len(allNodes) {
		return allNodes[pos]
	} else {
		panic(fmt.Errorf("getNodeByPos failed,i=%d,all=%d",pos,len(allNodes)))
	}
	return nil
}
func subNodeAllDone(pos int,calc bool) bool {
	b := true
	sub := getAllSubNodePos(pos)
	for _, ipos := range sub {
		if !isNodeDone(ipos,calc) {
			b = false
			break
		}
	}
	return  b
}
func isNodeDone(pos int,calc bool) bool {
	node := getNodeByPos(pos)
	if calc {
		return node.calcRewardDone
	}
	return node.done
}
func getAllSubNodePos(pos int) []int {
	node := getNodeByPos(pos)
	return  node.down
}
func getSubMaxLevel(pos int) int {
	subs := getAllSubNodePos(pos)
	max := 0
	for _,ipos := range subs {
		node := getNodeByPos(ipos)
		if node.maxLevel > max {
			max = node.maxLevel
		}
	}
	return max
}
func getMaxLevel(pos int) int {
	//max := getSubMaxLevel(pos)
	node := getNodeByPos(pos)
	return node.maxLevel
}
func getAllMoreThanLevelT(pos int,l int) int {
	// get all subs
	count := 0
	subs := getAllSubNodePos(pos)
	for _,ipos := range subs {
		level := getMaxLevel(ipos)
		if level >= l {
			count++
		}
		//node := getNodeByPos(ipos)
		//if node.selfLevelT >= l {
		//	count++
		//}
	}
	return count
}
func getNodeRedeem(pos int) *big.Int {
	node := getNodeByPos(pos)
	return node.selfRedeem
}
func calcRedeemInfo(pos int) (*big.Int,*big.Int) {
	// return allSubRedeem,maxSubRedeem
	subs := getAllSubNodePos(pos)
	allSubRedeem,maxSubRedeem := big.NewInt(0),big.NewInt(0)

	for _,ipos := range subs {
		node := getNodeByPos(ipos)
		allSubRedeem = new(big.Int).Add(allSubRedeem,node.allSubRedeem)
		if node.allSubRedeem.Cmp(maxSubRedeem) > 0 {
			maxSubRedeem = new(big.Int).Set(node.allSubRedeem)
		}
	}
	return allSubRedeem,maxSubRedeem
}
func getNearestAncestorLevel(pos int) int {

	var mystack Stack
	mystack.Push(pos)

	for ;mystack.IsEmpty(); {
		val,_ := mystack.Pop()
		curPos := val.(int)

		upLevel := getNodeByPos(curPos).selfLevelT
		if upLevel > 0 {
			return  upLevel
		} else {
			up := getNodeByPos(curPos).up
			if up == -1 {
				break
			}
			mystack.Push(up)
		}
	}
	return 0
}
func getAllSubRedeem(pos int) *big.Int {
	subs := getAllSubNodePos(pos)
	all := big.NewInt(0)
	for _,ipos := range subs {
		all = new(big.Int).Add(all,getNodeByPos(ipos).selfRedeem)
	}
	return all
}
func getSubLayer(allpos []int) ([]int,*big.Int) {
	layers := make([]int,0,0)
	all := big.NewInt(0)
	for _,ipos := range allpos {
		all = new(big.Int).Add(all,getAllSubRedeem(ipos))
		layers = append(layers,getAllSubNodePos(ipos)...)
	}
	return  layers,all
}
func getLayerAllRedeem(pos int,l int) []*big.Int {
	if l > 0 {
		array := make([][]int,l,l)
		array[0] = []int{pos}
		all := make([]*big.Int,l,l)
		ipos := 0
		for i:=0;i<l;i++ {
			if i == 0 {
				ipos = i
			} else {
				ipos = i -1
			}
			a,b := getSubLayer(array[ipos])
			array[i] = a
			all[i] = new(big.Int).Set(b)
		}
		return all
	}
	return nil
}
//===================================================================
func nodeIsT5(pos int) bool{
	// >= T4*3
	count := getAllMoreThanLevelT(pos,4)
	redeem := levelT2Redeem(5)
	return count >= 3 && getNodeRedeem(pos).Cmp(redeem) >= 0
}
func nodeIsT4(pos int) bool{
	// >= T3*3
	count := getAllMoreThanLevelT(pos,3)
	redeem := levelT2Redeem(4)
	return count >= 3 && getNodeRedeem(pos).Cmp(redeem) >= 0
}
func nodeIsT3(pos int) bool{
	// >= T2*2
	count := getAllMoreThanLevelT(pos,2)
	redeem := levelT2Redeem(3)
	return count >= 2 && getNodeRedeem(pos).Cmp(redeem) >= 0
}
func nodeIsT2(pos int) bool{
	// >= T1*2
	count := getAllMoreThanLevelT(pos,1)
	redeem := levelT2Redeem(2)
	return count >= 2 && getNodeRedeem(pos).Cmp(redeem) >= 0
}
func nodeIsT1(pos int) bool{
	allSubRedeem,maxSubRedeem := calcRedeemInfo(pos)
	leftRedeem := new(big.Int).Sub(allSubRedeem,maxSubRedeem)
	redeem := levelT2Redeem(1)

	return getNodeRedeem(pos).Cmp(redeem) >= 0 &&
		allSubRedeem.Cmp(T1Redeem0) >= 0 && leftRedeem.Cmp(T1Redeem1) >= 0
}


func checkNodeLevelTorReward(pos int,calc bool){
	// T5->T4->T3->T2->T1
	if !nodeIsT5(pos) {
		if !nodeIsT4(pos) {
			if !nodeIsT3(pos) {
				if !nodeIsT2(pos) {
					if !nodeIsT1(pos) {
						// node is common node
						makeNodeToCommon(pos,calc)
					} else {
						// node is T1
						makeNodeToT1(pos,calc)
					}
				} else {
					// node is T2
					makeNodeToT2(pos,calc)
				}
			} else {
				// node is T3
				makeNodeToT3(pos,calc)
			}
		} else {
			// node is T4
			makeNodeToT4(pos,calc)
		}
	} else {
		// node is T5
		makeNodeToT5(pos,calc)
	}
}
func checkNodeRewardT(pos int) {
	node := getNodeByPos(pos)
	if node.selfLevelT > 0 {
		rewardT := big.NewInt(0)
		allRedeemT := getNodeByPos(pos).allRedeemT
		for i:=0;i<6;i++ {
			if node.selfLevelT > i {
				unit := new(big.Int).Div(new(big.Int).Mul(allRedeemT[i].allRedeemT,big.NewInt(int64(RedeemRate))),big.NewInt(10000))
				rate1,rate2 := levelT2Rate(node.selfLevelT),levelT2Rate(i)
				dot := int64(rate1 - rate2)
				rewardT = new(big.Int).Add(rewardT,new(big.Int).Div(new(big.Int).Mul(unit,big.NewInt(dot)),big.NewInt(100)))
			}
		}
		node.rewardT = new(big.Int).Set(rewardT)
	}
}

func finalHandleForTorReward(pos int,calc bool) {
	node := getNodeByPos(pos)
	if !node.done {
		panic(fmt.Errorf("node not done,pos=%d",pos))
	}
	// define the node's level
	checkNodeLevelTorReward(pos,calc)
	// calc the node's reward include nonT
	if calc {
		checkNodeRewardT(pos)
	}
}
func finalHandleForL(pos int) {
	node := getNodeByPos(pos)
	lReward := big.NewInt(0)
	subs := getAllSubNodePos(pos)
	max := len(subs)
	if max > 10 {
		max = 10
	}
	redeems := getLayerAllRedeem(pos,max)
	for i,v := range redeems {
		unit := new(big.Int).Div(new(big.Int).Mul(v,big.NewInt(int64(RedeemRate))),big.NewInt(10000))
		rate := int64(layer2Rate(i+1))
		lReward = new(big.Int).Add(lReward,new(big.Int).Div(new(big.Int).Mul(unit,big.NewInt(rate)),big.NewInt(1000)))
	}
	node.rewardL = new(big.Int).Set(lReward)
}
//===================================================================

func scanNodesForLevel() {
	count,oldcount,allcount := 0,0,len(allNodes)
	for i:=0;i< len(allNodes);i++ {
		var mystack Stack
		mystack.Push(i)

		for {
			if mystack.IsEmpty() {
				break
			}
			val,_ := mystack.Pop()
			curPos := val.(int)
			if count != oldcount && 0 == count % 100 {
				println("scan0 count ",count,"all",allcount)
				oldcount = count
			}

			if !hasSubNode(curPos) {
				leaf := getNodeByPos(curPos)
				if !leaf.done {
					leaf.done = true
					makeLeafNode(curPos)
					count++
				}
			}
			if subNodeAllDone(curPos,false) {
				node := getNodeByPos(curPos)
				if !node.done {
					node.done = true
					finalHandleForTorReward(curPos,false)
					count++
				}
			} else {
				mystack.Push(curPos)
				subPos := getAllSubNodePos(curPos)
				for _, ipos := range subPos {
					if !isNodeDone(ipos,false) {
						mystack.Push(ipos)
					}
				}
			}
		}
	}
}
func scanNodesForReward() {
	count,oldcount,allcount := 0,0,len(allNodes)
	for i:=0;i< len(allNodes);i++ {
		var mystack Stack
		mystack.Push(i)

		for {
			if mystack.IsEmpty() {
				break
			}
			val,_ := mystack.Pop()
			curPos := val.(int)
			if count != oldcount && 0 == count % 100 {
				println("scan1 count ",count,"all",allcount)
				oldcount = count
			}

			if !hasSubNode(curPos) {
				leaf := getNodeByPos(curPos)
				if !leaf.calcRewardDone {
					count++
				}
				leaf.calcRewardDone = true
			}
			if subNodeAllDone(curPos,true) {
				node := getNodeByPos(curPos)
				if !node.calcRewardDone {
					node.calcRewardDone = true
					finalHandleForTorReward(curPos,true)
					finalHandleForL(curPos)
					count++
				}
			} else {
				mystack.Push(curPos)
				subPos := getAllSubNodePos(curPos)
				for _, ipos := range subPos {
					if !isNodeDone(ipos,true) {
						mystack.Push(ipos)
					}
				}
			}
		}
	}
}
func handleRow(record []string) {
	up := record[0]
	addr := record[1]
	value := record[2]
	value = strings.Trim(value," ")
	r,b := new(big.Float).SetString(value)
	if !b {
		panic(fmt.Errorf("load error,:%s",value))
	}
	r1,_ := new(big.Float).Mul(r,new(big.Float).SetInt(unit0)).Int64()
	redeem := big.NewInt(r1)
	addNodeFromOld(addr,up,redeem)
}
func LoadNodeFromFile() {
	fileName := "tnsin.csv"
	fmt.Println("准备读取文件.....文件名:",fileName)
	fs, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	fmt.Println("加载文件.....")
	pos := 0
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			log.Fatalf("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}
		if pos > 0 {
			handleRow(row)
		}
		pos++
		//fmt.Println(row)
	}
	fmt.Println("加载文件结束.....，读取",pos,"条记录")
}
func exportNodes() {
	datas := make([][]string,0,0)
	datas = append(datas,[]string{
		"上级地址","自己地址","自己总数","自己总数0","总算力",
	})
	unit1 := new(big.Float).SetInt(unit0)
	for k,v := range mapNodes {
		node := getNodeByPos(v)
		up := ""
		if node.up != -1 {
			up = getNodeByPos(node.up).addr
		}
		redeem0 := new(big.Float).Quo(new(big.Float).SetInt(node.selfRedeem),unit1)
		all := new(big.Int).Sub(new(big.Int).Set(node.allSubRedeem),new(big.Int).Set(node.selfRedeem))

		datas = append(datas,[]string {
			up,k,redeem0.Text('f',4),node.selfRedeem.String(),all.String(),
		})
	}
	WriteCsv(datas)
}
func WriteCsv(data [][]string) {
	//创建文件
	f, err := os.Create("tnsout.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// 写入UTF-8 BOM
	f.WriteString("\xEF\xBB\xBF")
	//创建一个新的写入文件流
	w := csv.NewWriter(f)
	//写入数据
	w.WriteAll(data)
	w.Flush()
}