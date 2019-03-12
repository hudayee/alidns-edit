# alidns-edit 
该脚本可自动获取当前ip，并对阿里云平台下域名进行解析。可设置为定时脚本，解决动态公网ip用户的痛处！

### 方法1
1.  拉取代码 ` git clone https://github.com/91hanbao/alidns-edit.git ` 
2.  编译生成可执行文件 ` go run build main.go `   
3.  编辑config.json文件
    ```
    {
        "AccessKeyId": "your accesskey",      //阿里云accesskeyid
        "Signature": "your signatrue",        //阿里云signatrue
        "DomainName": "example.com",          //需要设置解析的域名
        "RR": "www"                           //二级域名（主机记录）
      }
    ```
 4. 运行可执行文件（确保config.json和可执行文件在同一目录）
 
 ### 方法2
 1. 下载系统所对应架构的zip文件
 2. 解压后编辑config.json文件 参考方法 **1.3** 
 3. 运行可执行文件（确保config.json和可执行文件在同一目录）
 
### 设置定时任务
1. [window](https://blog.csdn.net/liu050604/article/details/82590504)
2. [linux](https://www.cnblogs.com/yjbjingcha/p/7006983.html)
