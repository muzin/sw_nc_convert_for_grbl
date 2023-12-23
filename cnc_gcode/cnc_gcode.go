package cnc_gcode

import (
	"github.com/muzin/go_rt/collection/array_list"
	"github.com/muzin/go_rt/collection/hash_map"
	"github.com/muzin/go_rt/interator/arrays"
	"github.com/muzin/go_rt/try"
	"os"
	"strconv"
	"strings"
)

var CncGcodeParseException = try.DeclareException("CncGcodeParseException")

type CncGcodeFile struct {
	head *CncGcodeCommand
	tail *CncGcodeCommand
}

// 解析
func Parse(path string) *CncGcodeFile {
	fileData, err := os.ReadFile(path)
	if err != nil {
		try.Throw(CncGcodeParseException.NewThrow("文件解析失败. " + err.Error()))
	}
	cncGcodeFile := ParseToCncGcodeFile(fileData)
	return cncGcodeFile
}

func (this *CncGcodeFile) GetHead() *CncGcodeCommand {
	return this.head
}

func (this *CncGcodeFile) GetTail() *CncGcodeCommand {
	return this.tail
}

func (this *CncGcodeFile) GetByLineNumber(lineNumber int) *CncGcodeCommand {

	gcodeCommand := this.head
	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {
		if gcodeCommand.LineNumber == lineNumber {
			return gcodeCommand
		}
	}
	return nil
}

func (this *CncGcodeFile) AddCncGcodeCommand(cmd *CncGcodeCommand, args ...interface{}) {
	if this.head == nil && this.tail == nil {
		this.head = cmd
		this.tail = cmd
	} else {
		tail := this.GetTail()
		tail.AddToBack(cmd)
		this.tail = cmd
	}
}

// 解析
func ParseToCncGcodeFile(data []byte) *CncGcodeFile {

	cncGcodeFile := &CncGcodeFile{}

	str := string(data)

	strsplit := strings.Split(str, "\n")

	for i := 0; i < len(strsplit); i++ {
		strsplit[i] = strings.TrimSpace(strsplit[i])
	}

	for i := 0; i < len(strsplit); i++ {
		gcodeSlice := strsplit[i]
		cncGcodeCommand := cncGcodeFile.ParseToCncGcodeCommand(gcodeSlice)
		cncGcodeFile.AddCncGcodeCommand(cncGcodeCommand)
	}
	return cncGcodeFile
}

func (this *CncGcodeFile) ParseToCncGcodeCommand(gcodeSlice string) *CncGcodeCommand {

	valid_gcode := strings.HasPrefix(gcodeSlice, "N")

	cncGcodeCommand := NewCncGcodeCommand()

	if valid_gcode {

		// 跳过 行号
		firstSpaceIndex := strings.Index(gcodeSlice, " ")

		getLineNumStrSlice := gcodeSlice[0:firstSpaceIndex]
		getGcodeSpliceStrSlice := strings.TrimSpace(gcodeSlice[firstSpaceIndex:])

		lineNumStr := strings.TrimSpace(getLineNumStrSlice)[1:]
		lineNum, _ := strconv.Atoi(lineNumStr)

		cncGcodeCommand.LineNumber = lineNum
		cncGcodeCommand.LineNumberStr = lineNumStr
		cncGcodeCommand.Codes = cncGcodeCommand.ParseCncWord(getGcodeSpliceStrSlice)

	} else {
		cncGcodeCommand.LineNumber = -1
		cncGcodeCommand.Codes = cncGcodeCommand.ParseCncWord(gcodeSlice, true)
		cncGcodeCommand.IsComment = true
	}

	return cncGcodeCommand

}

