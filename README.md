# vcjob-controller
参照kubernetes sample controller，使用code generator写一个监控volcano job的controller
目录
controller
├── LICENSE
├── README.md
├── deploy # 部署 Controller 的相关文件，如 Deployment、CRD、RBAC。
├── go.mod # Go mod package 
├── go.sum # Go mod package 
├── hack   # 存放一些常使用到的脚本
└── pkg    # code genrator生成的api server客户端库，控制器相关程序
