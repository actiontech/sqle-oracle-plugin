## SQLE Oracle Plugin
This is A Oracle SQL audit for  [SQLE](https://github.com/actiontech/sqle), which is an SQL audit platform.

### 1. Add Rule
You can define a rule in coresponding file, and then add it to the plugin by `AddRule()` method. More details please refer to [docs](https://actiontech.github.io/sqle-docs-cn/3.modules/3.7_auditplugin/auditplugin_development.html).

### 2. Build Binary
```bash
make docker_install # see binary file in bin/
```

The binary file is located in `bin/` folder, and you can use it to audit your database.

### 3. Begin to use
More details please refer to [docs](https://actiontech.github.io/sqle-docs-cn/3.modules/3.7_auditplugin/auditplugin_management.html).