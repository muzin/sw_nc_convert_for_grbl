package main

import (
	"fmt"
	"os"
	"strings"
	"sw_nc_convert_for_grbl/cnc_gcode"
	"sw_nc_convert_for_grbl/convert"
	"time"
)

func main() {

	//dialog4()

	//m3axis()

	input()

}

func input() {

	var millType int = 0
	var ncFilePath = ""

	fmt.Println("=========================================")
	fmt.Println("\t\t\tSolidWorks CNC 转换 GRBL 工具")
	fmt.Println("=========================================")
	fmt.Println("铣床类型:")
	fmt.Println(" 1: M3AXIS")
	fmt.Println(" 2: DIALOG4")
	fmt.Println("=========================================")
	fmt.Println("请输入铣床类型：")

	fmt.Scan(&millType)

	if !(millType > 0 && millType < 3) {
		fmt.Printf("请输入上文中的铣床类型!!!")
		os.Exit(255)
	}

	fmt.Println("=========================================")
	fmt.Println("请输入NC文件路径：")

	fmt.Scan(&ncFilePath)

	fmt.Println("")
	fmt.Println("正在解析...")
	time.Sleep(1 * time.Second)
	// 解析
	cncGcodeFile := cnc_gcode.Parse(ncFilePath)

	converterType := ""
	if millType == 1 {
		converterType = "sw/mill/m3axis"
	} else if millType == 2 {
		converterType = "sw/mill/dialog"
	}

	fmt.Println("正在加载转换器...")
	time.Sleep(500 * time.Millisecond)

	// 转换
	cncConverter := convert.GetCncConverter(converterType)

	fmt.Println("转换中...")
	time.Sleep(500 * time.Millisecond)

	newCncGcodeFile := cncConverter.Convert(cncGcodeFile)

	cncGcodeCommand := newCncGcodeFile.GetHead()
	cncGcodeCommand.AddToBack(newCncGcodeFile.ParseToCncGcodeCommand("(Email: sirius1@aliyun.com)"))
	cncGcodeCommand.AddToBack(newCncGcodeFile.ParseToCncGcodeCommand("(Auther: Sirius)"))
	cncGcodeCommand.AddToBack(newCncGcodeFile.ParseToCncGcodeCommand("(This Grbl Cnc Gcode Powered By SolidWorks CNC Convert GRBL Tool)"))

	ncIndex := strings.LastIndex(ncFilePath, ".nc")

	fmt.Println("写入到本地...")
	time.Sleep(500 * time.Millisecond)

	// 写到 本地
	newCncGcodeFile.WriteFile(ncFilePath[0:ncIndex] + "_" + strings.ReplaceAll(converterType, "/", "_") + "_for_grbl.nc")

	//fmt.Printf("%v\n", toolChangingIndexs)

	fmt.Printf("转换完成！！！")

}

func dialog4() {
	ncFilePath := "/Users/sirius/GolandProjects/sw_nc_convert_for_grbl/tmp/test_src8.nc"

	// 解析
	cncGcodeFile := cnc_gcode.Parse(ncFilePath)

	// 转换
	cncConverter := convert.GetCncConverter("sw/mill/dialog4")
	newCncGcodeFile := cncConverter.Convert(cncGcodeFile)

	ncIndex := strings.LastIndex(ncFilePath, ".nc")
	// 写到 本地
	newCncGcodeFile.WriteFile(ncFilePath[0:ncIndex] + "_sw_dialog4_for_grbl.nc")

	//fmt.Printf("%v\n", toolChangingIndexs)

	fmt.Printf("finish.")
}

func m3axis() {

	ncFilePath := "/Users/sirius/GolandProjects/sw_nc_convert_for_grbl/tmp/test_src9_sw_m3axis.nc"

	// 解析
	cncGcodeFile := cnc_gcode.Parse(ncFilePath)

	// 转换
	cncConverter := convert.GetCncConverter("sw/mill/m3axis")
	newCncGcodeFile := cncConverter.Convert(cncGcodeFile)

	cncGcodeCommand := newCncGcodeFile.GetHead()
	cncGcodeCommand.AddToBack(newCncGcodeFile.ParseToCncGcodeCommand("(This Grbl Cnc Gcode Powered By SolidWorks CNC Convert GRBL Tool)"))

	ncIndex := strings.LastIndex(ncFilePath, ".nc")
	// 写到 本地
	newCncGcodeFile.WriteFile(ncFilePath[0:ncIndex] + "_sw_m3axis_for_grbl.nc")

	//fmt.Printf("%v\n", toolChangingIndexs)

	fmt.Printf("finish.")
}
