
# SW_NC_Convert for grbl

本项目将SolidWorksCAM的m3axis后处理nc文件转换为支持grbl的刀路

### 支持中途手动换刀
  在换刀时，关闭主轴，暂停任务，换刀完毕，开启主轴，延迟5秒后继续开始工作

### 将不识别的G82钻孔指令进行处理
  grbl G82命令 P的时间为秒，将SolidWorks CAM生成的毫秒转换为秒


## 编译
在当前目录下，执行
```
CGO_ENABLE=0 GOOS=windows GOARCH=amd64 \
go build -o sw_nc_converter.exe \
main.go
```
