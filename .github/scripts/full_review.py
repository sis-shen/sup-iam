import os
import glob
import requests
import json
import sys
from pathlib import Path

# 获取所有代码文件（排除二进制文件和特定目录）
all_files = []
exclude_dirs = ["vendor", ".git", "test", "tests", "mock", "mocks", "node_modules", "dist", "build"]
exclude_extensions = [".exe", ".dll", ".so", ".dylib", ".pyc", ".pyo", ".class", ".o", ".a"]

# 常见代码文件扩展名
code_extensions = [
    ".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".java", ".c", ".cpp", ".h", ".hpp",
    ".rs", ".rb", ".php", ".cs", ".swift", ".kt", ".scala", ".sh", ".bash", ".zsh",
    ".yaml", ".yml", ".json", ".xml", ".html", ".css", ".scss", ".sql", ".proto",
    ".md", ".txt", ".conf", ".cfg", ".ini", ".toml", ".properties"
]

for file in glob.glob("**/*", recursive=True):
    # 跳过目录
    if os.path.isdir(file):
        continue
    
    # 跳过隐藏文件
    if os.path.basename(file).startswith("."):
        continue
    
    # 跳过排除的目录
    if any(exclude_dir in Path(file).parts for exclude_dir in exclude_dirs):
        continue
    
    # 检查文件扩展名
    ext = os.path.splitext(file)[1].lower()
    if ext in exclude_extensions:
        continue
    
    # 只包含代码文件
    if ext in code_extensions:
        all_files.append(file)

if not all_files:
    print("No code files found")
    sys.exit(0)

print(f"Found {len(all_files)} files to review")

# 分批处理
batch_size = 3  # 每批3个文件
all_reviews = []
api_key = os.environ.get("DEEPSEEK_API_KEY")

if not api_key:
    print("DEEPSEEK_API_KEY not set")
    sys.exit(1)

for i in range(0, len(all_files), batch_size):
    batch = all_files[i:i+batch_size]
    
    # 构建本次审查内容
    review_content = f"## Batch {i//batch_size + 1}/{((len(all_files)-1)//batch_size) + 1}\n\n"
    review_content += f"Review these {len(batch)} files:\n\n"
    
    for file in batch:
        try:
            with open(file, "r", encoding="utf-8", errors="ignore") as f:
                content = f.read()
                # 限制每个文件大小
                if len(content) > 5000:
                    content = content[:5000] + "\n\n... (file truncated)"
                
                # 根据文件扩展名选择代码块语言
                ext = os.path.splitext(file)[1].lower()
                lang = ext[1:] if ext else "text"
                if lang in ["yml", "yaml"]:
                    lang = "yaml"
                elif lang in ["js", "ts"]:
                    lang = "javascript"
                elif lang in ["jsx", "tsx"]:
                    lang = "typescript"
                
                review_content += f"### File: `{file}`\n"
                review_content += f"```{lang}\n{content}\n```\n\n"
        except Exception as e:
            review_content += f"### File: `{file}`\nError reading: {str(e)}\n\n"
    
    # 调用DeepSeek API审查每批文件
    try:
        response = requests.post(
            "https://api.deepseek.com/v1/chat/completions",
            headers={
                "Authorization": f"Bearer {api_key}",
                "Content-Type": "application/json"
            },
            json={
                "model": "deepseek-chat",
                "messages": [
                    {
                        "role": "system",
                        "content": """You are an expert code reviewer. Perform a comprehensive review of the provided code files. Focus on:
                        1. Code quality and readability
                        2. Potential bugs and edge cases
                        3. Performance issues
                        4. Best practices for each language
                        5. Security concerns
                        6. Error handling
                        7. Code structure and organization
                        8. Documentation quality
                        
                        Provide specific, actionable feedback. Prioritize critical issues."""
                    },
                    {"role": "user", "content": review_content}
                ],
                "temperature": 0.3,
                "max_tokens": 4000
            },
            timeout=90
        )
        
        if response.status_code == 200:
            result = response.json()
            review = result["choices"][0]["message"]["content"]
            all_reviews.append(f"## Batch {i//batch_size + 1}/{((len(all_files)-1)//batch_size) + 1}\n\n{review}\n\n")
            print(f"Completed batch {i//batch_size + 1}")
        else:
            all_reviews.append(f"## Batch {i//batch_size + 1}\n\nAPI Error: {response.status_code}\n{response.text}\n\n")
            
    except Exception as e:
        all_reviews.append(f"## Batch {i//batch_size + 1}\n\nError: {str(e)}\n\n")

# 统计文件类型
file_types = {}
for file in all_files:
    ext = os.path.splitext(file)[1].lower()
    file_types[ext] = file_types.get(ext, 0) + 1

type_summary = "\n".join([f"- {ext or 'no extension'}: {count} files" for ext, count in sorted(file_types.items())])

# 生成最终报告
report = f"""# 📊 Full Code Review Report

## 📈 Summary
- **Total Files**: {len(all_files)}
- **Review Date**: {__import__('datetime').datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
- **Repository**: {os.environ.get('GITHUB_REPOSITORY', 'Unknown')}
- **Commit**: {os.environ.get('GITHUB_SHA', 'Unknown')[:7]}
- **Branch**: {os.environ.get('GITHUB_REF_NAME', 'Unknown')}

## 📁 File Types
{type_summary}

---

## 📝 Review Results

{''.join(all_reviews)}

---

## 💡 Recommendations

Based on the review, here are the top recommendations:
1. Review all critical issues identified above
2. Consider implementing suggested improvements
3. Run tests after making changes
4. Update documentation as needed

> **Note**: This is an automated review. Please use human judgment for final decisions.
"""

with open("/tmp/review-output.txt", "w", encoding="utf-8") as f:
    f.write(report)

print(f"Full review completed. Reviewed {len(all_files)} files in {len(all_files)//batch_size + 1} batches")