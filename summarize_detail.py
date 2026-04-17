import re

with open('/workspace/all_diffs.patch', 'r') as f:
    diff_content = f.read()

files = diff_content.split('diff --git ')
for file_diff in files[1:]:
    lines = file_diff.split('\n')
    header = lines[0]
    file_path = header.split(' b/')[-1]
    
    adds = [l[1:] for l in lines if l.startswith('+') and not l.startswith('+++')]
    dels = [l[1:] for l in lines if l.startswith('-') and not l.startswith('---')]
    
    print(f"File: {file_path}")
    
    if "ci.yml" in file_path or ".golangci.yml" in file_path:
        print("Added CI/Linter workflow.")
        print("="*40)
        continue
    
    if "README.md" in file_path:
        print("Updated README.")
        print("="*40)
        continue

    funcs = []
    types = []
    
    for l in adds:
        m = re.search(r'^func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)', l)
        if m:
            funcs.append("+" + m.group(1))
        m2 = re.search(r'^type\s+([A-Z]\w*)\s+', l)
        if m2:
            types.append("+" + m2.group(1))

    for l in dels:
        m = re.search(r'^func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)', l)
        if m:
            funcs.append("-" + m.group(1))
        m2 = re.search(r'^type\s+([A-Z]\w*)\s+', l)
        if m2:
            types.append("-" + m2.group(1))
            
    if funcs:
        print("Functions: " + ", ".join(funcs))
    if types:
        print("Types: " + ", ".join(types))
        
    if not funcs and not types:
        if len(adds) > 0 or len(dels) > 0:
            print(f"Modifications without new exported funcs/types. +{len(adds)} -{len(dels)}")
            
    print("="*40)
