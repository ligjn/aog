package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// 导出结构体必须使用C兼容类型
type ComponentStatus struct {
	MissingServices     []map[string]interface{} `json:"missingServices"`
	MissingModels       []map[string]interface{} `json:"missingModels"`
	UnhealthyComponents []map[string]interface{} `json:"unhealthyComponents"`
}

var (
	userResponse   chan bool
	webServerMutex sync.Mutex
	webServerPort  = 5000
	lastStatus     *ComponentStatus
)

//export AOGInit
func AOGInit() {
	if !isAOGAvailable() {
		userResponse = make(chan bool)
		startWebServer()

		if !openBrowser(fmt.Sprintf("http://localhost:%d/install-prompt", webServerPort)) {
			return
		}

		select {
		case choice := <-userResponse:
			if !choice {
				return
			}
			if !downloadAOG() {
				return
			}
			if !installAOG() {
				return
			}
		case <-time.After(5 * time.Minute):
			return
		}
	}

	// AOG 可用，继续检查组件状态并安装缺失的组件
	showStatusDisplay()
}

func getComponentStatus() *C.char {
	data, _ := json.Marshal(lastStatus)
	return C.CString(string(data))
}

func main() {} // 必须保留空的main函数

// ========== 核心逻辑实现 ==========

func isAOGAvailable() bool {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:16688")
	return err == nil && resp.StatusCode == http.StatusOK
}

func startWebServer() {
	webServerMutex.Lock()
	defer webServerMutex.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/install-prompt", handleInstallPrompt)
	mux.HandleFunc("/user-response", handleUserResponse)
	mux.HandleFunc("/status-display", handleStatusDisplay)
	mux.HandleFunc("/component-status", handleComponentStatus)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", webServerPort))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Fatal(http.Serve(listener, mux))
	}()
}

func handleInstallPrompt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, installPromptHTML)
}

func handleUserResponse(w http.ResponseWriter, r *http.Request) {
	choice := r.URL.Query().Get("choice") == "true"
	userResponse <- choice
	w.WriteHeader(http.StatusOK)
}

func handleStatusDisplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, statusDisplayHTML)
}

func handleComponentStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data, _ := json.Marshal(lastStatus)
	w.Write(data)
}

// ========== 系统交互部分 ==========

func openBrowser(url string) bool {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start() == nil
}

func downloadAOG() bool {
	url := "http://120.232.136.73:31619/aogdev/aog.exe"
	userDir, _ := os.UserHomeDir()
	dest := filepath.Join(userDir, "AOG", "aog.exe")

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return false
	}

	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err == nil
}

func installAOG() bool {
	userDir, _ := os.UserHomeDir()
	aogPath := filepath.Join(userDir, "AOG", "aog.exe")

	cmd := exec.Command(aogPath, "server", "daemon")
	return cmd.Start() == nil
}

func showStatusDisplay() {
	// 实现状态显示逻辑
	// 读取 "relys.aog" 文件
	configPath := filepath.Join(getProjectRootDirectory(), ".aog")
	if !fileExists(configPath) {
		log.Fatalf("配置文件未找到: %s", configPath)
	}

	config := readConfig(configPath)
	lastStatus = checkComponent(config["service"].(map[string]interface{}))

	// 启动Web服务器并显示状态
	startWebServer()
	openBrowser(fmt.Sprintf("http://localhost:%d/status-display", webServerPort))
}

// ========== 工具方法 ==========

func getProjectRootDirectory() string {
	return filepath.Dir(os.Args[0])
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func readConfig(path string) map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("读取配置文件失败: %s", err)
	}
	defer file.Close()

	var config map[string]interface{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("解析配置文件失败: %s", err)
	}
	return config
}

func checkComponent(requiredServices map[string]interface{}) *ComponentStatus {
	status := getComponentResponse()
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal([]byte(status), &jsonResponse); err != nil {
		log.Fatalf("解析组件状态失败: %s", err)
	}

	result := &ComponentStatus{
		MissingServices:     []map[string]interface{}{},
		MissingModels:       []map[string]interface{}{},
		UnhealthyComponents: []map[string]interface{}{},
	}

	data := jsonResponse["data"].([]interface{})
	for serviceName, requiredService := range requiredServices {
		requiredModels := requiredService.(map[string]interface{})["models"].([]interface{})

		var actualService map[string]interface{}
		for _, s := range data {
			if s.(map[string]interface{})["service_name"].(string) == serviceName {
				actualService = s.(map[string]interface{})
				break
			}
		}

		if actualService == nil {
			// 缺少整个服务
			result.MissingServices = append(result.MissingServices, map[string]interface{}{
				"name":   serviceName,
				"status": "missing",
			})

			// 缺少服务的模型
			for _, requiredModel := range requiredModels {
				result.MissingModels = append(result.MissingModels, map[string]interface{}{
					"service": serviceName,
					"model":   requiredModel.(string),
					"status":  "missing",
				})
			}
		} else {
			actualModels := actualService["models"].([]interface{})
			actualModelSet := make(map[string]struct{})
			for _, m := range actualModels {
				actualModelSet[m.(string)] = struct{}{}
			}

			for _, requiredModel := range requiredModels {
				if _, found := actualModelSet[requiredModel.(string)]; !found {
					// 缺少模型
					result.MissingModels = append(result.MissingModels, map[string]interface{}{
						"service": serviceName,
						"model":   requiredModel.(string),
						"status":  "missing",
					})
				}
			}
		}
	}

	// 检查不健康的组件
	for _, part := range data {
		partName := part.(map[string]interface{})["service_name"].(string)
		partStatus := part.(map[string]interface{})["status"].(string)
		if partStatus != "1" {
			result.UnhealthyComponents = append(result.UnhealthyComponents, map[string]interface{}{
				"name":   partName,
				"status": partStatus,
			})
		}
	}

	return result
}

func getComponentResponse() string {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:16688/data")
	if err != nil {
		log.Fatalf("获取组件状态失败: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取组件状态响应失败: %s", err)
	}
	return string(body)
}

// ========== HTML模板 ==========

const installPromptHTML = `
<html><body style="padding:20px;font-family:Arial">
    <h2>安装确认</h2>
    <p>需要安装AOG组件才能继续，是否允许？</p>
    <button onclick="respond(true)">同意安装</button>
    <button onclick="respond(false)">取消</button>
    <script>
        function respond(choice) {
            fetch('/user-response?choice=' + choice)
                .then(() => window.close())
        }
    </script>
</body></html>`

const statusDisplayHTML = `
<html><body style="padding:20px;font-family:Arial">
    <h2>组件状态检查结果</h2>
    <div id="statusContainer">正在加载...</div>
    <button onclick="window.close()">关闭窗口</button>
    <script>
        fetch('/component-status')
            .then(r => r.json())
            .then(data => {
                let html = '<ul>';
                data.missingServices.forEach(function(s) {
                    html += '<li>缺少服务: ' + s.name + '</li>';
                });
                data.missingModels.forEach(function(m) {
                    html += '<li>缺少模型: ' + m.service + ' - ' + m.model + '</li>';
                });
                data.unhealthyComponents.forEach(function(u) {
                    html += '<li>不健康组件: ' + u.name + ' (状态: ' + u.status + ')</li>';
                });
                document.getElementById('statusContainer').innerHTML = html;
            });
    </script>
</body></html>`
