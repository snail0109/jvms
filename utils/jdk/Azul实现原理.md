# Azul JDK 实现原理

本文档详细分析 JVMS 中 Azul JDK 的实现原理，包括 API 调用机制、数据结构设计和下载流程。

## 1. 整体架构

Azul JDK 集成采用实时 API 调用的方式，通过 Azul 官方提供的 REST API 获取可用版本列表和下载链接，然后集成到 JVMS 的统一版本管理系统中。

### 核心文件结构
```
utils/jdk/azul.go          # Azul API 调用实现
internal/cmdCli/utils.go   # 版本列表集成逻辑
utils/web/web.go           # 通用下载功能
```

## 2. API 调用机制

### 2.1 API 端点构建

```go
func AzulApiEndpoint() string {
    var api = AzulApi() + "?os=$OS&arch=$ARCH&archive_type=zip&java_package_type=jdk&javafx_bundled=false&latest=true&release_status=ga&availability_types=CA&certifications=tck&page=1&page_size=100"
    api = strings.Replace(api, "$OS", runtime.GOOS, 1)
    api = strings.Replace(api, "$ARCH", runtime.GOARCH, 1)
    return api
}

func AzulApi() string {
    return "https://api.azul.com/metadata/v1/zulu/packages"
}
```

### 2.2 API 参数详解

| 参数 | 值 | 说明 |
|------|-----|------|
| `os` | `runtime.GOOS` | 自动检测操作系统（windows/linux/darwin） |
| `arch` | `runtime.GOARCH` | 自动检测架构（amd64/386/arm64） |
| `archive_type` | `zip` | 指定下载格式为 ZIP 压缩包 |
| `java_package_type` | `jdk` | 只获取 JDK 包，排除 JRE |
| `javafx_bundled` | `false` | 不包含 JavaFX 组件 |
| `latest` | `true` | 只获取每个主版本的最新版本 |
| `release_status` | `ga` | 只获取正式发布版本（General Availability） |
| `availability_types` | `CA` | 只获取商业可用版本 |
| `certifications` | `tck` | 获取 TCK（Technology Compatibility Kit）认证版本 |
| `page` | `1` | 分页参数：页码 |
| `page_size` | `100` | 分页参数：每页记录数 |

### 2.3 完整 API URL 示例

```
https://api.azul.com/metadata/v1/zulu/packages?os=windows&arch=amd64&archive_type=zip&java_package_type=jdk&javafx_bundled=false&latest=true&release_status=ga&availability_types=CA&certifications=tck&page=1&page_size=100
```

## 3. 数据结构设计

### 3.1 AzulJDK 结构体

```go
type AzulJDK struct {
    PackageUUID        string `json:"package_uuid"`        // 包的唯一标识符
    Name               string `json:"name"`                // 包的完整名称
    JavaVersion        []int  `json:"java_version"`        // Java 版本号数组 [主版本, 次版本, 补丁版本]
    OpenjdkBuildNumber int    `json:"openjdk_build_number"` // OpenJDK 构建号
    Latest             bool   `json:"latest"`              // 是否为该主版本的最新版本
    DownloadURL        string `json:"download_url"`        // 直接下载链接
    Product            string `json:"product"`             // 产品名称（通常为 "zulu"）
    DistroVersion      []int  `json:"distro_version"`      // Zulu 发行版版本号
    AvailabilityType   string `json:"availability_type"`   // 可用性类型
    ShortName          string                             // 处理后的短名称（用于 JVMS）
}
```

### 3.2 字段说明

- **PackageUUID**: 每个 JDK 包的唯一标识，用于精确引用
- **Name**: 完整包名，格式如 `zulu17.44.53-ca-jdk17.0.8-win_x64`
- **JavaVersion**: 数组格式的版本号，如 `[17, 0, 8]` 表示 Java 17.0.8
- **DownloadURL**: 可直接访问的下载链接，无需额外认证
- **ShortName**: 从 Name 提取的简化版本名，用于 JVMS 的版本管理

## 4. API 调用流程

### 4.1 完整调用链

```go
func AzulJDKs() []AzulJDK {
    // 步骤 1: 构建平台特定的 API URL
    url := AzulApiEndpoint()
    
    // 步骤 2: 发起 HTTP GET 请求
    body := call(url)
    
    // 步骤 3: 准备数据结构
    var jdks []AzulJDK
    
    // 步骤 4: 解析 JSON 响应
    err := json.Unmarshal(body, &jdks)
    if err != nil {
        fmt.Printf("error %v \n", err)
    }
    
    // 步骤 5: 后处理 - 提取短名称
    for i := 0; i < len(jdks); i++ {
        lastIndex := strings.LastIndex(jdks[i].Name, "-")
        jdks[i].ShortName = jdks[i].Name[0:lastIndex]
    }
    
    return jdks
}
```

### 4.2 HTTP 请求实现

```go
func call(url string) []byte {
    res, err := http.Get(url)          // 发起 GET 请求
    if err != nil {
        fmt.Printf("error: %v\n", err)
    }
    body, err := io.ReadAll(res.Body)  // 读取完整响应体
    if err != nil {
        fmt.Printf("error: %v\n", err)
    }
    return body                        // 返回原始 JSON 数据
}
```

## 5. 版本管理集成

### 5.1 集成到 JVMS 版本系统