// 解析 cnc 指令字
// @param str 指令字字符串
// @param isComment 第二个参数为是否 为描述
// 返回 cnc 指令字 集合
func (*CncGcodeCommand) ParseCncWord(str string, args ...interface{}) *array_list.ArrayList {

	isComment := false
	if len(args) > 0 {
		if args[0] != nil {
			isComment = args[0].(bool)
		}
	}

	list := array_list.NewArrayList()

	if isComment {
		cncWord := NewCncWord("")
		cncWord.IsWord = false
		cncWord.Comment = str
		list.Add(cncWord)
		return list
	}

	// 解析 指令字
	cncWordStrSplit := strings.Split(str, " ")
	for i := 0; i < len(cncWordStrSplit); i++ {
		cncWordStrItem := cncWordStrSplit[i]

		cncWord := NewCncWord(cncWordStrItem)
		list.Add(cncWord)
	}

	return list
}

type CncGcodeCommand struct {
	LineNumber int

	LineNumberStr string

	Codes *array_list.ArrayList // []CncWord

	IsComment bool

	Last *CncGcodeCommand

	Next *CncGcodeCommand
}

func NewCncGcodeCommand() *CncGcodeCommand {
	return &CncGcodeCommand{}
}

func (this *CncGcodeCommand) AddToFront(cmd *CncGcodeCommand) {
	cmd.Last = this.Last
	cmd.Next = this

	if this.Last != nil {
		this.Last.Next = cmd
	}
	this.Last = cmd
}

func (this *CncGcodeCommand) AddToBack(cmd *CncGcodeCommand) {
	cmd.Last = this
	cmd.Next = this.Next

	if this.Next != nil {
		this.Next.Last = cmd
	}
	this.Next = cmd
}

func (this *CncGcodeCommand) AddCncWord(word *CncWord) {
	this.Codes.Add(word)
}

func (this *CncGcodeCommand) RemoveCncWordByStr(str string) {
	var indexs []int
	for i := 0; i < this.Codes.Size(); i++ {
		cncWord := this.Codes.Get(i).(*CncWord)
		if cncWord.String() == str {
			indexs = append(indexs, i)
		}
	}

	for i := len(indexs) - 1; i >= 0; i-- {
		this.Codes.Remove(i)
	}

	this.Codes.Get(0)
}

func (this *CncGcodeCommand) GetLastCncWordByWord(word string) *CncWord {
	for i := this.Codes.Size() - 1; i >= 0; i-- {
		cncWord := this.Codes.Get(i).(*CncWord)
		if cncWord.Word == word {
			itemPtr := this.Codes.Get(i)
			item := itemPtr.(*CncWord)
			return item
		}
	}
	return nil
}

func (this *CncGcodeCommand) RemoveLastCncWordByWord(word string) *CncWord {
	for i := this.Codes.Size() - 1; i >= 0; i-- {
		cncWord := this.Codes.Get(i).(*CncWord)
		if cncWord.Word == word {
			itemPtr := this.Codes.Remove(i)
			item := itemPtr.(*CncWord)
			return item
		}
	}
	return nil
}

func (this *CncGcodeCommand) GetCncWordLength() int {
	return this.Codes.Size()
}

func (this *CncGcodeCommand) GetCncWordByIndex(index int) *CncWord {
	itemPtr := this.Codes.Get(index)
	item := itemPtr.(*CncWord)
	return item
}

func (this *CncGcodeCommand) PopCncWord() *CncWord {
	codesLength := this.Codes.Size()
	if codesLength > 0 {
		itemPtr := this.Codes.Remove(codesLength - 1)
		item := itemPtr.(*CncWord)
		return item
	} else {
		return nil
	}
}

func (this *CncGcodeCommand) ExistsCncWord(str string) bool {
	for i := 0; i < this.Codes.Size(); i++ {
		cncWord := this.Codes.Get(i).(*CncWord)
		if strings.TrimSpace(str) == cncWord.String() {
			return true
		}
	}
	return false
}

func (this *CncGcodeCommand) FlushLineNumber(lineNumber int) {
	if this.LineNumber >= 0 {
		this.LineNumber = lineNumber
		if this.Next != nil {
			this.Next.FlushLineNumber(this.LineNumber + 1)
		}
	} else {
		if this.Next != nil {
			this.Next.FlushLineNumber(lineNumber)
		}
	}

}

func (this *CncGcodeCommand) GetLineNumber() int {
	return this.LineNumber
}

