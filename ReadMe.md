
# SW_NC_Convert for grbl

本项目将SolidWorksCAM的dialog4后处理nc文件转换为支持grbl的刀路

- 支持中途换刀
    - 在换刀时，关闭主轴，暂停任务，换刀完毕，开启主轴，延迟5秒后开始继续工作
- 将不识别的G82/G83钻孔指令进行拆分

