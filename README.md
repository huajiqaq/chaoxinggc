# chaoxinggc
超星图书馆座位预约脚本 golong版本
借鉴 https://github.com/bear-zd/ChaoXingReserveSeat

## 注意

本项目可能仅对部分学校有效 因为本项目目标学校使用的是带seatId的版本 你可能需要根据 https://github.com/bear-zd/ChaoXingReserveSeat 修改代码来使用

该版本不支持验证码，因为本人并未遇到验证码，如果有验证，请自行修改代码

本项目仅支持本地部署 首先安装go 运行路径下的 start.bat 即可

该项目仅实现了单账号 但你可以指定-u 来运行多个程序实现多账户 例如 go run . -u config1.json

## 获取roomid（图书馆id）和seatid（座位号）

在使用之前需要先在如下获取图书馆对应的id和座位号，下面的配置里已经提供了上海大学图书馆的id。对于不知道id的，可以通过如下方式进行：

![image-20231012153826054](https://zideapicbed.oss-cn-shanghai.aliyuncs.com/img/image-20231012153826054.png)

在进入预约图书馆列表界面时断开网络，点击你想预约的图书馆的`选座`按钮，会提示网页无法打开，此时点击`右上角的三条杠`，选择`复制链接`，会得到类似这样的链接：

> https://office.chaoxing.com/front/apps/seat/select?id=5483&day=2023-10-12&backLevel=2&pageToken=0f46f3acc7be4c60862cb9815870ddfd

其中的`id=5483`的5483即为对应图书馆的id，将其填写到config.json中，座位联网后自己挑即可（详细填写参见后面的setting）

## config配置
之后编辑config.json并填写座位预约相关信息即可
```json
{
    "reserve": [
        {"username": "XXXXXXXX", //https://passport2.chaoxing.com/mlogin?loginType=1&newversion=true&fid=&  在这个网站查看是否可以顺利登陆 
        "password": "XXXXXXXX",
        "time": ["08:00","22:00"], // 预约的起始时间
        "roomid":"2609", //2609:四楼外圈,5483:四楼内圈,2610:五楼外圈,5484:五楼内圈
        "seatid":"002", // 注意要用0补全至3位数，例如6号座位应该填006
        "daysofweek": ["Monday" , "Tuesday", "Wednesday", "Thursday", "Friday"] // 指定抢的日期
        "starttime" : "07:00:00" // 开始抢座时间
        },
    ]
}
```