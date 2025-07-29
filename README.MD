# 📚 Go Book Downloader

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.20+-blue.svg" alt="Go version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
</p>

<p align="center">
  一个用 Go 语言编写的小说下载工具，支持多线程下载并将小说保存为文本文件
</p>

## 🌟 特性

- 🚀 **多线程下载** - 提高下载速度和效率
- ⚙️ **自定义配置** - 支持正则表达式自定义匹配规则
- 🧵 **并发控制** - 可调节的线程数控制
- 📝 **多种输出格式** - 保存为纯文本文件
- 📊 **详细日志** - 多级别日志记录，便于调试
- 📂 **本地文件支持** - 可从本地文件读取网页内容

## 📦 从源代码编译

确保你的系统已安装 Go 1.21 或更高版本。 \
删除多余的mod文件

```bash
# 克隆项目
git clone <repository-url>

# 进入项目目录
cd go_bookdownloader

# 编译项目
go build
```

## 🚀 快速开始

### 基本用法

```bash
./godown.exe -u "https://example.com/book/123.html"
```

### 命令行参数详解

| 参数 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `-u` | string | 小说主页URL | 必填 |
| `-t` | int | 下载线程数 | 20 |
| `-c` | string | 配置文件路径 | 空 |
| `-o` | string | 输出文件名 | 小说名称 |
| `-l` | string | 日志等级 (debug/info/warn/error) | info |
| `-f` | string | 本地主页文件路径 | 空 |

### 配置文件格式

配置文件采用 JSON 格式，包含以下字段：

```json
{
    "name_regex": "匹配小说名称的正则表达式",
    "chapter_regex": "匹配章节链接的正则表达式",
    "title_regex": "匹配章节标题的正则表达式",
    "content_regex": "匹配章节内容的正则表达式"
}
```

#### 示例配置文件

```json
{
    "name_regex": "<h1\\s+class=\"book-name\">.*?<a\\s+[^>]*title=\"([^\"]+)\"",
    "chapter_regex": "<div class=\"col-md-6 item\">\\s*<a\\s+href=\"([^\"]+)\"[^>]*>[^<]+</a>\\s*</div>",
    "title_regex": "<h1\\s+class=\"cont-title\">([^<]+)</h1>",
    "content_regex": "<p>(.*?)</p>"
}
```

## 💡 使用示例

### 1. 基本下载
```bash
./godown.exe -u "https://example.com/book/123.html"
```

### 2. 指定线程数下载
```bash
./godown.exe -u "https://example.com/book/123.html" -t 30
```

### 3. 使用配置文件下载
```bash
./godown.exe -u "https://example.com/book/123.html" -c "./config.json"
```

### 4. 指定输出文件名
```bash
./godown.exe -u "https://example.com/book/123.html" -o "我的小说"
```

### 5. 调试模式下载
```bash
./godown.exe -u "https://example.com/book/123.html" -l debug
```

### 6. 从本地文件读取主页内容
```bash
./godown.exe -u "https://example.com/book/123.html" -f "./index.html"
```

## 📁 输出格式

程序会将下载的小说保存为 `.txt` 文本文件，格式如下：

```
第一章 章节标题
章节内容...

第二章 章节标题
章节内容...
```

## ⚠️ 注意事项

- 请遵守网站的使用条款和版权规定
- 不要在短时间内发送过多请求，以免给网站服务器造成压力
- 如果遇到下载失败，可以尝试降低线程数重试
- 请合理使用本工具，尊重作者和网站的劳动成果

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。