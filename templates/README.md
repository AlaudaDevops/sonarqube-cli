# SonarQube CLI 模板与配置

本目录包含 SonarQube CLI 的配置示例和资源模板。

## 配置文件 (config.yaml)

用于定义需要同步到 SonarQube 的资源，包括：
- `projects`: 项目列表
- `users`: 用户列表及其所属组
- `groups`: 用户组列表

### 示例配置

```yaml
projects:
  - key: my-project
    name: My Project
    visibility: public

users:
  - login: developer
    name: Developer User
    email: dev@example.com
    password: password123
    groups:
      - sonar-users

groups:
  - name: sonar-users
    description: Default group for sonar users
```

## 使用方法

执行同步命令时指定配置文件：

```bash
sonarqube-cli resources --endpoint http://sonar.example.com --token <your-token> --config templates/config.yaml
```