func (this *CncGcodeCommand) GetOldLineNumber() string {
	return this.LineNumberStr
}

func (this *CncGcodeCommand) GetCodeString() string {
	codes := this.Codes
	ret := ""
	for i := 0; i < codes.Size(); i++ {
		ret += codes.Get(i).(*CncWord).String() + " "
	}
	return strings.TrimSpace(ret)
}

func (this *CncGcodeCommand) IsRepeatGcodeWord() bool {
	codes := this.Codes
	cncWordArr := make([]interface{}, codes.Size())
	for i := 0; i < codes.Size(); i++ {
		cncWordArr[i] = codes.Get(i)
	}
	collect := hash_map.NewHashMap()
	arrays.Reduce(cncWordArr, func(collection interface{}, item interface{}, index int) interface{} {
		cncWord := item.(*CncWord)

		if collect.ContainsKey(cncWord.Word) {
			collect.Put(cncWord.Word, collect.Get(cncWord.Word).(int)+1)
		} else {
			collect.Put(cncWord.Word, 1)
		}

		return collect
	}, collect)

	values := collect.Values()
	for i := 0; i < len(values); i++ {
		if values[i].(int) > 1 {
			return true
		}
	}

	return false
}

func (this *CncGcodeCommand) String() string {
	ret := ""
	if !this.IsComment {
		if strings.HasPrefix(this.LineNumberStr, "*") {
			ret += "N" + this.LineNumberStr + " "
		} else {
			ret += "N" + strconv.Itoa(this.LineNumber) + " "
		}
	}
	if this.Codes != nil {
		for i := 0; i < this.Codes.Size(); i++ {
			codeWord := this.Codes.Get(i).(*CncWord)
			ret += codeWord.String() + " "
		}
	}
	return ret
}

// Cnc Gcode 指令字
type CncWord struct {

	// 指令字
	Word string
	// 指令字 的值
	Val string

	// 是否是 指令字
	IsWord bool

	// 描述
	Comment string
}

func NewCncWord(str string) *CncWord {
	if len(str) > 1 {
		return &CncWord{
			Word:    string(str[0:1]),
			Val:     string(str[1:]),
			IsWord:  true,
			Comment: "",
		}
	} else {
		return &CncWord{
			Word:    "",
			Val:     "",
			IsWord:  true,
			Comment: "",
		}
	}

}

func (this *CncWord) GetValStr() string {
	return this.Val
}

func (this *CncWord) GetValOfNumber() float64 {
	val := this.Val
	f, _ := strconv.ParseFloat(val, 10)
	return f
}

func (this *CncWord) String() string {
	if this.IsWord {
		return this.Word + this.Val
	} else {
		return this.Comment
	}
}

func (this *CncGcodeFile) FindRepeatedGcodeWordCommand() *hash_map.HashMap {

	gcodeCommand := this.GetHead()

	hashMap := hash_map.NewHashMap()

	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {
		isRepeatGcodeWord := gcodeCommand.IsRepeatGcodeWord()
		if isRepeatGcodeWord {
			hashMap.Put(gcodeCommand.LineNumber, gcodeCommand)
		}

	}

	return hashMap
}

func (this *CncGcodeFile) FindExistGcodeWordCommand(word string) *hash_map.HashMap {
	gcodeCommand := this.GetHead()
	hashMap := hash_map.NewHashMap()
	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {

		exists := gcodeCommand.ExistsCncWord(word)
		if exists {
			hashMap.Put(gcodeCommand.LineNumber, gcodeCommand)
		}

	}
	return hashMap
}

func (this *CncGcodeFile) FormatToByte() []byte {
	retStr := ""
	gcodeCommand := this.GetHead()
	for gcodeCommand != nil {
		retStr += strings.TrimSpace(gcodeCommand.String()) + "\r\n"
		gcodeCommand = gcodeCommand.Next
	}
	return []byte(retStr)
}

func (this *CncGcodeFile) WriteFile(path string) error {
	gcodeBytes := this.FormatToByte()
	err := os.WriteFile(path, gcodeBytes, 0755)
	return err
}
