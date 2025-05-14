# AogChecker 使用指南

## 目录

- Go
- Python
- C/C++
- C#
- Node.js

## 说明
AOGInit("path/to/.aog") 方法可以选择输入一个参数，即.aog文件的路径；若不输入则默认为项目根目录。

## Go

### 编译 Go 程序为 DLL

在 Go 程序目录中运行以下命令：

```sh
go build -o AogChecker.dll -buildmode=c-shared AogChecker.go
```

### 使用示例

在 Go 中，您可以直接调用 Go 函数，而不需要通过 DLL。

## Python

### 安装 `ctypes` 库

`ctypes` 是 Python 标准库的一部分，无需额外安装。

### 使用示例

创建一个新的 Python 文件，例如 `testChecker.py`：

```python
from ctypes import CDLL

# 加载 DLL 文件
lib = CDLL('./AogChecker.dll')

# 设置函数原型
lib.AOGInit.argtypes = []
lib.AOGInit.restype = None

# 调用 AOGInit 函数
lib.AOGInit()
print("AOGInit called successfully")
```

运行 Python 脚本：

```sh
python testChecker.py
```

## C/C++

### 创建静态库
```sh
cd go_lib
go build -o aogchecker.a -buildmode=c-archive AogChecker.go
```

### 使用示例

创建一个新的 C/C++ 文件，例如 `main.cpp`：

```c
#include "AogChecker.h"

int main() { 
    AOGInit(""); // 这里需要输入参数，参数可以为空
    return 0;
}

```

编译和运行 C 程序：

```sh
gcc main.c aogchecker.a -o main
```

## C#

### 创建 C# 类库
```sh
cd dotnet_lib
dotnet build --configuration Release
```

### 将 dotnetCheckerLib.dll 复制到他们的项目目录
在 /bin/Release/ 文件夹中找到 aog-cheker.dll 并将其复制到项目目录中。

### 在项目中添加引用

在 .csproj 文件中手动添加：

```xml
<ItemGroup>
    <Reference Include="aog-checker">
        <HintPath>path/to/aog-checker.dll</HintPath>
    </Reference>
</ItemGroup>
```

或者使用 Visual Studio 手动添加：

右键 项目 > 添加引用 > 浏览 > 选择 aog-checker.dll

### 在代码中使用

```csharp
using aog-checker;

class Program
{
    static void Main()
    {
        var checker = new AogChecker();
        checker.AOGInit();
    }
}
```

运行 C# 程序：

```sh
dotnet run
```

## Node.js

### 安装checker

```sh
npm install path/to/aog-checker-1.0.1.tgz
```

### 使用示例

创建一个新的 JavaScript 文件，例如 `index.js`：

```javascript
const checker = require('aog-checker');

checker.AOGInit();
console.log('AOGInit called successfully');
```
​
