package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// Node 定义一个汉字节点
type Node struct {
	Word     string  //汉字
	Py       string  //拼音
	Emission float64 //汉字出现的概率
	MaxScore float64 //最大分数
	PreNode  *Node   //下一个汉字节点
}

// EmissionMap 读取整理好的汉字拼音数据，获取的map数据，数据格式类型为
//{
// "ni":{"你"：0.91，"尼"：0.789,...},
// "wo":{"我"：0.91，"窝"：0.189,...},
//}
// EmissionMap 读取整理好的汉字拼音数据，获取的map数据
var EmissionMap = make(map[string]map[string]float64)

// WordsArray 读取大量词语数据获取的词语数组
var WordsArray = make([]string, 1)

// FreqMap 词语出现的概率值，这里会将词语位置颠倒
// {
//    "好你"：0.789，
//    "兴高": 0.567,
//｝
var FreqMap = make(map[string]float64)

// InputSequence 输入的序列
var InputSequence = make([]map[string]*Node, 1)

// 数据缓存
var viterbiCache = make(map[string]float64)

// Pinyins 输入的已经切分的拼音数组
var Pinyins []string

// ReadPinyinData 读取词表，放入EmissionMap变量
func ReadPinyinData() {
	f, err := os.Open("test.txt")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err.Error())
			break
		}
		// line:鼥 0.750684002197 1 ba
		line = strings.TrimSpace(line)
		if len(line) <= 1 {
			continue
		}

		PinyinArray := strings.Split(line, " ")
		// 拿到汉字
		word := PinyinArray[0]
		f, _ := strconv.ParseFloat(PinyinArray[1], 64)

		// 权重
		em := math.Atan(f) / (math.Pi / 2)

		// 拿到拼音 py:ba
		py := PinyinArray[len(PinyinArray)-1]
		if len(strings.Split(word, "")) > 1 {
			continue
		}
		if EmissionMap[py] == nil {
			EmissionMap[py] = make(map[string]float64)
		}

		// 这个拼音下、这个字、权重是多少
		EmissionMap[py][word] = em
		// fmt.Println(EmissionMap)
		//		for k, _ := range EmissionMap[py] {
		//			EmissionMap[py][k] = 1 / float64(len(EmissionMap[py]))
		//		}

		pyArray := strings.Split(py, "")
		py = pyArray[0]
		if len(strings.Split(word, "")) > 1 {
			break
		}
		if EmissionMap[py] == nil {
			EmissionMap[py] = make(map[string]float64)
		}

		// 取这个拼音的首字母存一份
		EmissionMap[py][word] = em
	}
	// js, _ := json.MarshalIndent(EmissionMap, "", "\t")
	// fmt.Println(string(js))
}
func readWords() {
	f, err := os.Open("RenMinData.txt")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err.Error())
			break
		}
		line = strings.TrimSpace(line)
		if len(line) <= 1 {
			continue
		}
		WordsArray = append(WordsArray, "<s>")
		words := strings.Split(line, "")
		for j := 0; j < len(words); j++ {
			WordsArray = append(WordsArray, strings.TrimSpace(words[j]))
		}
		WordsArray = append(WordsArray, "</s>")

	}

	WordsArray = WordsArray[1:]

	for k := range WordsArray {
		i := k

		key := ""
		for j := i; (i-j < 6) && (j >= 0); j-- {
			key += WordsArray[j]
			FreqMap[key] = FreqMap[key] + 1

		}

	}
}
func gettransprop(args ...string) float64 {
	key := ""
	for _, v := range args {
		key += v
	}

	c2 := float64(FreqMap[key] + 1.0)
	c1 := float64(len(WordsArray))

	return c2 / c1
}

func getInitProp(word string) float64 {
	return gettransprop(word, "<s>")
}

// GetInputSequence 获取输入的拼音
func GetInputSequence(pys []string) {
	Pinyins = pys
	for _, v := range Pinyins {
		if EmissionMap[v] == nil {
			continue
		}
		mymap := make(map[string]*Node)
		// EmissionMap 一维Key为拼音，二维key为汉字
		for w, r := range EmissionMap[v] {
			mymap[w] = &Node{Word: w, Emission: r, Py: v}
		}
		InputSequence = append(InputSequence, mymap)
	}
	fmt.Printf("%q\n", InputSequence)
	InputSequence = InputSequence[1:]
}
func getKey(t int, k string) string {
	return strings.Join([]string{strconv.Itoa(t), k}, "_")
}

func viterbi(t int, k string) float64 {
	fmt.Printf("t:%v,k:%v\n", t, k)
	if viterbiCache[getKey(t, k)] != 0 {
		return viterbiCache[getKey(t, k)]
	}
	node := InputSequence[t][k]
	if t == 0 {
		stateTransfer := getInitProp(k)
		emissionProp := EmissionMap[Pinyins[t]][k]

		node.MaxScore = 1.0 * stateTransfer * emissionProp
		viterbiCache[getKey(t, k)] = node.MaxScore
		return node.MaxScore
	}
	n := t - 1

	for i, v := range InputSequence[n] {
		fmt.Printf("k=%v,i=%v\n", k, v)
		stateTransfer := gettransprop(k, i)
		emissionProp := EmissionMap[Pinyins[n]][i]
		if len(Pinyins)-1 == t {
			emissionProp *= EmissionMap[Pinyins[t]][k]
		}

		score := viterbi(n, i) * stateTransfer * emissionProp
		if score > node.MaxScore {
			node.MaxScore = score
			node.PreNode = v
		}
	}

	viterbiCache[getKey(t, k)] = node.MaxScore
	return node.MaxScore
}

// Translate 整合输入的拼音
func Translate(pys []string) {
	GetInputSequence(pys)
	// 使用viterbi算法求解最大路径
	var maxNode *Node
	maxScore := 0.0
	for word, node := range InputSequence[len(InputSequence)-1] {
		fmt.Println("**************************")
		score := viterbi(len(pys)-1, word)
		if score > maxScore {
			maxScore = score
			maxNode = node
		}
	}
	fmt.Println("==============================")
	fmt.Println(len(pys))
	// 回溯输出最大路径
	results := make([]string, 1)
	for {
		results = append(results, maxNode.Word)
		if maxNode.PreNode != nil {
			maxNode = maxNode.PreNode
		} else {
			break
		}

	}
	results = results[1:]
	for i := len(results) - 1; i >= 0; i-- {
		fmt.Print(results[i])
	}

}

func main() {
	// s := "我是中国人"
	// fmt.Println(strings.Split(s, " "))
	// fmt.Println(len(strings.Split(s, " ")))
	ReadPinyinData()
	// readWords()
	Translate([]string{"bai", "du"})
}
