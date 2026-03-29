# SonarQube CLI

用于管理 SonarQube 临时资源的命令行工具。

## 功能

- 创建临时用户、组、项目和权限模板
- 清理临时资源
- 支持多插件并发测试

## 构建

```bash
make build
```

## 使用

### 创建资源

```bash
sonarqube-cli resources create \
  --config config.yaml \
  --task-run-id abc123 \
  --plugin tektoncd
```

### 清理资源

```bash
sonarqube-cli resources cleanup \
  --config config.yaml \
  --task-run-id abc123 \
  --plugin tektoncd
```

## 配置

参考 `config.example.yaml` 创建配置文件。

支持环境变量替换：
- `${TASK_RUN_ID}` - Tekton TaskRun ID
- `${SONARQUBE_MANAGER_TOKEN}` - 管理员 Token
