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
  --plugin tektoncd
```

输出示例：
```
SONARQUBE_TOKEN=squ_xxxxxxxxxxxxx
SONARQUBE_USER=test-tektoncd-abc123
```

#### 清理资源

```bash
sonarqube-cli resources cleanup \
  --config config.yaml \
  --task-run-id abc123 \
  --plugin tektoncd
```

### 细粒度命令

#### 删除项目

```bash
sonarqube-cli project delete \
  --endpoint https://sonarqube.example.com \
  --token <manager-token> \
  --key tektoncd-pipeline-abc123
```

#### 删除用户

```bash
sonarqube-cli user delete \
  --endpoint https://sonarqube.example.com \
  --token <manager-token> \
  --login test-tektoncd-abc123
```

#### 删除用户组

```bash
sonarqube-cli group delete \
  --endpoint https://sonarqube.example.com \
  --token <manager-token> \
  --name test-tektoncd-abc123
```

#### 撤销 Token

```bash
sonarqube-cli token revoke \
  --endpoint https://sonarqube.example.com \
  --token <manager-token> \
  --login test-tektoncd-abc123 \
  --name test-token-abc123
```

## 配置

参考 `config.example.yaml` 创建配置文件。

支持环境变量和模板变量替换：
- `{{TASK_RUN_ID}}` 或 `${TASK_RUN_ID}` - Tekton TaskRun ID
- `{{PLUGIN_NAME}}` 或 `${PLUGIN_NAME}` - 插件名称
- `${SONARQUBE_MANAGER_TOKEN}` - 管理员 Token（环境变量）

## 在 Tekton Pipeline 中使用

```bash
# 创建资源并获取 Token
OUTPUT=$(sonarqube-cli resources create \
  --config config.yaml \
  --task-run-id ${TEKTON_TASK_RUN_ID} \
  --plugin tektoncd)

# 提取 Token
TEMP_TOKEN=$(echo "$OUTPUT" | grep SONARQUBE_TOKEN | cut -d= -f2)

# 使用 Token 执行扫描
sonar-scanner \
  -Dsonar.host.url=https://sonarqube.example.com \
  -Dsonar.login=$TEMP_TOKEN \
  -Dsonar.projectKey=tektoncd-pipeline-${TEKTON_TASK_RUN_ID}
```

