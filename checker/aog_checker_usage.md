# AogChecker 使用指南

## 目录

- Go
- Python
- C++
- C#
- Node.js

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

## C++

### 创建头文件 AogChecker.h

```cpp
#ifndef AOGCHECKER_H
#define AOGCHECKER_H

#ifdef __cplusplus
extern "C" {
#endif

void AOGInit();

#ifdef __cplusplus
}
#endif

#endif // AOGCHECKER_H
```

### 使用示例

创建一个新的 C++ 文件，例如 `main.cpp`：

```cpp
#include <iostream>
#include "AogChecker.h"

int main()
{
    // 调用 AOGInit 函数
    AOGInit();
    std::cout << "AOGInit called successfully" << std::endl;
    return 0;
}
```

编译和运行 C++ 程序：

```sh
g++ -o AogCheckerApp main.cpp -L. -lAogChecker
./AogCheckerApp
```

## C#

### 创建包装类

创建一个新的 C# 文件，例如 `AogChecker.cs`：

```csharp
using System;
using System.Runtime.InteropServices;

namespace AOGChecker
{
    public static class AogChecker
    {
        // 导入 AOGInit 函数
        [DllImport("AogChecker.dll", CallingConvention = CallingConvention.Cdecl)]
        public static extern void AOGInit();
    }
}
```

### 使用示例

创建一个新的 C# 控制台应用程序，并在其中引用并调用 Go DLL 中的 `AOGInit` 函数。

```sh
dotnet new console -n AogCheckerApp
cd AogCheckerApp
```

编辑 `Program.cs` 文件：

```csharp
using System;
using AOGChecker;

class Program
{
    static void Main(string[] args)
    {
        // 调用 AOGInit 函数
        AogChecker.AOGInit();
        Console.WriteLine("AOGInit called successfully");
    }
}
```

将生成的 `AogChecker.dll` 文件复制到 C# 项目的输出目录（例如 `bin/Debug/net6.0`）。

运行 C# 程序：

```sh
dotnet run
```

## Node.js

### 安装 `ffi-napi` 和 `ref-napi`

```sh
npm install ffi-napi ref-napi
```

### 使用示例

创建一个新的 JavaScript 文件，例如 `index.js`：

```javascript
const ffi = require('ffi-napi');
const ref = require('ref-napi');

// 定义返回类型和参数类型
const voidType = ref.types.void;

// 加载 DLL 文件
const AogChecker = ffi.Library('./AogChecker.dll', {
    'AOGInit': [voidType, []]
});

// 调用 AOGInit 函数
AogChecker.AOGInit();
console.log("AOGInit called successfully");
```

运行 Node.js 脚本：

```sh
node index.js
```
