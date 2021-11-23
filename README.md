## SQLE Oracle Plugin
This is A Oracle SQL audit for  [SQLE](https://github.com/actiontech/sqle), which is an SQL audit platform.


### 1. 添加规则
在文件中定义规则与规则的处理函数，最后通过插件库的 `AddRule()` 方法将规则注册到插件中。详细说明请参考文档 [数据库审核插件开发](https://actiontech.github.io/sqle-docs-cn/3.modules/3.7_auditplugin/auditplugin_development.html) 中**快速开始部分**。

### 2. 构建二进制
```bash
make docker_install # see binary file in bin/
```

插件的二进制文件最终生成在 `./bin` 目录下。

### 3. 使用
参考文档：[数据库审核插件使用](https://actiontech.github.io/sqle-docs-cn/3.modules/3.7_auditplugin/auditplugin_management.html)