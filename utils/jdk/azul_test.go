package jdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestAzulJDKs 测试 Azul JDK 列表获取功能
func TestAzulJDKs(t *testing.T) {
	fmt.Println("=== Azul JDK 列表测试 ===")
	fmt.Printf("运行环境: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// 测试 API 端点构建
	endpoint := AzulApiEndpoint()
	fmt.Printf("API 端点: %s\n", endpoint)
	fmt.Println()

	// 获取 JDK 列表
	fmt.Println("正在获取 Azul JDK 列表...")
	start := time.Now()
	jdks := AzulJDKs()
	duration := time.Since(start)

	fmt.Printf("请求耗时: %v\n", duration)
	fmt.Printf("获取到 %d 个 JDK 版本\n", len(jdks))
	fmt.Println()

	if len(jdks) == 0 {
		t.Error("未获取到任何 JDK 版本")
		return
	}

	// 验证数据结构
	for i, jdk := range jdks {
		if jdk.DownloadURL == "" {
			t.Errorf("JDK %d 的下载链接为空", i)
		}
		if jdk.Name == "" {
			t.Errorf("JDK %d 的名称为空", i)
		}
		if len(jdk.JavaVersion) == 0 {
			t.Errorf("JDK %d 的版本信息为空", i)
		}
	}

	// 显示详细信息
	printJDKDetails(jdks)

	// 验证下载链接
	fmt.Println("=== 下载链接验证 ===")
	validateDownloadLinks(jdks[:min(3, len(jdks))]) // 只验证前3个链接
}

// TestAzulApiEndpoint 测试 API 端点构建
func TestAzulApiEndpoint(t *testing.T) {
	endpoint := AzulApiEndpoint()

	// 验证基础 URL
	if !strings.Contains(endpoint, "https://api.azul.com/metadata/v1/zulu/packages") {
		t.Error("API 端点不包含正确的基础 URL")
	}

	// 验证必要参数
	requiredParams := []string{
		"os=" + runtime.GOOS,
		"arch=" + runtime.GOARCH,
		"archive_type=zip",
		"java_package_type=jdk",
		"latest=true",
		"release_status=ga",
	}

	for _, param := range requiredParams {
		if !strings.Contains(endpoint, param) {
			t.Errorf("API 端点缺少必要参数: %s", param)
		}
	}
}

// printJDKDetails 打印 JDK 详细信息
func printJDKDetails(jdks []AzulJDK) {
	fmt.Println("=== JDK 详细信息 ===")

	// 按 Java 主版本分组
	versionMap := make(map[int][]AzulJDK)
	for _, jdk := range jdks {
		if len(jdk.JavaVersion) > 0 {
			majorVersion := jdk.JavaVersion[0]
			versionMap[majorVersion] = append(versionMap[majorVersion], jdk)
		}
	}

	// 打印分组信息
	for majorVersion, jdkList := range versionMap {
		fmt.Printf("\n--- Java %d ---\n", majorVersion)
		for _, jdk := range jdkList {
			// fmt.Printf("  版本: %s\n", jdk.ShortName)
			// fmt.Printf("  完整名称: %s\n", jdk.Name)
			// fmt.Printf("  Java版本: %v\n", jdk.JavaVersion)
			// fmt.Printf("  发行版版本: %v\n", jdk.DistroVersion)
			// fmt.Printf("  构建号: %d\n", jdk.OpenjdkBuildNumber)
			fmt.Printf("  下载链接: %s\n", jdk.DownloadURL)
			// fmt.Printf("  是否最新: %t\n", jdk.Latest)
			// fmt.Printf("  可用性类型: %s\n", jdk.AvailabilityType)
			// fmt.Println("  ---")
		}
	}
}

// validateDownloadLinks 验证下载链接的有效性
func validateDownloadLinks(jdks []AzulJDK) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for i, jdk := range jdks {
		fmt.Printf("验证链接 %d: %s\n", i+1, jdk.ShortName)

		// 发送 HEAD 请求检查链接有效性
		resp, err := client.Head(jdk.DownloadURL)
		if err != nil {
			fmt.Printf("  ❌ 错误: %v\n", err)
			continue
		}
		resp.Body.Close()

		// 检查状态码
		if resp.StatusCode == http.StatusOK {
			contentLength := resp.Header.Get("Content-Length")
			contentType := resp.Header.Get("Content-Type")
			fmt.Printf("  ✅ 状态: %d\n", resp.StatusCode)
			fmt.Printf("  📦 文件大小: %s bytes\n", contentLength)
			fmt.Printf("  📄 内容类型: %s\n", contentType)
		} else {
			fmt.Printf("  ❌ 状态码: %d\n", resp.StatusCode)
		}
		fmt.Println()
	}
}

// BenchmarkAzulJDKs 性能测试
func BenchmarkAzulJDKs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		AzulJDKs()
	}
}

// TestAzulJDKStructure 测试数据结构解析
func TestAzulJDKStructure(t *testing.T) {
	// 模拟 API 响应数据
	mockResponse := `[
		{
			"package_uuid": "test-uuid-123",
			"name": "zulu17.44.53-ca-jdk17.0.8-win_x64",
			"java_version": [17, 0, 8],
			"openjdk_build_number": 7,
			"latest": true,
			"download_url": "https://cdn.azul.com/zulu/bin/test.zip",
			"product": "zulu",
			"distro_version": [17, 44, 53],
			"availability_type": "CA"
		}
	]`

	var jdks []AzulJDK
	err := json.Unmarshal([]byte(mockResponse), &jdks)
	if err != nil {
		t.Fatalf("JSON 解析失败: %v", err)
	}

	if len(jdks) != 1 {
		t.Fatalf("期望解析 1 个 JDK，实际得到 %d 个", len(jdks))
	}

	jdk := jdks[0]

	// 验证字段
	if jdk.PackageUUID != "test-uuid-123" {
		t.Errorf("PackageUUID 解析错误: %s", jdk.PackageUUID)
	}

	if jdk.Name != "zulu17.44.53-ca-jdk17.0.8-win_x64" {
		t.Errorf("Name 解析错误: %s", jdk.Name)
	}

	if len(jdk.JavaVersion) != 3 || jdk.JavaVersion[0] != 17 {
		t.Errorf("JavaVersion 解析错误: %v", jdk.JavaVersion)
	}

	if !jdk.Latest {
		t.Error("Latest 字段解析错误")
	}

	// 测试短名称提取
	lastIndex := strings.LastIndex(jdk.Name, "-")
	expectedShortName := jdk.Name[0:lastIndex]
	jdk.ShortName = expectedShortName

	if jdk.ShortName != "zulu17.44.53-ca-jdk17.0.8" {
		t.Errorf("短名称提取错误: %s", jdk.ShortName)
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