```go
func getJdkVersions(cfx *entity.Config) ([]entity.JdkVersion, error) {
    // ... 其他 JDK 提供商的处理 ...
    
    // Azul JDKs 集成
    azulJdks := jdk.AzulJDKs()                    // 获取 Azul JDK 列表
    for _, azulJdk := range azulJdks {
        versions = append(versions, entity.JdkVersion{
            Version: azulJdk.ShortName,           // 使用处理后的短名称
            Url:     azulJdk.DownloadURL,         // 使用 API 返回的直接下载链接
        })
    }
    
    return versions, nil
}
```

### 5.2 统一的 JdkVersion 结构

```go
type JdkVersion struct {
    Version string `json:"version"`  // 版本标识符
    Url     string `json:"url"`      // 下载链接
}
```

## 6. 下载机制

### 6.1 下载流程

```go
func GetJDK(download string, v string, url string) (string, bool) {
    // 构建本地文件路径
    fileName := filepath.Join(download, fmt.Sprintf("%s.zip", v))
    
    // 清理已存在的文件
    os.Remove(fileName)
    
    if url == "" {
        fmt.Printf("JDK %s isn't available right now.", v)
        return "", false
    } else {
        fmt.Printf("Downloading jdk version %s...\n", v)
        
        // 调用通用下载函数
        if Download(url, fileName) {
            fmt.Println("Complete")
            return fileName, true
        } else {
            return "", false
        }
    }
    return "", false
}
```

### 6.2 带进度条的下载实现

```go
func Download(url string, target string) bool {
    response, err := client.Get(url)
    if err != nil || response.StatusCode != 200 {
        return false
    }
    defer response.Body.Close()

    output, err := os.Create(target)
    if err != nil {
        return false
    }
    defer output.Close()

    // 创建进度条
    bar := pb.New(int(response.ContentLength)).SetUnits(pb.U_BYTES_DEC)
    bar.ShowSpeed = true       // 显示下载速度
    bar.ShowTimeLeft = true    // 显示剩余时间
    bar.ShowFinalTime = true   // 显示完成时间
    bar.SetWidth(80)
    bar.Start()

    // 多路写入：同时写入文件和更新进度条
    writer := io.MultiWriter(output, bar)
    _, err = io.Copy(writer, response.Body)
    bar.Finish()

    return err == nil
}
```

## 7. API 响应示例

### 7.1 典型的 JSON 响应结构

```json
[
  {
    "package_uuid": "b8e5c8ca-bc76-4a43-9b95-ee1f9fc96f0d",
    "name": "zulu17.44.53-ca-jdk17.0.8-win_x64",
    "java_version": [17, 0, 8],
    "openjdk_build_number": 7,
    "latest": true,
    "download_url": "https://cdn.azul.com/zulu/bin/zulu17.44.53-ca-jdk17.0.8-win_x64.zip",
    "product": "zulu",
    "distro_version": [17, 44, 53],
    "availability_type": "CA"
  },
  {
    "package_uuid": "f2a5d6b1-3c8e-4f91-a2b4-d7e9f1c3a5b7",
    "name": "zulu11.66.15-ca-jdk11.0.20-win_x64",
    "java_version": [11, 0, 20],
    "openjdk_build_number": 8,
    "latest": true,
    "download_url": "https://cdn.azul.com/zulu/bin/zulu11.66.15-ca-jdk11.0.20-win_x64.zip",
    "product": "zulu",
    "distro_version": [11, 66, 15],
    "availability_type": "CA"
  }
]
```

### 7.2 名称处理逻辑

```go
// 原始名称: "zulu17.44.53-ca-jdk17.0.8-win_x64"
// 短名称提取:
lastIndex := strings.LastIndex(jdks[i].Name, "-")
jdks[i].ShortName = jdks[i].Name[0:lastIndex]
// 结果: "zulu17.44.53-ca-jdk17.0.8"
```

## 8. 设计优势

### 8.1 实时性
- 直接调用 Azul 官方 API，确保获取最新可用版本
- 无需维护静态版本列表，自动同步新发布的版本

### 8.2 平台适配
- 自动检测运行环境（操作系统和架构）
- API 返回匹配当前平台的 JDK 包

### 8.3 质量保证
- 只获取正式发布（GA）版本
- 确保版本通过 TCK 认证
- 过滤掉测试版和预览版

### 8.4 统一接口
- 与其他 JDK 提供商（Oracle、Adoptium）使用相同的数据结构
- 用户体验一致，无需了解底层实现差异

### 8.5 直接下载
- API 返回可直接访问的 CDN 链接
- 无需额外的重定向或认证步骤
- 支持断点续传和并发下载

## 9. 错误处理

### 9.1 网络错误处理
- HTTP 请求失败时输出错误信息
- JSON 解析失败时继续执行，不影响其他 JDK 提供商

### 9.2 版本兼容性
- 动态适配不同的 API 响应格式
- 容错处理确保部分失败不影响整体功能

## 10. 扩展性考虑

### 10.1 参数可配置化
- 可以通过配置文件自定义 API 参数
- 支持企业环境的特殊需求（如内网 API 代理）

### 10.2 缓存机制
- 可以添加本地缓存减少 API 调用频率
- 提高响应速度和离线可用性

### 10.3 版本过滤
- 可以添加自定义版本过滤逻辑
- 支持特定版本范围或标签的筛选