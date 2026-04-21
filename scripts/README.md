# URL Scraper - inCompass URL Lookup Tool

## 功能说明

该脚本用于自动化获取目标网站的 `_token` 和 `g-recaptcha-response` 字段。

## 系统要求

- Python 3.8+
- Chrome 浏览器

## 安装步骤

### 1. 安装 Python

从 https://www.python.org/downloads/ 下载并安装 Python 3.8 或更高版本。
安装时请勾选 "Add Python to PATH"。

### 2. 安装依赖

```bash
pip install -r requirements.txt
```

或手动安装：

```bash
pip install selenium pandas openpyxl
```

### 3. 安装 Chrome 浏览器

确保已安装 Chrome 浏览器（脚本会自动下载对应版本的 ChromeDriver）。

## 使用方法

### 准备输入文件

创建一个 Excel 文件（.xlsx），第一列为要查询的网址：

```
要查询的网址
https://www.baidu.com/
https://www.google.com/
https://www.github.com/
```

### 运行脚本

```bash
python url_scraper.py sample_urls.xlsx output_results.xlsx
```

或者：

```bash
py url_scraper.py sample_urls.xlsx output_results.xlsx
```

### 参数说明

- 第一个参数：输入 Excel 文件路径
- 第二个参数（可选）：输出 Excel 文件路径，默认为 `output_results.xlsx`

## 输出格式

输出的 Excel 文件包含以下列：

| 要查询的网址 | _token字段 | g-recaptcha-response字段 | 状态 |
|-------------|-----------|-------------------------|------|
| https://www.baidu.com/ | xxxxxx | xxxxxx | success |

## 注意事项

1. **reCAPTCHA v3**: 该网站使用 Google reCAPTCHA v3，评分过低可能导致请求被拒绝。如遇问题，尝试将 `headless=True` 改为 `headless=False` 运行。

2. **请求频率**: 脚本在每个请求之间有 2 秒延迟，避免触发频率限制。

3. **Chrome 版本**: 确保 Chrome 浏览器版本较新，脚本会自动管理 ChromeDriver。

## 如果遇到问题

### Chrome 版本不匹配

```bash
pip install webdriver-manager
```

然后修改脚本中的驱动初始化部分。

### reCAPTCHA 验证失败

尝试以非无头模式运行，便于观察：

```python
scraper = URLScraper(headless=False)  # 改为 False
```

### 找不到 Python

确保 Python 已添加到系统 PATH，或使用完整路径：

```bash
C:\Python39\python.exe url_scraper.py sample_urls.xlsx
```