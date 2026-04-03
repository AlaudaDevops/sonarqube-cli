# SonarQube CLI

用于管理 SonarQube 临时资源的命令行工具。

## 功能

- 创建临时用户、组、项目和权限模板
- 自动生成临时用户 Token
- 清理临时资源
- 支持多插件并发测试
- 支持变量替换（{{TASK_RUN_ID}}、${TASK_RUN_ID}）
- 细粒度资源管理命令

## 构建

```bash
make build
```

## 使用

### 批量资源管理

#### 创建资源

```bash
sonarqube-cli resources create \
  --config config.yaml \
  --task-run-id abc123 \
  --plugin tektoncd \
  --manager-token-file /secure/manager.token \
  --temp-user-password-file /secure/temp-user.password \
  --token-file /tmp/sonarqube.env \
  --state-file /tmp/sonarqube.state.yaml
```

输出示例：
```
Token written to /tmp/sonarqube.env
```

#### 清理资源

```bash
sonarqube-cli resources cleanup \
  --config config.yaml \
  --plugin tektoncd \
  --manager-token-file /secure/manager.token \
  --state-file /tmp/sonarqube.state.yaml
```

`--state-file` 必须来自同一 SonarQube 实例上的 `resources create`，并会在 cleanup 成功后自动删除。
`--manager-token-file` 和 `--temp-user-password-file` 优先于环境变量，适合在 CI 中避免通过进程环境传递敏感信息。

### 细粒度命令

#### 删除项目

```bash
sonarqube-cli project delete \
  --endpoint https://sonarqube.example.com \
  --manager-token-file /secure/manager.token \
  --key tektoncd-pipeline-abc123
```

#### 删除用户

```bash
sonarqube-cli user delete \
  --endpoint https://sonarqube.example.com \
  --manager-token-file /secure/manager.token \
  --login test-tektoncd-abc123
```

#### 删除用户组

```bash
sonarqube-cli group delete \
  --endpoint https://sonarqube.example.com \
  --manager-token-file /secure/manager.token \
  --name test-tektoncd-abc123
```

#### 撤销 Token

```bash
sonarqube-cli token revoke \
  --endpoint https://sonarqube.example.com \
  --manager-token-file /secure/manager.token \
  --login test-tektoncd-abc123 \
  --name test-token-abc123
```

## 配置

参考 `config.example.yaml` 创建配置文件。

支持环境变量和模板变量替换：
- `{{TASK_RUN_ID}}` 或 `${TASK_RUN_ID}` - Tekton TaskRun ID
- `{{PLUGIN_NAME}}` 或 `${PLUGIN_NAME}` - 插件名称
- `${SONARQUBE_MANAGER_TOKEN}` - 管理员 Token（环境变量，或通过 `--manager-token-file` 覆盖）
- `${TEMP_USER_PASSWORD}` - 临时测试用户密码（环境变量，或通过 `--temp-user-password-file` 覆盖）

## 在 Tekton Pipeline 中使用

```bash
# 创建资源并获取 Token
sonarqube-cli resources create \
  --config config.yaml \
  --task-run-id ${TEKTON_TASK_RUN_ID} \
  --plugin tektoncd \
  --manager-token-file /workspace/secrets/manager.token \
  --temp-user-password-file /workspace/secrets/temp-user.password \
  --token-file /workspace/sonarqube.env \
  --state-file /workspace/sonarqube.state.yaml

# 读取 Token
. /workspace/sonarqube.env

# 使用 Token 执行扫描
sonar-scanner \
  -Dsonar.host.url=https://sonarqube.example.com \
  -Dsonar.login=$SONARQUBE_TOKEN \
  -Dsonar.projectKey=tektoncd-pipeline-${TEKTON_TASK_RUN_ID}

# 扫描完成后清理；使用 create 生成的状态文件，避免按模板配置盲删现有资源
sonarqube-cli resources cleanup \
  --config config.yaml \
  --manager-token-file /workspace/secrets/manager.token \
  --state-file /workspace/sonarqube.state.yaml
```
