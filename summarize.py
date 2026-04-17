import re

with open('/workspace/all_diffs.patch', 'r') as f:
    diff_content = f.read()

files = diff_content.split('diff --git ')
for file_diff in files[1:]:
    lines = file_diff.split('\n')
    header = lines[0]
    file_path = header.split(' b/')[-1]
    
    adds = [l for l in lines if l.startswith('+') and not l.startswith('+++')]
    dels = [l for l in lines if l.startswith('-') and not l.startswith('---')]
    
    # Extract added/modified function names or struct names
    funcs = set()
    for l in adds + dels:
        m = re.search(r'func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)', l)
        if m:
            funcs.add(m.group(1))
        m2 = re.search(r'type\s+([A-Z]\w*)\s+struct', l)
        if m2:
            funcs.add(m2.group(1))
            
    print(f"File: {file_path}")
    print(f"Added lines: {len(adds)}, Deleted lines: {len(dels)}")
    if funcs:
        print(f"Functions/Types modified: {', '.join(funcs)}")
    print("-" * 40)
